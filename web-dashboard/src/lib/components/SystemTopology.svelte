<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { nodes } from '$lib/stores.svelte';
	import { Server, Cpu, Shield, Database } from 'lucide-svelte';
	import { scale } from 'svelte/transition';

	let activeSessions = $derived($nodes.reduce((acc, n) => acc + n.current_instances, 0));
	let nodeCount = $derived($nodes.length);

	interface Projectile {
		id: number;
		x: number;
		y: number;
		targetX: number;
		targetY: number;
		progress: number;
		type: 'data' | 'security' | 'heartbeat';
	}

	let projectiles = $state<Projectile[]>([]);
	let nextId = 0;

	function spawnProjectile(
		type: 'data' | 'security' | 'heartbeat',
		from: { x: number; y: number },
		to: { x: number; y: number }
	) {
		const id = nextId++;
		projectiles.push({
			id,
			x: from.x,
			y: from.y,
			targetX: to.x,
			targetY: to.y,
			progress: 0,
			type
		});
	}

	function updateProjectiles() {
		projectiles = projectiles.filter((p) => {
			p.progress += 0.02;
			p.x = p.x + (p.targetX - p.x) * 0.05;
			p.y = p.y + (p.targetY - p.y) * 0.05;
			return p.progress < 1;
		});
		requestAnimationFrame(updateProjectiles);
	}

	onMount(() => {
		const handle = requestAnimationFrame(updateProjectiles);
		return () => cancelAnimationFrame(handle);
	});

	// Pulsing effect for nodes
	$effect(() => {
		$nodes;
		untrack(() => {
			// Trigger a pulse from Master to a random node
			if ($nodes.length > 0) {
				const targetIdx = Math.floor(Math.random() * $nodes.length);
				spawnProjectile('heartbeat', { x: 500, y: 400 }, getNodePos(targetIdx, $nodes.length));
			}
		});
	});

	function getNodePos(index: number, total: number) {
		const radius = 300;
		const angle = (index / total) * Math.PI * 2;
		return {
			x: 500 + Math.cos(angle) * radius,
			y: 400 + Math.sin(angle) * radius
		};
	}
</script>

<div
	class="w-full h-full bg-slate-950/40 border border-white/5 rounded-[2.5rem] relative overflow-hidden backdrop-blur-xl shadow-2xl group"
