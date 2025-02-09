import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	optimizeDeps: {
		include: ['lucide-svelte']
	},
	server: {
		host: '0.0.0.0',
		port: 5173,
		strictPort: true
	},
	ssr: {
		noExternal: ['lucide-svelte'] // Prevents treating it as an external dependency
	}
});
