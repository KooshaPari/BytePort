"""
BytePort Python SDK — typed client for the BytePort deployment API.

Generated targets: Python 3.10+ (asyncio + pydantic v2).

This SDK is a thin typed wrapper around the BytePort REST API. It targets
the exact endpoint surface described in `docs/openapi.yml` (Wave 1 deliverable)
and uses pydantic v2 for schema validation.

Design goals:
    * Strict typing on every field (no `Any` in hot paths).
    * Sync + async call variants (`deploy()` and `deploy_async()`).
    * Context manager support for both clients.
    * Centralised retry on 429 / 5xx with exponential backoff.
    * Offline-first: cache the agent card once discovered via
      `/.well-known/agent.json`.

Usage:
    >>> import byteport
    >>> client = byteport.Client(base_url="https://api.byteport.dev")
    >>> deployment = client.deploy(
    ...     name="api",
    ...     image="nginx:1.27",
    ...     region="us-west-2",
    ...     instance_count=2,
    ... )
    >>> deployment.id
    UUID('...')

Async usage:
    >>> async with byteport.AsyncClient(base_url=...) as client:
    ...     deployment = await client.deploy(...)
"""

from __future__ import annotations

import asyncio
import logging
import os
import time
import uuid
from contextlib import asynccontextmanager, contextmanager
from dataclasses import dataclass
from enum import Enum
from typing import Any, Iterator

import httpx
from pydantic import BaseModel, Field, HttpUrl

__all__ = [
    "Client",
    "AsyncClient",
    "Deployment",
    "DeploymentStatus",
    "ServiceHealth",
    "AgentCard",
    "BytePortError",
    "RateLimitError",
    "NotFoundError",
]

logger = logging.getLogger("byteport")

DEFAULT_TIMEOUT = 30.0
DEFAULT_RETRIES = 3
DEFAULT_BACKOFF = 0.5
DEFAULT_PAGE_SIZE = 50
MAX_PAGE_SIZE = 100


class BytePortError(Exception):
    """Base exception for all BytePort SDK errors."""

    def __init__(self, message: str, status_code: int | None = None):
        super().__init__(message)
        self.message = message
        self.status_code = status_code


class RateLimitError(BytePortError):
    """Raised when the server returns 429 (rate limit hit)."""


class NotFoundError(BytePortError):
    """Raised when the server returns 404."""


class AuthenticationError(BytePortError):
    """Raised when the server returns 401 / 403."""


class DeploymentStatus(str, Enum):
    """Lifecycle of a deployment."""

    PENDING = "pending"
    PROVISIONING = "provisioning"
    DEPLOYING = "deploying"
    RUNNING = "running"
    FAILED = "failed"
    TERMINATED = "terminated"
    UNKNOWN = "unknown"


class ServiceHealth(str, Enum):
    """Service health states."""

    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"


class Deployment(BaseModel):
    """A BytePort deployment record."""

    id: uuid.UUID
    name: str = Field(min_length=1, max_length=128)
    image: str = Field(min_length=1)
    region: str = Field(min_length=1)
    instance_count: int = Field(ge=0, le=100)
    status: DeploymentStatus
    estimated_cost_usd_monthly: float = Field(ge=0)
    created_at: str
    updated_at: str

    @property
    def is_running(self) -> bool:
        """True iff deployment has reached Running state."""
        return self.status == DeploymentStatus.RUNNING


class AgentCard(BaseModel):
    """The /.well-known/agent.json discovery document."""

    schema_version: str = Field(default="1.0")
    name: str
    version: str
    description: str | None = None
    capabilities: dict[str, Any] = Field(default_factory=dict)
    endpoints: dict[str, Any] = Field(default_factory=dict)
    metadata: dict[str, Any] = Field(default_factory=dict)


@dataclass
class RetryPolicy:
    """Policy for retrying failed HTTP requests.

    The default retries 5xx and 429 with exponential backoff
    (0.5s, 1s, 2s, 4s capped at 8s).
    """

    max_attempts: int = DEFAULT_RETRIES
    initial_backoff: float = DEFAULT_BACKOFF
    backoff_multiplier: float = 2.0
    max_backoff: float = 8.0


