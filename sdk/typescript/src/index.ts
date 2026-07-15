// BytePort TypeScript SDK — zod-validated client for the BytePort API
//
// Exports:
//   Client        — HTTP client wrapping fetch()
//   schemas       — zod schemas for all wire types
//   errors        — typed error classes
//   utils         — helpers (pagination, polling, retry)

import { z } from "zod";

// ────────────────────────────────────────────
//  Wire Schemas
// ────────────────────────────────────────────

export const DeploymentSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1).max(255),
  status: z.enum(["pending", "deploying", "active", "failed", "terminated"]),
  namespace: z.string().default("default"),
  replicas: z.number().int().min(0).default(1),
  endpoints: z.array(z.string().url()).optional(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime().optional(),
});
export type Deployment = z.infer<typeof DeploymentSchema>;

export const DeploymentListSchema = z.object({
  items: z.array(DeploymentSchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  page_size: z.number().int().positive().max(100),
});
export type DeploymentList = z.infer<typeof DeploymentListSchema>;

export const CreateDeploymentSchema = z.object({
  name: z.string().min(1).max(255),
  namespace: z.string().default("default"),
  replicas: z.number().int().min(0).max(100).default(1),
  image: z.string().min(1).optional(),
  env: z.record(z.string()).optional(),
});
export type CreateDeployment = z.infer<typeof CreateDeploymentSchema>;

export const ComputeResourceSchema = z.object({
  id: z.string().uuid(),
  provider: z.enum(["aws", "gcp", "azure", "local"]),
  region: z.string().min(1),
  instance_type: z.string().min(1),
  status: z.enum(["provisioning", "ready", "busy", "draining", "offline"]),
  labels: z.record(z.string()).default({}),
  created_at: z.string().datetime(),
});
export type ComputeResource = z.infer<typeof ComputeResourceSchema>;

export const AgentCardSchema = z.object({
  name: z.string(),
  version: z.string(),
  description: z.string(),
  tools: z.array(z.string()),
  capabilities: z.object({
    streaming: z.boolean().optional(),
    batch: z.boolean().optional(),
    a2a: z.boolean().optional(),
  }).optional(),
});
export type AgentCard = z.infer<typeof AgentCardSchema>;

export const TelemetryEventSchema = z.object({
  name: z.string().min(1),
  attributes: z.record(z.union([z.string(), z.number(), z.boolean()])).optional(),
  timestamp: z.string().datetime().optional(),
});
export type TelemetryEvent = z.infer<typeof TelemetryEventSchema>;

export const ErrorResponseSchema = z.object({
  error: z.string(),
  error_code: z.string().optional(),
  details: z.unknown().optional(),
});
export type ErrorResponse = z.infer<typeof ErrorResponseSchema>;

// ────────────────────────────────────────────
//  Errors
// ────────────────────────────────────────────

export class BytePortError extends Error {
  constructor(
    message: string,
    public status?: number,
    public code?: string,
  ) {
    super(message);
    this.name = "BytePortError";
  }
}

export class ValidationError extends BytePortError {
  constructor(message: string, public issues?: z.ZodIssue[]) {
    super(message, 400, "VALIDATION_ERROR");
    this.name = "ValidationError";
  }
}

export class NotFoundError extends BytePortError {
  constructor(resource: string, id: string) {
    super(`${resource} '${id}' not found`, 404, "NOT_FOUND");
    this.name = "NotFoundError";
  }
}

export class RateLimitError extends BytePortError {
  constructor(public retryAfter: number) {
    super("Rate limit exceeded", 429, "RATE_LIMITED");
    this.name = "RateLimitError";
  }
}

// ────────────────────────────────────────────
//  Client
// ────────────────────────────────────────────

export interface ClientOptions {
  baseUrl?: string;
  apiKey?: string;
  timeout?: number;
  retries?: number;
}

export class Client {
  private baseUrl: string;
  private apiKey?: string;
  private timeout: number;
  private retries: number;

  constructor(opts: ClientOptions = {}) {
    this.baseUrl = opts.baseUrl ?? "http://localhost:8080/api/v1";
    this.apiKey = opts.apiKey;
    this.timeout = opts.timeout ?? 30_000;
    this.retries = opts.retries ?? 3;
  }

  // ── Deployments ──

  async listDeployments(opts?: {
    namespace?: string;
    page?: number;
    pageSize?: number;
  }): Promise<DeploymentList> {
    const params = new URLSearchParams();
    if (opts?.namespace) params.set("namespace", opts.namespace);
    if (opts?.page) params.set("page", String(opts.page));
    if (opts?.pageSize) params.set("page_size", String(opts.pageSize));
    const data = await this.request<unknown>(
      "GET",
      `/deployments?${params}`,
    );
    return DeploymentListSchema.parse(data);
  }

  async getDeployment(id: string): Promise<Deployment> {
    const data = await this.request<unknown>("GET", `/deployments/${id}`);
    return DeploymentSchema.parse(data);
  }

  async createDeployment(input: CreateDeployment): Promise<Deployment> {
    const parsed = CreateDeploymentSchema.parse(input);
    const data = await this.request<unknown>("POST", "/deployments", parsed);
    return DeploymentSchema.parse(data);
  }

  async deleteDeployment(id: string): Promise<void> {
    await this.request("DELETE", `/deployments/${id}`);
  }

  // ── Compute ──

  async listCompute(opts?: { status?: string }): Promise<ComputeResource[]> {
    const params = new URLSearchParams();
    if (opts?.status) params.set("status", opts.status);
    const data = await this.request<unknown[]>(
      "GET",
      `/compute?${params}`,
    );
    return z.array(ComputeResourceSchema).parse(data);
  }

  // ── Agent Card ──

  async getAgentCard(): Promise<AgentCard> {
    const data = await this.request<unknown>("GET", "/.well-known/agent.json");
    return AgentCardSchema.parse(data);
  }

  // ── Telemetry ──

  async sendTelemetry(events: TelemetryEvent[]): Promise<void> {
    const parsed = z.array(TelemetryEventSchema).parse(events);
    await this.request("POST", "/telemetry", parsed);
  }

  // ── Internal ──

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    let lastError: Error | null = null;
    for (let attempt = 0; attempt <= this.retries; attempt++) {
      try {
        const controller = new AbortController();
        const timer = setTimeout(() => controller.abort(), this.timeout);

        const headers: Record<string, string> = {
          "Content-Type": "application/json",
          Accept: "application/json",
        };
        if (this.apiKey) {
          headers["Authorization"] = `Bearer ${this.apiKey}`;
        }

        const resp = await fetch(`${this.baseUrl}${path}`, {
          method,
          headers,
          body: body ? JSON.stringify(body) : undefined,
          signal: controller.signal,
        });
        clearTimeout(timer);

        if (resp.status === 429) {
          const retryAfter = parseInt(
            resp.headers.get("Retry-After") ?? "1",
            10,
          );
          throw new RateLimitError(retryAfter);
        }
        if (resp.status === 404) {
          throw new NotFoundError("resource", path);
        }
        if (!resp.ok) {
          const text = await resp.text().catch(() => "unknown error");
          throw new BytePortError(text, resp.status);
        }
        if (resp.status === 204) return undefined as T;
        return (await resp.json()) as T;
      } catch (err) {
        lastError = err instanceof Error ? err : new Error(String(err));
        if (attempt < this.retries) {
          const delay = Math.min(1000 * 2 ** attempt, 10_000);
          await new Promise((r) => setTimeout(r, delay));
        }
      }
    }
    throw lastError ?? new BytePortError("request failed");
  }
}

// ────────────────────────────────────────────
//  Utils
// ────────────────────────────────────────────

export async function* paginate<T>(
  client: Client,
  fetchFn: (page: number) => Promise<{ items: T[]; total: number; page: number }>,
): AsyncGenerator<T, void, undefined> {
  let page = 1;
  let fetched = 0;
  let total = Infinity;
  while (fetched < total) {
    const result = await fetchFn(page);
    for (const item of result.items) {
      yield item;
      fetched++;
    }
    total = result.total;
    page++;
  }
}
