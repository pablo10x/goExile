<script lang="ts">
	import { backgroundConfig, siteSettings } from '$lib/stores';
	import { fade } from 'svelte/transition';

	let { 
		type = 'digital_stream', 
		mixBlend = 'normal'
	} = $props<{
		type?: string;
		mixBlend?: string;
	}>();

	let currentSettings = $derived(($backgroundConfig.settings as any)[type] || { speed: 1, color: '#f97316', opacity: 0.5, scale: 1 });
	let activeSpeed = $derived(currentSettings.speed ?? 1);
	let activeColor = $derived(currentSettings.color ?? '#f97316');
	let activeOpacity = $derived(currentSettings.opacity ?? 0.5);
	let activeScale = $derived(currentSettings.scale ?? 1);
	
	// Link to global aesthetic overrides if present
	let globalSpeed = $derived($siteSettings.aesthetic.bg_anim_speed ?? 1);
	let globalOpacity = $derived($siteSettings.aesthetic.bg_anim_opacity ?? 0.5);
</script>

<div class="absolute inset-0 pointer-events-none overflow-hidden select-none" style="mix-blend-mode: {mixBlend}; opacity: {globalOpacity * activeOpacity}; transform: scale({activeScale})">
	{#if !$siteSettings.performance?.low_power_mode}
		{#if type === 'digital_stream'}
			<div class="absolute inset-0" in:fade>
				<div class="digital-lines will-change-transform" style="--line-color: {activeColor}; --anim-speed: {10 / (activeSpeed * globalSpeed)}s"></div>
			</div>
		{:else if type === 'circuit_grid'}
			<div class="absolute inset-0" in:fade>
				<div class="circuit-pattern will-change-transform" style="--grid-color: {activeColor}; --anim-speed: {20 / (activeSpeed * globalSpeed)}s"></div>
			</div>
		{:else if type === 'neon_pulse'}
			<div class="absolute inset-0" in:fade>
				<div class="neon-blobs will-change-transform" style="--neon-color: {activeColor}; --anim-speed: {15 / (activeSpeed * globalSpeed)}s"></div>
			</div>
		{:else if type === 'noise_static'}
			<div class="absolute inset-0" in:fade>
				<div class="noise-texture"></div>
			</div>
		{:else if type === 'glass_refraction'}
			<div class="absolute inset-0" in:fade>
				<div class="glass-shards will-change-transform" style="--glass-color: {activeColor}; --anim-speed: {25 / (activeSpeed * globalSpeed)}s"></div>
			</div>
		{:else if type === 'cyber_scan'}
			<div class="absolute inset-0" in:fade>
				<div class="scan-horizontal will-change-[top]" style="--scan-color: {activeColor}; --anim-speed: {4 / (activeSpeed * globalSpeed)}s"></div>
				<div class="scan-vertical will-change-[left]" style="--scan-color: {activeColor}; --anim-speed: {7 / (activeSpeed * globalSpeed)}s"></div>
			</div>
		{:else if type === 'vector_wave'}
			<div class="absolute inset-0" in:fade>
				<svg class="w-full h-full">
					<defs>
						<linearGradient id="waveGradient" x1="0%" y1="0%" x2="100%" y2="0%">
							<stop offset="0%" stop-color="transparent" />
							<stop offset="50%" stop-color={activeColor} />
							<stop offset="100%" stop-color="transparent" />
						</linearGradient>
					</defs>
					<path class="wave-path will-change-transform" d="M0 50 Q 250 100 500 50 T 1000 50 V 100 H 0 Z" style="fill: url(#waveGradient); --anim-speed: {12 / (activeSpeed * globalSpeed)}s" />
					<path class="wave-path-2 will-change-transform" d="M0 50 Q 250 0 500 50 T 1000 50 V 100 H 0 Z" style="fill: url(#waveGradient); opacity: 0.5; --anim-speed: {18 / (activeSpeed * globalSpeed)}s" />
				</svg>
			</div>
		{/if}
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

	/* Neon Pulse */
	.neon-blobs {
		width: 100%;
		height: 100%;
		background: radial-gradient(circle at 20% 30%, var(--neon-color) 0%, transparent 40%),
					radial-gradient(circle at 80% 70%, var(--neon-color) 0%, transparent 40%);
		filter: blur(80px);
		animation: neonMove var(--anim-speed) ease-in-out infinite alternate;
	}
	@keyframes neonMove {
		0% { transform: scale(1) translate(0, 0); opacity: 0.5; }
		100% { transform: scale(1.2) translate(5%, 5%); opacity: 1; }
	}

	/* Noise Static */
	.noise-texture {
		width: 100%;
		height: 100%;
		background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.65' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E");
	}

	/* Glass Refraction */
	.glass-shards {
		width: 100%;
		height: 100%;
		background: repeating-linear-gradient(45deg, var(--glass-color) 0%, transparent 5%, var(--glass-color) 10%);
		background-size: 200% 200%;
		animation: glassMove var(--anim-speed) linear infinite;
	}
	@keyframes glassMove {
		from { background-position: 0% 0%; }
		to { background-position: 100% 100%; }
	}

	/* Cyber Scan */
	.scan-horizontal, .scan-vertical {
		position: absolute;
		background: linear-gradient(to bottom, transparent, var(--scan-color), transparent);
	}
	.scan-horizontal {
		width: 100%;
		height: 2px;
		top: 0;
		animation: scanH var(--anim-speed) linear infinite;
	}
	.scan-vertical {
		width: 2px;
		height: 100%;
		left: 0;
		background: linear-gradient(to right, transparent, var(--scan-color), transparent);
		animation: scanV var(--anim-speed) linear infinite;
	}
	@keyframes scanH {
		0% { top: -10%; }
		100% { top: 110%; }
	}
	@keyframes scanV {
		0% { left: -10%; }
		100% { left: 110%; }
	}

	/* Vector Wave */
	.wave-path, .wave-path-2 {
		animation: waveMove var(--anim-speed) linear infinite;
	}
	@keyframes waveMove {
		from { transform: translateX(0); }
		to { transform: translateX(-500px); }
	}
</style>