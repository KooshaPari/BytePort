import { afterEach, describe, expect, it, vi } from 'vitest';

import config, { apiHelpers } from './config';

describe('BytePort frontend configuration helpers', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('normalizes API and NanoVMS endpoint paths', () => {
		expect(apiHelpers.getApiUrl('/projects')).toBe('http://localhost:8081/projects');
		expect(apiHelpers.getApiUrl('projects')).toBe('http://localhost:8081/projects');
		expect(apiHelpers.getNvmsUrl('/health')).toBe('http://localhost:3000/health');
		expect(apiHelpers.getNvmsUrl('health')).toBe('http://localhost:3000/health');
	});

	it('returns credentialed JSON request defaults', () => {
		expect(apiHelpers.getDefaultOptions()).toEqual({
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include'
		});
	});

	it('merges request headers and returns successful responses', async () => {
		const response = new Response(JSON.stringify({ ok: true }), { status: 200 });
		const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(response);

		await expect(
			apiHelpers.makeRequest('/projects', {
				method: 'POST',
				headers: { 'X-Request-ID': 'unit-test' }
			})
		).resolves.toBe(response);
		expect(fetchMock).toHaveBeenCalledWith('/projects', {
			headers: {
				'Content-Type': 'application/json',
				'X-Request-ID': 'unit-test'
			},
			credentials: 'include',
			method: 'POST'
		});
	});

	it('surfaces API error messages', async () => {
		vi.spyOn(globalThis, 'fetch').mockResolvedValue(
			new Response(JSON.stringify({ message: 'service unavailable' }), {
				status: 503,
				statusText: 'Service Unavailable'
			})
		);

		await expect(apiHelpers.makeRequest('/health')).rejects.toThrow('service unavailable');
	});

	it('exposes the foundation-safe default endpoints', () => {
		expect(config.api.timeout).toBe(30000);
		expect(config.api.baseUrl).toBe('http://localhost:8081');
		expect(config.api.nvmsUrl).toBe('http://localhost:3000');
		expect(config.deployment.containerRuntime).toBe('podman');
		expect(config.features.podmanDeployment).toBe(true);
		expect(config.features.dockerDeployment).toBe(false);
		expect(config.deployment.supportedTypes).toEqual(['nodejs', 'go', 'python', 'rust', 'static']);
	});
});
