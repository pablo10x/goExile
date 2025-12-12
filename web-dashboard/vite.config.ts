/// <reference types="vitest" />
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 3000,
		proxy: {
			'/api': 'http://localhost:8081',
			'/events': 'http://localhost:8081',
			'/login': 'http://localhost:8081',
			'/logout': 'http://localhost:8081',
			'/health': 'http://localhost:8081'
		}
	},
	test: {
		include: ['src/**/*.{test,spec}.{js,ts}'],
		environment: 'jsdom',
		globals: true,
		setupFiles: ['src/setupTests.ts']
	}
});