class _BaseClient:
    """Shared sync/async HTTP logic."""

    def __init__(
        self,
        base_url: str,
        api_token: str | None = None,
        timeout: float = DEFAULT_TIMEOUT,
        retry_policy: RetryPolicy | None = None,
        agent_card_cache_ttl: float = 3600.0,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_token = api_token or os.environ.get("BYTEPORT_API_TOKEN", "")
        self.timeout = timeout
        self.retry_policy = retry_policy or RetryPolicy()
        self._agent_card: AgentCard | None = None
        self._agent_card_at: float = 0.0
        self._agent_card_cache_ttl = agent_card_cache_ttl

    def _headers(self) -> dict[str, str]:
        """Build default request headers, including bearer if available."""
        h = {"Accept": "application/json", "User-Agent": "byteport-sdk/1.0"}
        if self.api_token:
            h["Authorization"] = f"Bearer {self.api_token}"
        return h

    def _should_retry(self, status: int) -> bool:
        """Retry on 429 or 5xx server-side errors."""
        return status == 429 or (status >= 500 and status < 600)

    def _sleep_for_retry(self, attempt: int) -> float:
        """Exponential backoff bounded by policy."""
        delay = min(
            self.retry_policy.initial_backoff
            * (self.retry_policy.backoff_multiplier ** attempt),
            self.retry_policy.max_backoff,
        )
        return delay


class Client(_BaseClient):
    """Synchronous BytePort API client.

    >>> client = Client(base_url="https://api.byteport.dev")
    >>> for d in client.list_deployments():
    ...     print(d.name)
    """

    def __init__(self, **kwargs: Any) -> None:
        super().__init__(**kwargs)
        self._client = httpx.Client(
            base_url=self.base_url,
            headers=self._headers(),
            timeout=self.timeout,
        )

    def close(self) -> None:
        """Release the underlying HTTP client."""
        self._client.close()

    def __enter__(self) -> "Client":
        return self

    def __exit__(self, *exc: Any) -> None:
        self.close()

    def _request(self, method: str, path: str, **kwargs: Any) -> Any:
        """Issue a request with retry semantics."""
        url = path if path.startswith("/") else f"/{path}"
        last_exc: BytePortError | None = None
        for attempt in range(self.retry_policy.max_attempts):
            try:
                resp = self._client.request(method, url, **kwargs)
            except httpx.HTTPError as exc:
                last_exc = BytePortError(str(exc))
                time.sleep(self._sleep_for_retry(attempt))
                continue
            if self._should_retry(resp.status_code):
                time.sleep(self._sleep_for_retry(attempt))
                continue
            if resp.status_code == 404:
                raise NotFoundError(f"{method} {url} → 404", resp.status_code)
            if resp.status_code in (401, 403):
                raise AuthenticationError(
                    f"{method} {url} → {resp.status_code}", resp.status_code
                )
            if resp.status_code == 429:
                raise RateLimitError(f"{method} {url} → 429", resp.status_code)
            if resp.status_code >= 400:
                raise BytePortError(
                    f"{method} {url} → {resp.status_code}: {resp.text}",
                    resp.status_code,
                )
            return resp.json() if resp.content else None
        raise last_exc or BytePortError("Exceeded retries")

    def health(self) -> dict[str, Any]:
        """GET /healthz — liveness probe.

        Returns server uptime and current epoch.
        """
        return self._request("GET", "/healthz")

    def readiness(self) -> dict[str, Any]:
        """GET /readyz — readiness probe (DB/cache/etc.)."""
        return self._request("GET", "/readyz")

    def agent_card(self, *, force_refresh: bool = False) -> AgentCard:
        """Fetch /.well-known/agent.json (cached for `agent_card_cache_ttl` s)."""
        now = time.time()
        if (
            not force_refresh
            and self._agent_card is not None
            and (now - self._agent_card_at) < self._agent_card_cache_ttl
        ):
            return self._agent_card
        data = self._request("GET", "/.well-known/agent.json")
        self._agent_card = AgentCard.model_validate(data)
        self._agent_card_at = now
        return self._agent_card

    def deploy(
        self,
        *,
        name: str,
        image: str,
        region: str,
        instance_count: int = 1,
        **extra: Any,
    ) -> Deployment:
        """POST /api/v1/deployments — create a deployment."""
        body = {
            "name": name,
            "image": image,
            "region": region,
            "instance_count": instance_count,
            **extra,
        }
        data = self._request("POST", "/api/v1/deployments", json=body)
        return Deployment.model_validate(data)

    def list_deployments(
        self, *, page: int = 1, page_size: int = DEFAULT_PAGE_SIZE
    ) -> list[Deployment]:
        """GET /api/v1/deployments — paginated list."""
        page_size = min(page_size, MAX_PAGE_SIZE)
        data = self._request(
            "GET",
            "/api/v1/deployments",
            params={"page": page, "page_size": page_size},
        )
        items = data.get("items", data) if isinstance(data, dict) else data
        return [Deployment.model_validate(item) for item in items]

    def get_deployment(self, deployment_id: uuid.UUID | str) -> Deployment:
        """GET /api/v1/deployments/{id} — fetch a deployment."""
        data = self._request(
            "GET", f"/api/v1/deployments/{deployment_id}"
        )
        return Deployment.model_validate(data)

    def terminate_deployment(self, deployment_id: uuid.UUID | str) -> None:
        """DELETE /api/v1/deployments/{id} — terminate."""
        self._request("DELETE", f"/api/v1/deployments/{deployment_id}")


class AsyncClient(_BaseClient):
    """Asynchronous (httpx.AsyncClient) variant of `Client`.

    >>> import asyncio, byteport
    >>> async def main():
    ...     async with byteport.AsyncClient(base_url=...) as c:
    ...         return await c.deploy(name="api", image="nginx", region="us-west-2")
    >>> asyncio.run(main())
    """

    def __init__(self, **kwargs: Any) -> None:
        super().__init__(**kwargs)
        self._client = httpx.AsyncClient(
            base_url=self.base_url,
            headers=self._headers(),
            timeout=self.timeout,
        )

    async def __aenter__(self) -> "AsyncClient":
        return self

    async def __aexit__(self, *exc: Any) -> None:
        await self.aclose()

    async def aclose(self) -> None:
        """Release the underlying async HTTP client."""
        await self._client.aclose()

    async def _request(self, method: str, path: str, **kwargs: Any) -> Any:
        url = path if path.startswith("/") else f"/{path}"
        last_exc: BytePortError | None = None
        for attempt in range(self.retry_policy.max_attempts):
            try:
                resp = await self._client.request(method, url, **kwargs)
            except httpx.HTTPError as exc:
                last_exc = BytePortError(str(exc))
                await asyncio.sleep(self._sleep_for_retry(attempt))
                continue
            if self._should_retry(resp.status_code):
                await asyncio.sleep(self._sleep_for_retry(attempt))
                continue
            if resp.status_code == 404:
                raise NotFoundError(
                    f"{method} {url} → 404", resp.status_code
                )
            if resp.status_code in (401, 403):
                raise AuthenticationError(
                    f"{method} {url} → {resp.status_code}", resp.status_code
                )
            if resp.status_code == 429:
                raise RateLimitError(
                    f"{method} {url} → 429", resp.status_code
                )
            if resp.status_code >= 400:
                raise BytePortError(
                    f"{method} {url} → {resp.status_code}: {resp.text}",
                    resp.status_code,
                )
            return resp.json() if resp.content else None
        raise last_exc or BytePortError("Exceeded retries")

    async def health(self) -> dict[str, Any]:
        """GET /healthz."""
        return await self._request("GET", "/healthz")

    async def readiness(self) -> dict[str, Any]:
        """GET /readyz."""
        return await self._request("GET", "/readyz")

    async def agent_card(
        self, *, force_refresh: bool = False
    ) -> AgentCard:
        """Fetch /.well-known/agent.json with TTL cache."""
        now = time.time()
        if (
            not force_refresh
            and self._agent_card is not None
            and (now - self._agent_card_at) < self._agent_card_cache_ttl
        ):
            return self._agent_card
        data = await self._request("GET", "/.well-known/agent.json")
        self._agent_card = AgentCard.model_validate(data)
        self._agent_card_at = now
        return self._agent_card

    async def deploy(
        self,
        *,
        name: str,
        image: str,
        region: str,
        instance_count: int = 1,
        **extra: Any,
    ) -> Deployment:
        """POST /api/v1/deployments."""
        body = {
            "name": name,
            "image": image,
            "region": region,
            "instance_count": instance_count,
            **extra,
        }
        data = await self._request(
            "POST", "/api/v1/deployments", json=body
        )
        return Deployment.model_validate(data)

    async def list_deployments(
        self, *, page: int = 1, page_size: int = DEFAULT_PAGE_SIZE
    ) -> list[Deployment]:
        """GET /api/v1/deployments."""
        page_size = min(page_size, MAX_PAGE_SIZE)
        data = await self._request(
            "GET",
            "/api/v1/deployments",
            params={"page": page, "page_size": page_size},
        )
        items = data.get("items", data) if isinstance(data, dict) else data
        return [Deployment.model_validate(item) for item in items]

    async def get_deployment(
        self, deployment_id: uuid.UUID | str
    ) -> Deployment:
        """GET /api/v1/deployments/{id}."""
        data = await self._request(
            "GET", f"/api/v1/deployments/{deployment_id}"
        )
        return Deployment.model_validate(data)

    async def terminate_deployment(
        self, deployment_id: uuid.UUID | str
    ) -> None:
        """DELETE /api/v1/deployments/{id}."""
        await self._request(
            "DELETE", f"/api/v1/deployments/{deployment_id}"
        )


@contextmanager
def _client(
    base_url: str,
    **kwargs: Any,
) -> Iterator[Client]:
    """Convenience context manager for sync clients."""
    c = Client(base_url=base_url, **kwargs)
    try:
        yield c
    finally:
        c.close()


@asynccontextmanager
async def _async_client(
    base_url: str,
    **kwargs: Any,
) -> AsyncIterator[AsyncClient]:
    """Convenience context manager for async clients."""
    c = AsyncClient(base_url=base_url, **kwargs)
    try:
        yield c
    finally:
        await c.aclose()
