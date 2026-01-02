<script lang="ts">
	import { onMount } from 'svelte';

	let { active = false } = $props<{ active?: boolean }>();

	let randomHex = $state<string[]>([]);
	let sparks = $state<{top: number, left: number, delay: number}[]>([]);

	function updateData() {
		if (!active) return;
		randomHex = Array.from({ length: 6 }, () => 
			`0x${Math.random().toString(16).slice(2, 10).toUpperCase()}::OK`
		);
	}

	function initSparks() {
		sparks = Array.from({ length: 6 }, (_, i) => ({
			top: Math.random() * 100,
			left: Math.random() * 100,
			delay: i * 0.5
		}));
	}

	onMount(() => {
		initSparks();
		const interval = setInterval(() => {
			if (active) updateData();
		}, 5000);
		return () => clearInterval(interval);
	});

	$effect(() => {
		if (active && randomHex.length === 0) {
			updateData();
		}
	});
</script>

{#if active}
	<div 
		class="absolute inset-0 pointer-events-none overflow-hidden z-[1]"
		style="--hover-intensity: 0.5"
	>
		<!-- Terminal Data Stream -->
		<div 
			class="absolute inset-0 p-4 font-mono text-[7px] uppercase tracking-tighter select-none overflow-hidden mix-blend-screen"
			style="color: #f97316; opacity: calc(var(--hover-intensity) * 0.6);"
		>
			<div 
				class="animate-terminal-scroll space-y-2 will-change-transform"
				style="animation-duration: 20s"
			>
				{#each randomHex as line}
					<div class="flex flex-col gap-0.5">
						<div class="font-bold text-white/30">{line}</div>
						<div class="opacity-60">UPLINK_STABLE</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Glitch Sparks -->
		<div class="absolute inset-0 overflow-hidden">
			{#each sparks as spark}
				<div 
					class="absolute w-[1.5px] h-[1.5px] bg-white shadow-[0_0_10px_#f97316] opacity-0 animate-glitch-spark will-change-[opacity,transform]" 
					style="top: {spark.top}%; left: {spark.left}%; --delay: {spark.delay}s"
				></div>
			{/each}
		</div>

		<!-- Horizontal Scanning Beam -->
		<div class="absolute w-full h-[1.5px] bg-[#f97316] shadow-[0_0_10px_#f97316] animate-scan-line opacity-30 will-change-[top,opacity]"></div>
	</div>
{/if}

<style>
	@keyframes terminal-scroll {
		from { transform: translateY(0); }
		to { transform: translateY(-50%); }
	}
	.animate-terminal-scroll {
		animation: terminal-scroll linear infinite;
	}

	@keyframes glitch-spark {
		0%, 100% { opacity: 0; transform: scale(0) rotate(0deg); }
		1% { opacity: 1; transform: scale(1.5) rotate(45deg); }
		3% { opacity: 0; transform: scale(0) translate(20px, -20px); }
	}
	.animate-glitch-spark {
		animation: glitch-spark 4s infinite;
		animation-delay: var(--delay);
	}

	@keyframes scan-line {
		0% { top: -10%; opacity: 0; }
		10% { opacity: 1; }
		90% { opacity: 1; }
		100% { top: 110%; opacity: 0; }
	}
	.animate-scan-line {
		animation: scan-line 2.5s ease-in-out infinite;
	}
</style>