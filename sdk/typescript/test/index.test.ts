// BytePort TypeScript SDK — test suite
import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  Client,
  DeploymentSchema,
  DeploymentListSchema,
  CreateDeploymentSchema,
  BytePortError,
  RateLimitError,
  paginate,
} from "../src/index";

// ── Schema Tests ──

describe("DeploymentSchema", () => {
  it("parses a valid deployment", () => {
    const result = DeploymentSchema.parse({
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "web-app",
      status: "active",
      created_at: "2026-07-01T00:00:00Z",
    });
    expect(result.name).toBe("web-app");
    expect(result.status).toBe("active");
    expect(result.replicas).toBe(1); // default
  });

  it("rejects invalid status", () => {
    expect(() =>
      DeploymentSchema.parse({
        id: "550e8400-e29b-41d4-a716-446655440000",
        name: "bad",
        status: "unknown_status",
        created_at: "2026-07-01T00:00:00Z",
      }),
    ).toThrow();
  });

  it("rejects missing required fields", () => {
    expect(() => DeploymentSchema.parse({})).toThrow();
  });
});

describe("DeploymentListSchema", () => {
  it("parses a valid list response", () => {
    const result = DeploymentListSchema.parse({
      items: [
        {
          id: "550e8400-e29b-41d4-a716-446655440000",
          name: "web",
          status: "active",
          created_at: "2026-07-01T00:00:00Z",
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    });
    expect(result.items).toHaveLength(1);
    expect(result.total).toBe(1);
  });
});

describe("CreateDeploymentSchema", () => {
  it("parses a valid create request", () => {
    const result = CreateDeploymentSchema.parse({
      name: "my-app",
      replicas: 3,
    });
    expect(result.name).toBe("my-app");
    expect(result.replicas).toBe(3);
    expect(result.namespace).toBe("default");
  });
});

// ── Client Tests ──

describe("Client", () => {
  let client: Client;
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    vi.stubGlobal("fetch", mockFetch);
    client = new Client({ baseUrl: "http://test.local/api/v1" });
  });

  it("listDeployments returns parsed response", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Map(),
      json: () =>
        Promise.resolve({
          items: [
            {
              id: "550e8400-e29b-41d4-a716-446655440000",
              name: "web",
              status: "active",
              created_at: "2026-07-01T00:00:00Z",
            },
          ],
          total: 1,
          page: 1,
          page_size: 20,
        }),
    });

    const result = await client.listDeployments();
    expect(result.items).toHaveLength(1);
    expect(result.total).toBe(1);
    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining("/deployments"),
      expect.any(Object),
    );
  });

  it("throws RateLimitError on 429", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 429,
      headers: new Map([["retry-after", "5"]]),
      text: () => Promise.resolve("rate limited"),
    });

    await expect(client.listDeployments()).rejects.toThrow(RateLimitError);
  });

  it("throws BytePortError on 500", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      headers: new Map(),
      text: () => Promise.resolve("internal error"),
    });

    await expect(client.listDeployments()).rejects.toThrow(BytePortError);
  });

  it("retries on network failure", async () => {
    mockFetch
      .mockRejectedValueOnce(new Error("network lost"))
      .mockRejectedValueOnce(new Error("network lost"))
      .mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Map(),
        json: () =>
          Promise.resolve({
            items: [],
            total: 0,
            page: 1,
            page_size: 20,
          }),
      });

    const result = await client.listDeployments();
    expect(result.items).toEqual([]);
    expect(mockFetch).toHaveBeenCalledTimes(3);
  });
});

// ── Util Tests ──

describe("paginate", () => {
  it("yields all items across pages", async () => {
    const items = await Array.fromAsync(
      paginate(
        new Client({ baseUrl: "http://test.local" }),
        async (page) => ({
          items:
            page === 1
              ? [{ id: "a", name: "x", status: "active", created_at: "z" }]
              : [{ id: "b", name: "y", status: "active", created_at: "z" }],
          total: 2,
          page,
        }),
      ),
    );
    expect(items).toHaveLength(2);
    expect(items[0].id).toBe("a");
    expect(items[1].id).toBe("b");
  });
});
