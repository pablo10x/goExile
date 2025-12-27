<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { config, restartRequired } from '$lib/stores';
	import type { ServerConfig } from '$lib/stores';

	let loading = $state(true);
	let error = $state<string | null>(null);
	let categoryConfigs = $state<ServerConfig[]>([]);
	let saving = $state(false);

	let category = $derived($page.params.category || '');
	const categoryTitles: Record<string, string> = {
		system: 'System Configuration',
		spawner: 'Spawner Configuration',
		security: 'Security Configuration'
	};

	async function loadCategoryConfig() {
		try {
			loading = true;
			error = null;

			const response = await fetch(`/api/config/category/${category}`);
			if (!response.ok) {
				throw new Error(`Failed to load ${category} configuration: ${response.statusText}`);
			}

			const configData = await response.json();
			categoryConfigs = configData;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load configuration';
			console.error('Failed to load configuration:', e);
		} finally {
			loading = false;
		}
	}

	async function updateConfig(key: string, value: string) {
		try {
			saving = true;

			const response = await fetch(`/api/config/${key}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ value })
			});

			if (!response.ok) {
				throw new Error(`Failed to update ${key}: ${response.statusText}`);
			}

			// Update local state
			const configIndex = categoryConfigs.findIndex((c) => c.key === key);
			if (configIndex !== -1) {
				const updatedConfig = { ...categoryConfigs[configIndex], value };

				// Check if restart is required
				if (updatedConfig.requires_restart) {
					restartRequired.set(true);
				}

				categoryConfigs = [
					...categoryConfigs.slice(0, configIndex),
					updatedConfig,
					...categoryConfigs.slice(configIndex + 1)
				];
			}

			// Also update global config store
			if ($config) {
				const globalIndex = $config.findIndex((c) => c.key === key);
				if (globalIndex !== -1) {
					const updatedGlobal = { ...$config[globalIndex], value };
					config.set([
						...$config.slice(0, globalIndex),
						updatedGlobal,
						...$config.slice(globalIndex + 1)
					]);
				}
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to update configuration';
			console.error('Failed to update configuration:', e);
		} finally {
			saving = false;
		}
	}

	onMount(() => {
		loadCategoryConfig();

		// Make updateConfig available globally for inline event handlers
		(window as any).updateConfig = updateConfig;
	});
</script>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold text-slate-100 mb-2">
				{categoryTitles[category] || 'Configuration'}
			</h1>
			<p class="text-slate-500 dark:text-slate-400">Manage {category} settings and preferences</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/config"
				class="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-slate-900 dark:text-white rounded-lg transition-colors duration-200 flex items-center gap-2"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="w-4 h-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path d="M19 12H5M12 19l-7-7 7-7" />
				</svg>
				Back
			</a>
			<button
				onclick={loadCategoryConfig}
				disabled={saving}
				class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-slate-900 dark:text-white rounded-lg transition-colors duration-200 flex items-center gap-2 disabled:opacity-50"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="w-4 h-4"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path d="M23 4v6h-6"></path>
					<path d="M1 20v-6h6"></path>
					<path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
				</svg>
				Refresh
			</button>
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="bg-red-500/10 border border-red-500/30 rounded-lg p-4 text-red-300">
			<div class="flex items-center gap-3">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="w-5 h-5"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="8" x2="12" y2="12"></line>
					<line x1="12" y1="16" x2="12.01" y2="16"></line>
				</svg>
				<span>{error}</span>
			</div>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
		</div>
	{:else if categoryConfigs.length === 0}
		<div class="text-center py-12">
			<div class="text-slate-500 dark:text-slate-400 mb-2">
				No configuration settings found for {category}
			</div>
			<div class="text-slate-500 text-sm">
				This category may not have any configurable options yet.
			</div>
		</div>
	{:else}
		<!-- Configuration Settings -->
		<div class="space-y-4">
			{#each categoryConfigs as configItem (configItem.key)}
				<div
					class="bg-slate-800/50 backdrop-blur-sm border border-slate-300 dark:border-slate-700 rounded-xl p-6"
				>
					<div class="flex items-start justify-between mb-4">
						<div class="flex-1">
							<div class="flex items-center gap-3 mb-2">
								<h3 class="text-lg font-semibold text-slate-100">{configItem.key}</h3>
								{#if configItem.is_read_only}
									<span
										class="px-2 py-1 bg-slate-700 text-slate-700 dark:text-slate-300 text-xs rounded"
										>Read-only</span
									>
								{/if}
								{#if configItem.requires_restart}
									<span class="px-2 py-1 bg-orange-500/20 text-orange-300 text-xs rounded"
										>Restart Required</span
									>
								{/if}
							</div>
							<p class="text-slate-500 dark:text-slate-400 text-sm mb-3">
								{configItem.description}
							</p>
							<div class="flex items-center gap-4">
								<div class="flex-1 max-w-md">
									{#if configItem.type === 'bool'}
										<label class="flex items-center cursor-pointer">
											<input
												type="checkbox"
												class="sr-only peer"
												checked={configItem.value === 'true'}
												onchange={(e) =>
													updateConfig(configItem.key, e.currentTarget.checked ? 'true' : 'false')}
												disabled={configItem.is_read_only}
											/>
											<div
												class="relative w-11 h-6 bg-gray-600 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"
											></div>
										</label>
									{:else if configItem.type === 'int'}
										<input
											type="number"
											value={configItem.value}
											onchange={(e) => updateConfig(configItem.key, e.currentTarget.value)}
											disabled={configItem.is_read_only}
											class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
										/>
									{:else}
										<input
											type="text"
											value={configItem.value}
											onchange={(e) => updateConfig(configItem.key, e.currentTarget.value)}
											disabled={configItem.is_read_only}
											class="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded-lg text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
										/>
									{/if}
								</div>
								<div class="text-xs text-slate-500">
									Type: {configItem.type}<br />
									Updated: {new Date(configItem.updated_at).toLocaleString()}<br />
									By: {configItem.updated_by}
								</div>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
