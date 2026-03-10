<script lang="ts">
	import { onMount } from 'svelte';

	let { status = 'ONLINE', players = 0, servers = 0 } = $props<{
		status: string;
		players: number;
		servers: number;
	}>();

	const fmt = (n: number) => n.toLocaleString();

	let visibleText = $state('');
	let fullText = '';
	let isTyping = $state(true);

	onMount(() => {
		fullText = `STATUS: ${status} -- SESSIONS: ${fmt(players)} -- NODES: ${fmt(servers)}`;
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

<div
	class="w-full bg-black/40 border-b border-white/5 relative z-40 overflow-hidden font-mono text-[8px] md:text-[9px] tracking-widest"
>
	<div
		class="px-6 py-1 flex flex-wrap items-center justify-between gap-2 text-slate-600 h-6 md:h-auto uppercase font-bold"
	>
		<div class="flex items-center gap-2 md:gap-4 whitespace-nowrap">
			<span class="text-slate-500">
				> {visibleText}<span
					class="{isTyping
						? 'animate-pulse'
						: 'opacity-0'} inline-block w-1.5 h-2.5 bg-sky-500 align-middle ml-1"
				></span>
			</span>
		</div>

		<div class="hidden sm:block flex-1 mx-8 overflow-hidden relative h-3 opacity-20">
			<div class="absolute whitespace-nowrap animate-marquee text-slate-400">
				/// INFRASTRUCTURE OPTIMIZED /// DATABASE REPLICATION ACTIVE /// NETWORK LATENCY: 1.2MS /// ALL SYSTEMS OPERATIONAL ///
			</div>
		</div>

		<div class="hidden md:flex items-center gap-2 text-slate-600 opacity-40">
			<span>BUILD 0.9.4B</span>
		</div>
	</div>
</div>

<style>
	@keyframes marquee {
		0% {
			transform: translateX(100%);
		}
		100% {
			transform: translateX(-100%);
		}
	}
	.animate-marquee {
		animation: marquee 20s linear infinite;
	}
</style>
