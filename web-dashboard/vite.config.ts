/// <reference types="vitest" />
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit(), tailwindcss()],
	server: {
		port: 3001,
		proxy: {
			'/api': 'http://127.0.0.1:8081',
			'/events': 'http://127.0.0.1:8081',
			'/health': 'http://127.0.0.1:8081'
		}
	},
	test: {
		include: ['src/**/*.{test,spec}.{js,ts}'],
		environment: 'jsdom',
		globals: true,
		setupFiles: ['src/setupTests.ts']
	}
});
