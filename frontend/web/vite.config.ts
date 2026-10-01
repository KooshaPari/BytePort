import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	// `tailwindcss()` must precede `sveltekit()`: it resolves the bare
	// `@import 'tailwindcss'` specifier in src/app.css as a package. Without it,
	// Vite's CSS pipeline joins bare specifiers to the importing file's own
	// directory and fails with `ENOENT .../frontend/web/tailwindcss`.
	plugins: [tailwindcss(), sveltekit()],
	optimizeDeps: {
		include: ['lucide-svelte']
	},
	test: {
		environment: 'node',
		// Playwright e2e specs are owned by @playwright/test, not vitest.
		exclude: ['tests/e2e/**', '**/node_modules/**', '**/dist/**', '**/.svelte-kit/**']
	},
	server: {
		host: '0.0.0.0',
		port: 5173,
		strictPort: true,
		hmr: {
			host: 'localhost',
			port: 5173,
			protocol: 'ws'
		},
		headers: {
			'Access-Control-Allow-Origin': '*',
			'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
			'Access-Control-Allow-Headers': 'Content-Type',
			'Access-Control-Allow-Credentials': 'true'
		}
	},
	ssr: {
		noExternal: ['lucide-svelte'] // Prevents treating it as an external dependency
	}
});
