/// <reference types="vitest" />
import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/api': 'http://localhost:8081',
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
