import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	root: fileURLToPath(new URL('.', import.meta.url)),
	test: {
		include: ['src/**/*.test.ts'],
		environment: 'node',
		coverage: {
			provider: 'v8',
			reporter: ['text', 'json', 'json-summary'],
			reportsDirectory: 'target/coverage',
			include: ['src/lib/config.ts'],
			exclude: ['src/**/*.test.ts'],
			thresholds: {
				lines: 60,
				functions: 60,
				branches: 60,
				statements: 60
			}
		}
	}
});
