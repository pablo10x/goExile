<script lang="ts">
	import { onMount } from 'svelte';

	export let status = 'ONLINE';
	export let players = 0;
	export let servers = 0;

	const fmt = (n: number) => n.toLocaleString();

	let visibleText = "";
	let fullText = "";
	let isTyping = true;

	onMount(() => {
		fullText = `SYSTEM: ${status} -- ONLINE: ${fmt(players)} -- SERVERS: ${fmt(servers)}`;
		let i = 0;
		const speed = 30;

		const typeInterval = setInterval(() => {
			if (i < fullText.length) {
				visibleText += fullText.charAt(i);
				i++;
			} else {
				isTyping = false;
				clearInterval(typeInterval);
			}
		}, speed);

		return () => clearInterval(typeInterval);
	});
</script>

<div class="w-full bg-[#050505] border-b-2 border-neutral-800 relative z-40 overflow-hidden font-mono text-[8px] md:text-[10px] tracking-tight">
	<div class="max-w-7xl mx-auto px-2 md:px-4 py-1 flex flex-wrap items-center justify-between gap-2 text-neutral-500 h-6 md:h-auto uppercase font-black">
		
		<div class="flex items-center gap-2 md:gap-4 whitespace-nowrap">
			<span class="text-neutral-500 opacity-80">
				> {visibleText}<span class="{isTyping ? 'animate-pulse' : 'opacity-0'} inline-block w-1.5 h-3 bg-rust align-middle ml-1"></span>
			</span>
		</div>

		<div class="hidden sm:block flex-1 mx-4 overflow-hidden relative h-3 opacity-30">
			<div class="absolute whitespace-nowrap animate-marquee text-[#888888]">
				/// SYSTEM UPDATE: SECTOR 7 PACIFIED /// WEATHER WARNING: ACID RAIN IN ZONE 4 /// NEW FACTION DETECTED /// SYSTEM UPDATE: SECTOR 7 PACIFIED ///
			</div>
		</div>

		<div class="hidden md:flex items-center gap-2 text-[#888888] opacity-40">
			<span>v0.9.4b</span>
		</div>
	</div>
</div>

<style>
	@keyframes marquee {
		0% { transform: translateX(100%); }
		100% { transform: translateX(-100%); }
	}
	.animate-marquee {
		animation: marquee 20s linear infinite;
	}
</style>