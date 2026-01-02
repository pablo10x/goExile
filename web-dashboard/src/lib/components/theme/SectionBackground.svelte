<script lang="ts">
	import { fade } from 'svelte/transition';

	let { 
		type = 'digital_stream', 
		mixBlend = 'normal'
	} = $props<{
		type?: string;
		mixBlend?: string;
	}>();

	// Hardcoded Industrial Parameters
	const activeColor = '#f97316';
	const activeOpacity = 0.2;
</script>

<div class="absolute inset-0 pointer-events-none overflow-hidden select-none" style="mix-blend-mode: {mixBlend}; opacity: {activeOpacity};">
	{#if type === 'digital_stream'}
		<div class="absolute inset-0" in:fade>
			<div class="digital-lines will-change-transform" style="--line-color: {activeColor}; --anim-speed: 10s"></div>
		</div>
	{:else if type === 'circuit_grid'}
		<div class="absolute inset-0" in:fade>
			<div class="circuit-pattern will-change-transform" style="--grid-color: {activeColor}; --anim-speed: 20s"></div>
		</div>
	{:else if type === 'none'}
		<!-- No background -->
	{:else}
		<!-- Fallback to digital stream if unknown -->
		<div class="absolute inset-0" in:fade>
			<div class="digital-lines will-change-transform" style="--line-color: {activeColor}; --anim-speed: 10s"></div>
		</div>
	{/if}
</div>

<style>
	/* Digital Stream */
	.digital-lines {
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, var(--line-color) 1px, transparent 1px);
		background-size: 40px 100%;
		mask-image: linear-gradient(to bottom, transparent, black, transparent);
		animation: streamMove var(--anim-speed) linear infinite;
	}
	@keyframes streamMove {
		from { background-position: 0 0; }
		to { background-position: 400px 0; }
	}

	/* Circuit Grid */
	.circuit-pattern {
		width: 100%;
		height: 100%;
		background-image: 
			radial-gradient(var(--grid-color) 1px, transparent 1px),
			linear-gradient(var(--grid-color) 0.5px, transparent 0.5px),
			linear-gradient(90deg, var(--grid-color) 0.5px, transparent 0.5px);
		background-size: 60px 60px, 60px 60px, 60px 60px;
		background-position: 0 0;
		animation: circuitMove var(--anim-speed) linear infinite;
	}
	@keyframes circuitMove {
		from { background-position: 0 0; }
		to { background-position: 60px 60px; }
	}
</style>