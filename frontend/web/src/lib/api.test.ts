import { describe, expect, it } from 'vitest';
import { API_PORT, apiUrl, getApiBaseUrl } from './api';

describe('getApiBaseUrl', () => {
	it('falls back to localhost outside Tauri', () => {
		// Vitest runs in node: no window, so no Tauri internals to detect.
		expect(getApiBaseUrl()).toBe(`http://localhost:${API_PORT}`);
	});
});

describe('apiUrl', () => {
	it('joins a leading-slash path without doubling the separator', () => {
		expect(apiUrl('/user/1')).toBe(`http://localhost:${API_PORT}/user/1`);
	});

	it('normalizes a path missing the leading slash', () => {
		expect(apiUrl('user/1')).toBe(`http://localhost:${API_PORT}/user/1`);
	});
});