>
	<!-- Background Grid -->
	<div class="absolute inset-0 opacity-[0.03] pointer-events-none">
		<svg width="100%" height="100%">
			<defs>
				<pattern id="topology-grid" width="40" height="40" patternUnits="userSpaceOnUse">
					<path d="M 40 0 L 0 0 0 40" fill="none" stroke="white" stroke-width="1" />
				</pattern>
			</defs>
			<rect width="100%" height="100%" fill="url(#topology-grid)" />
		</svg>
	</div>

	<!-- System Labels -->
	<div class="absolute top-8 left-10 z-20 space-y-1 pointer-events-none">
		<h3 class="text-xs font-bold text-slate-500 uppercase tracking-[0.3em]">
			Infrastructure Topology
		</h3>
		<p class="text-[10px] font-bold text-sky-500/60 uppercase tracking-widest">
			Real-time Node Distribution
		</p>
	</div>

	<div class="absolute bottom-8 right-10 z-20 flex gap-8 pointer-events-none">
		<div class="flex flex-col items-end">
			<span class="text-[9px] font-bold text-slate-600 uppercase tracking-widest">Clusters</span>
			<span class="text-sm font-mono font-bold text-white">{nodeCount}</span>
		</div>
		<div class="flex flex-col items-end">
			<span class="text-[9px] font-bold text-slate-600 uppercase tracking-widest">Active Links</span
			>
			<span class="text-sm font-mono font-bold text-sky-400">{activeSessions}</span>
		</div>
	</div>

	<svg viewBox="0 0 1000 800" class="w-full h-full relative z-10">
		<defs>
			<filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
				<feGaussianBlur stdDeviation="3" result="blur" />
				<feComposite in="SourceGraphic" in2="blur" operator="over" />
			</filter>
			<radialGradient id="master-grad" cx="50%" cy="50%" r="50%">
				<stop offset="0%" stop-color="#0ea5e9" stop-opacity="0.2" />
				<stop offset="100%" stop-color="#0ea5e9" stop-opacity="0" />
			</radialGradient>
		</defs>

		<!-- Connection Lines -->
		{#each $nodes as node, i}
			{@const pos = getNodePos(i, $nodes.length)}
			<line
				x1="500"
				y1="400"
				x2={pos.x}
				y2={pos.y}
				stroke="rgba(14, 165, 233, 0.1)"
				stroke-width="1.5"
				stroke-dasharray="4 4"
			/>
		{/each}

		<!-- Projectiles -->
		{#each projectiles as p (p.id)}
			<circle cx={p.x} cy={p.y} r="2" fill="#0ea5e9" filter="url(#glow)">
				<animate attributeName="opacity" values="0;1;0" dur="1s" repeatCount="indefinite" />
			</circle>
		{/each}

		<!-- Master Node -->
		<g class="cursor-pointer">
			<circle cx="500" cy="400" r="100" fill="url(#master-grad)" />
			<circle
				cx="500"
				cy="400"
				r="45"
				fill="#020617"
				stroke="#0ea5e9"
				stroke-width="2"
				filter="url(#glow)"
			/>
			<foreignObject x="475" y="375" width="50" height="50">
				<div class="w-full h-full flex items-center justify-center">
					<Server class="w-6 h-6 text-sky-400" />
				</div>
			</foreignObject>
			<text
				x="500"
				y="465"
				text-anchor="middle"
				class="text-[10px] font-bold fill-sky-400 uppercase tracking-[0.4em]">Core System</text
			>
		</g>

		<!-- Distributed Nodes -->
		{#each $nodes as node, i}
			{@const pos = getNodePos(i, $nodes.length)}
			<g class="cursor-pointer group/node" transition:scale>
				<circle
					cx={pos.x}
					cy={pos.y}
					r="35"
					fill="rgba(2, 6, 23, 0.8)"
					stroke="rgba(255,255,255,0.05)"
					stroke-width="1"
					class="backdrop-blur-xl"
				/>
				<circle
					cx={pos.x}
					cy={pos.y}
					r="20"
					fill="#020617"
					stroke={node.status === 'Online' ? '#10b981' : '#64748b'}
					stroke-width="1.5"
				/>
				<foreignObject x={pos.x - 10} y={pos.y - 10} width="20" height="20">
					<div class="w-full h-full flex items-center justify-center">
						<Cpu
							class="w-3.5 h-3.5 {node.status === 'Online' ? 'text-emerald-400' : 'text-slate-500'}"
						/>
					</div>
				</foreignObject>
				<text
					x={pos.x}
					y={pos.y + 45}
					text-anchor="middle"
					class="text-[8px] font-bold fill-slate-400 uppercase tracking-widest">{node.name}</text
				>
			</g>
		{/each}

		<!-- External Resources -->
		<!-- Security Node -->
		<g transform="translate(150, 150)">
			<circle r="40" fill="rgba(244, 63, 94, 0.05)" />
			<circle r="25" fill="#020617" stroke="#f43f5e" stroke-width="1.5" />
			<foreignObject x="-10" y="-10" width="20" height="20">
				<div class="w-full h-full flex items-center justify-center">
					<Shield class="w-4 h-4 text-rose-500" />
				</div>
			</foreignObject>
			<text
				y="45"
				text-anchor="middle"
				class="text-[8px] font-bold fill-rose-500 uppercase tracking-widest">Security</text
			>
		</g>

		<!-- Database Node -->
		<g transform="translate(850, 650)">
			<circle r="40" fill="rgba(16, 185, 129, 0.05)" />
			<circle r="25" fill="#020617" stroke="#10b981" stroke-width="1.5" />
			<foreignObject x="-10" y="-10" width="20" height="20">
				<div class="w-full h-full flex items-center justify-center">
					<Database class="w-4 h-4 text-emerald-500" />
				</div>
			</foreignObject>
			<text
				y="45"
				text-anchor="middle"
				class="text-[8px] font-bold fill-emerald-500 uppercase tracking-widest">Database</text
			>
		</g>
	</svg>

	<!-- Atmospheric Overlay -->
	<div
		class="absolute inset-0 bg-gradient-to-t from-slate-950 via-transparent to-transparent opacity-40 pointer-events-none"
	></div>
</div>

<style>
</style>
