<script lang="ts">
	import { 
		notifications, 
		siteSettings, 
		backgroundConfig, 
		theme,
		stats
	} from '$lib/stores';
	import { 
		Palette, 
		Monitor, 
		Zap, 
		CloudRain, 
		Waves, 
		Wind, 
		Activity, 
		Shield, 
		Cloud, 
		RefreshCw, 
		AlertCircle,
		CheckCircle,
		LayoutDashboard,
		Server,
		Database,
		Terminal,
		Search,
		Plus,
		ChevronDown,
		ChevronRight,
		Menu,
		Edit3,
		Globe,
		FileJson,
		Eye,
		Clock,
		HardDrive,
		Network,
		Settings as SettingsIcon,
		Users,
		Sparkles,
		Box,
		Layers,
		Dna,
		Lock,
		Radio,
		Type,
		Move,
		Check
	} from 'lucide-svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import Button from '$lib/components/Button.svelte';
	import Icon from '$lib/components/theme/Icon.svelte';
	import CardHoverOverlay from '$lib/components/theme/CardHoverOverlay.svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	// Calibration Subsystems
	let activeSubsystem = $state<'chromatic' | 'atmospheric' | 'geometric' | 'structural' | 'typography' | 'kinetic' | 'dynamics' | 'interface' | 'buttons'>('chromatic');
	let bgSubTab = $state<'global' | 'cards'>('global');

	const subsystems = [
		{ id: 'chromatic', label: 'Colors', icon: Palette, desc: 'Color palette & brightness' },
		{ id: 'atmospheric', label: 'Effects', icon: Cloud, desc: 'CRT, Scanlines & Overlays' },
		{ id: 'geometric', label: 'Layout', icon: Box, desc: 'Borders, Corners & Sizing' },
		{ id: 'buttons', label: 'Buttons', icon: Zap, desc: 'Action button aesthetics' },
		{ id: 'typography', label: 'Fonts', icon: Type, desc: 'Text styles & spacing' },
		{ id: 'interface', label: 'Interface', icon: LayoutDashboard, desc: 'Dashboard visibility' },
		{ id: 'dynamics', label: 'Animations', icon: Zap, desc: 'Text & UI movement speed' },
		{ id: 'kinetic', label: 'Physics', icon: Move, desc: 'Hover effects & button feel' },
		{ id: 'structural', label: 'Backgrounds', icon: Layers, desc: 'Page background styles' }
	];

	const fontOptions = [
		'Inter', 'Space Grotesk', 'Michroma', 'Orbitron', 'Red Hat Mono', 
		'Syncopate', 'Kanit', 'JetBrains Mono', 'Rajdhani', 'Black Ops One',
		'Roboto', 'Lato', 'Montserrat', 'Poppins', 'Playfair Display', 
		'Fira Code', 'Open Sans'
	];

	const backgroundEngines = [
		{ id: 'digital_stream', label: 'Data Stream', icon: FileJson },
		{ id: 'circuit_grid', label: 'Circuit Board', icon: Network },
		{ id: 'neon_pulse', label: 'Neon Mist', icon: Waves },
		{ id: 'noise_static', label: 'Analog Noise', icon: Activity },
		{ id: 'glass_refraction', label: 'Prism Glass', icon: Box },
		{ id: 'cyber_scan', label: 'Scanning', icon: Terminal },
		{ id: 'vector_wave', label: 'Vector Flow', icon: Globe },
		{ id: 'none', label: 'Minimal Black', icon: Zap }
	];

	// Preset Engine
	const presets = [
		{ 
			id: 'deep_command', 
			name: 'DEEP_COMMAND', 
			accent: '#c2410c', 
			bg: '#020202',
			card_bg: '#1e293b',
			text_dim: '#475569',
			header_bg: '#000000cc',
			sidebar_bg: '#020202cc',
			desc: 'High-alert tactical interface'
		},
		{ 
			id: 'mercury_protocol', 
			name: 'MERCURY_PROTO', 
			accent: '#06b6d4', 
			bg: '#050505',
			card_bg: '#1e293b',
			text_dim: '#64748b',
			header_bg: '#000000cc',
			sidebar_bg: '#050505cc',
			desc: 'Clean, cold digital aesthetic'
		},
		{ 
			id: 'solar_flare', 
			name: 'SOLAR_FLARE', 
			accent: '#f59e0b', 
			bg: '#080705',
			card_bg: '#1c1917',
			text_dim: '#78716c',
			header_bg: '#000000cc',
			sidebar_bg: '#080705cc',
			desc: 'High-luminance emergency hud'
		},
		{ 
			id: 'void_walker', 
			name: 'VOID_WALKER', 
			accent: '#a855f7', 
			bg: '#020105',
			card_bg: '#171717',
			text_dim: '#525252',
			header_bg: '#000000cc',
			sidebar_bg: '#020105cc',
			desc: 'Minimalist stealth operation'
		},
		{ 
			id: 'orbital_station', 
			name: 'ORBITAL_STATION', 
			accent: '#38bdf8', 
			bg: '#18181b',
			card_bg: '#27272a',
			text_dim: '#71717a',
			header_bg: '#18181bcc',
			sidebar_bg: '#18181bcc',
			desc: 'High-visibility balanced interface'
		}
	];

	function applyPreset(p: typeof presets[0]) {
		siteSettings.update(s => ({
			...s,
			aesthetic: {
				...s.aesthetic,
				primary_color: p.accent + 'ff',
				accent_color: p.accent + 'ff',
				bg_color: p.bg + 'ff',
				card_bg_color: (p.card_bg || '#1e293b') + 'ff',
				text_color_dim: (p.text_dim || '#666666') + 'ff',
				header_bg_color: (p.header_bg || '#000000cc'),
				sidebar_bg_color: (p.sidebar_bg || '#18181bcc'),
				card_glow_color: p.accent,
				theme_preset: p.id
			}
		}));
		notifications.add({ type: 'success', message: `CALIBRATION_PRESET: ${p.name}_APPLIED` });
	}

	// Color Helper Functions
	function getHex(color: string) {
		if (!color) return '#000000';
		return color.substring(0, 7);
	}

	function getAlpha(color: string) {
		if (!color || color.length < 9) return 1;
		const hex = color.substring(7, 9);
		return parseInt(hex, 16) / 255;
	}

	function updateColor(key: string, hex: string, alpha: number) {
		const alphaHex = Math.round(alpha * 255).toString(16).padStart(2, '0');
		siteSettings.update(s => ({
			...s,
			aesthetic: {
				...s.aesthetic,
				[key]: hex + alphaHex
			}
		}));
	}

	function resetToDefault() {
		if (confirm('Are you sure you want to reset all aesthetic settings to default?')) {
			location.reload(); 
		}
	}
</script>

<div class="relative z-10 w-full space-y-10 pb-32 font-jetbrains">
	<!-- System Calibration Header -->
	<div class="flex flex-col xl:flex-row xl:items-end justify-between gap-8 border-l-4 border-rust pl-6 sm:pl-10 py-4 bg-[#0a0a0a] industrial-frame shadow-2xl relative overflow-hidden">
		<div class="absolute inset-0 bg-[url('/grid.svg')] bg-center opacity-[0.02] pointer-events-none"></div>
		
		<div class="flex items-center gap-8 relative z-10">
			<div class="p-5 bg-rust/10 border border-rust/30 shadow-2xl industrial-frame group">
				<Dna class="w-12 h-12 text-rust animate-pulse group-hover:scale-110 transition-transform" />
			</div>
			<div>
				<div class="flex items-center gap-3 mb-2">
					<span class="bg-rust text-white px-3 py-0.5 text-[9px] font-black uppercase tracking-[0.3em]">Module: Theme_Lab</span>
					<div class="w-px h-3 bg-stone-800"></div>
					<span class="text-stone-600 text-[9px] font-black uppercase tracking-widest">Access: root_admin</span>
				</div>
				<h1 class="text-5xl font-heading font-black text-white uppercase tracking-tighter leading-none">
					THEME_<span class="text-rust">EDITOR</span>
				</h1>
				<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-[0.4em] font-black mt-3 italic">
					Customize colors, animations and interface styles
				</p>
			</div>
		</div>

		<div class="flex items-center gap-4 p-2 relative z-10">
			<div class="flex bg-black/40 p-1 border border-stone-800">
				<button
					onclick={resetToDefault}
					class="px-6 py-3 hover:bg-white hover:text-black text-stone-500 font-black text-[10px] uppercase tracking-widest transition-all"
				>
					Reset Defaults
				</button>
				<div class="w-px h-10 bg-stone-800"></div>
				<div class="px-6 py-3 flex items-center gap-3">
					<div class="w-2 h-2 bg-emerald-500 animate-flicker"></div>
					<span class="text-[10px] font-black text-emerald-500/70 uppercase tracking-widest">Live Preview</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Main Workbench -->
	<div class="grid grid-cols-1 2xl:grid-cols-12 gap-8 items-start">
		
		<!-- Left: Subsystem Navigation & Controls -->
		<div class="2xl:col-span-8 space-y-8">
			
			<!-- Subsystem Tabs -->
			<div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4">
				{#each subsystems as sub}
					<button
						onclick={() => activeSubsystem = sub.id as any}
						class="flex flex-col items-start p-5 border transition-all duration-500 relative group overflow-hidden {activeSubsystem === sub.id ? 'bg-rust text-white border-rust shadow-2xl shadow-rust/20' : 'bg-stone-950/50 border-stone-800 text-stone-600 hover:border-stone-700'}"
					>
						<div class="flex items-center justify-between w-full mb-4">
							<sub.icon class="w-5 h-5 {activeSubsystem === sub.id ? 'text-white' : 'text-stone-700 group-hover:text-rust'}" />
							<div class="w-1.5 h-1.5 rounded-full {activeSubsystem === sub.id ? 'bg-white animate-pulse' : 'bg-stone-900'}"></div>
						</div>
						<span class="text-[9px] font-black uppercase tracking-widest mb-1">{sub.label}</span>
						
						{#if activeSubsystem === sub.id}
							<div class="absolute bottom-0 left-0 w-full h-0.5 bg-white"></div>
						{/if}
					</button>
				{/each}
			</div>

			<!-- Active Calibration Deck -->
			<div class="modern-industrial-card glass-panel !rounded-none min-h-[600px] flex flex-col border-stone-800">
				<div class="p-8 border-b border-stone-800 bg-stone-950/40 flex items-center justify-between">
					<div class="flex items-center gap-4">
						<div class="w-2 h-8 bg-rust"></div>
						<div>
							<h3 class="font-heading font-black text-lg text-white uppercase tracking-widest">{subsystems.find(s => s.id === activeSubsystem)?.label}</h3>
							<span class="text-[9px] text-stone-600 uppercase tracking-widest italic">Subsystem_Active_Node_0x{activeSubsystem.substring(0,2).toUpperCase()}</span>
						</div>
					</div>
				</div>

				<div class="p-10 flex-1 overflow-y-auto custom-scrollbar bg-[radial-gradient(circle_at_top_right,rgba(249,115,22,0.02),transparent)]">
					{#if activeSubsystem === 'chromatic'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<!-- Primary Matrix -->
							<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-10">
								{#each [
									{ key: 'primary_color', label: 'Primary Accent', desc: 'Main brand & button color' },
									{ key: 'secondary_color', label: 'Secondary Accent', desc: 'Muted highlights' },
									{ key: 'card_bg_color', label: 'Card Background', desc: 'Inner panel surface' },
									{ key: 'text_color_dim', label: 'Dim Labels', desc: 'Small telemetry text' },
									{ key: 'card_glow_color', label: 'Glow Color', desc: 'Pulse & border radiance' },
									{ key: 'text_color_primary', label: 'Primary Text', desc: 'Main reading color' },
									{ key: 'bg_color', label: 'Page Background', desc: 'Deep background color' },
									{ key: 'header_bg_color', label: 'Header Background', desc: 'Top bar color' },
									{ key: 'sidebar_bg_color', label: 'Sidebar Background', desc: 'Left navigation color' },
									{ key: 'terminal_bg_color', label: 'Terminal Background', desc: 'Console & modal background' },
									{ key: 'scrollbar_thumb_color', label: 'Scrollbar Handle', desc: 'Visible part of scrollbar' },
									{ key: 'scrollbar_track_color', label: 'Scrollbar Track', desc: 'Scrollbar background line' },
									{ key: 'success_color', label: 'Success Color', desc: 'Green status indicators' },
									{ key: 'warning_color', label: 'Warning Color', desc: 'Amber status indicators' },
									{ key: 'danger_color', label: 'Danger Color', desc: 'Red status indicators' },
									{ key: 'info_color', label: 'Info Color', desc: 'Blue status indicators' }
								] as color}
									{@const currentVal = ($siteSettings.aesthetic as any)[color.key]}
									<div class="space-y-5">
										<div class="flex justify-between items-start">
											<div class="space-y-1">
												<h4 class="font-black text-white uppercase text-[10px] tracking-widest">{color.label}</h4>
												<p class="text-[8px] text-stone-600 uppercase tracking-tight">{color.desc}</p>
											</div>
											<span class="text-[9px] font-mono text-stone-700 uppercase">{currentVal}</span>
										</div>
										
										<div class="relative h-12 group p-1 bg-stone-900/50 border border-stone-800">
											<input 
												type="color" 
												value={getHex(currentVal)} 
												oninput={e => updateColor(color.key, e.currentTarget.value, getAlpha(currentVal))} 
												class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10" 
											/>
											<div class="absolute inset-1 flex items-center px-4 gap-4 bg-stone-950 overflow-hidden industrial-frame transition-colors group-hover:border-rust/30">
												<div class="w-6 h-6 border border-white/5 shadow-xl relative shrink-0">
													<div class="absolute inset-0 bg-[url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAYAAACp8Z5+AAAAAXNSR0IArs4c6QAAACBJREFUGFdjZEADJEYGhj9MDIyMDAwMDAwMDAwMDAwMDAwMABYDAwMAAAAASUVORK5CYII=')] opacity-20"></div>
													<div class="absolute inset-0" style="background-color: {currentVal}"></div>
												</div>
												<span class="text-[10px] font-mono text-stone-500 uppercase tracking-tighter">HEX_VAL: {getHex(currentVal)}</span>
											</div>
										</div>

										<div class="space-y-2 px-1">
											<div class="flex justify-between items-center">
												<span class="text-[8px] font-mono text-stone-700 uppercase tracking-widest">Opacity_Calibration</span>
												<span class="text-[9px] font-mono text-rust-light">{(getAlpha(currentVal) * 100).toFixed(0)}%</span>
											</div>
											<div class="relative flex items-center h-1 bg-stone-900">
												<input 
													type="range" 
													min="0" 
													max="1" 
													step="0.01" 
													value={getAlpha(currentVal)} 
													oninput={e => updateColor(color.key, getHex(currentVal), parseFloat(e.currentTarget.value))} 
													class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
												/>
												<div class="absolute top-0 left-0 h-full bg-rust" style="width: {getAlpha(currentVal) * 100}%"></div>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{:else if activeSubsystem === 'typography'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Font_Family_Selection</h4>
									
									<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
										<div class="space-y-3">
											<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block ml-1">Header_Typeface</span>
											<select 
												bind:value={$siteSettings.aesthetic.font_header}
												class="w-full bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-[10px] text-stone-300 focus:border-rust outline-none uppercase tracking-widest appearance-none cursor-pointer"
											>
												{#each fontOptions as font}
													<option value={font}>{font}</option>
												{/each}
											</select>
										</div>
										<div class="space-y-3">
											<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block ml-1">Body_Typeface</span>
											<select 
												bind:value={$siteSettings.aesthetic.font_body}
												class="w-full bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-[10px] text-stone-300 focus:border-rust outline-none uppercase tracking-widest appearance-none cursor-pointer"
											>
												{#each fontOptions as font}
													<option value={font}>{font}</option>
												{/each}
											</select>
										</div>
										<div class="space-y-3">
											<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block ml-1">Terminal_Typeface</span>
											<select 
												bind:value={$siteSettings.aesthetic.font_mono}
												class="w-full bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-[10px] text-stone-300 focus:border-rust outline-none uppercase tracking-widest appearance-none cursor-pointer"
											>
												{#each fontOptions as font}
													<option value={font}>{font}</option>
												{/each}
											</select>
										</div>
									</div>

									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4 mt-12">Micro_Typography_Rules</h4>
									
									{#each [
										{ key: 'letter_spacing', label: 'Tracking_Spread', min: -0.05, max: 0.5, step: 0.01, unit: 'em' },
										{ key: 'line_height', label: 'Vertical_Density', min: 1, max: 2, step: 0.1, unit: 'lh' },
										{ key: 'paragraph_spacing', label: 'Block_Separation', min: 0, max: 3, step: 0.1, unit: 'em' },
										{ key: 'font_size_base', label: 'Kernel_Size', min: 10, max: 20, step: 1, unit: 'px' },
										{ key: 'text_glow_intensity', label: 'Glow_Radiation', min: 0, max: 1, step: 0.05, unit: 'α' }
									] as slider}
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{slider.label}</span>
												<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[slider.key])}{slider.unit}</span>
											</div>
											<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
												<input 
													type="range" 
													min={slider.min} 
													max={slider.max} 
													step={slider.step} 
													value={($siteSettings.aesthetic as any)[slider.key]} 
													oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [slider.key]: parseFloat(e.currentTarget.value) } }))} 
													class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
												/>
												<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic as any)[slider.key] - slider.min) / (slider.max - slider.min)) * 100}%"></div>
											</div>
										</div>
									{/each}

									<div class="grid grid-cols-1 gap-6 pt-4 border-t border-stone-800/50">
										<div class="space-y-4">
											<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block">Heading_Weight_Level</span>
											<div class="grid grid-cols-6 gap-2">
												{#each ['100', '300', '400', '500', '700', '900'] as weight}
													<button
														onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, heading_weight: weight } }))}
														class="py-2 border text-[9px] font-black transition-all {$siteSettings.aesthetic.heading_weight === weight ? 'bg-rust text-white border-rust' : 'bg-stone-950 border-stone-800 text-stone-600'}"
													>
														{weight}
													</button>
												{/each}
											</div>
										</div>
										<div class="space-y-4">
											<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block">Base_Weight_Level</span>
											<div class="grid grid-cols-6 gap-2">
												{#each ['100', '300', '400', '500', '700', '900'] as weight}
													<button
														onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, font_weight_base: weight } }))}
														class="py-2 border text-[9px] font-black transition-all {$siteSettings.aesthetic.font_weight_base === weight ? 'bg-rust text-white border-rust' : 'bg-stone-950 border-stone-800 text-stone-600'}"
													>
														{weight}
													</button>
												{/each}
											</div>
										</div>
										<div class="space-y-4">
											<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block">Terminal_Weight_Level</span>
											<div class="grid grid-cols-6 gap-2">
												{#each ['100', '300', '400', '500', '700', '900'] as weight}
													<button
														onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, font_weight_mono: weight } }))}
														class="py-2 border text-[9px] font-black transition-all {$siteSettings.aesthetic.font_weight_mono === weight ? 'bg-rust text-white border-rust' : 'bg-stone-950 border-stone-800 text-stone-600'}"
													>
														{weight}
													</button>
												{/each}
											</div>
										</div>
									</div>

									<div class="space-y-4 pt-4 border-t border-stone-800/50">
										<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block">Text_Transformation_Protocol</span>
										<div class="grid grid-cols-2 gap-4">
											{#each ['uppercase', 'none'] as mode}
												<button
													onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, text_transform: mode as any } }))}
													class="py-3 border-2 text-[10px] font-black uppercase transition-all {$siteSettings.aesthetic.text_transform === mode ? 'bg-rust/10 border-rust text-white' : 'bg-stone-950 border-stone-800 text-stone-600'}"
												>
													{mode}
												</button>
											{/each}
										</div>
									</div>
								</div>

								<div class="space-y-8 bg-stone-950/30 p-8 border border-stone-800 industrial-frame">
									<h4 class="text-[10px] font-black text-rust uppercase tracking-widest italic mb-4">Uplink_Text_Preview</h4>
									<div class="space-y-6">
										<div class="space-y-2">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Heading_H1</span>
											<h1 class="text-3xl text-white truncate">Critical_Breach</h1>
										</div>
										<div class="space-y-2">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Body_Paragraph</span>
											<p class="text-xs text-stone-400 leading-relaxed uppercase">Neural interface synchronization complete. All decentralized modules are reporting nominal operational capacity within sector Delta-9.</p>
										</div>
										<div class="space-y-2">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Telemetry_Data</span>
											<div class="font-mono text-[10px] text-emerald-500">0xAF77 :: STATUS_OK :: 124.5Hz</div>
										</div>
									</div>
								</div>
							</div>
						</div>
					{:else if activeSubsystem === 'buttons'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<div class="space-y-10">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Global Button Geometry</h4>
									
									<div class="space-y-8">
										{#each [
											{ key: 'border_radius', label: 'Corner Rounding', min: 0, max: 24, step: 1, unit: 'px' }
										] as slider}
											<div class="space-y-4">
												<div class="flex justify-between items-center">
													<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{slider.label}</span>
													<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.buttons[slider.key]}{slider.unit}</span>
												</div>
												<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
													<input 
														type="range" 
														min={slider.min} 
														max={slider.max} 
														step={slider.step} 
														value={$siteSettings.aesthetic.buttons[slider.key]} 
														oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, buttons: { ...s.aesthetic.buttons, [slider.key]: parseFloat(e.currentTarget.value) } } }))} 
														class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
													/>
													<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic.buttons as any)[slider.key] - slider.min) / (slider.max - slider.min)) * 100}%"></div>
												</div>
											</div>
										{/each}

										<div class="grid grid-cols-2 gap-6">
											<div class="space-y-3">
												<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block ml-1">Font_Weight</span>
												<select 
													bind:value={$siteSettings.aesthetic.buttons.font_weight}
													class="w-full bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-[10px] text-stone-300 focus:border-rust outline-none uppercase tracking-widest appearance-none cursor-pointer"
												>
													{#each ['100', '300', '400', '500', '700', '900'] as weight}
														<option value={weight}>{weight}</option>
													{/each}
												</select>
											</div>
											<div class="space-y-3">
												<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block ml-1">Text_Transform</span>
												<select 
													bind:value={$siteSettings.aesthetic.buttons.text_transform}
													class="w-full bg-stone-950 border border-stone-800 px-4 py-3 font-jetbrains text-[10px] text-stone-300 focus:border-rust outline-none uppercase tracking-widest appearance-none cursor-pointer"
												>
													<option value="uppercase">UPPERCASE</option>
													<option value="none">NONE</option>
													<option value="capitalize">Capitalize</option>
												</select>
											</div>
										</div>
									</div>

									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4 mt-12">Variant Calibration</h4>
									
									<div class="space-y-10">
										{#each ['primary', 'secondary', 'danger', 'ghost'] as variant}
											<div class="space-y-6 p-6 bg-stone-900/30 border border-stone-800 industrial-frame">
												<h5 class="text-[10px] font-black text-white uppercase tracking-widest flex items-center gap-3">
													<div class="w-1.5 h-1.5 rounded-full" style="background-color: {$siteSettings.aesthetic.buttons[variant].bg_color}"></div>
													{variant.toUpperCase()} STYLE
												</h5>
												
												<div class="grid grid-cols-2 gap-6">
													{#each [
														{ key: 'bg_color', label: 'Background' },
														{ key: 'text_color', label: 'Text' },
														{ key: 'border_color', label: 'Border' },
														{ key: 'hover_bg', label: 'Hover BG' }
													] as color}
														{@const currentVal = ($siteSettings.aesthetic.buttons as any)[variant][color.key]}
														<div class="space-y-3">
															<span class="text-[8px] font-black text-stone-600 uppercase tracking-widest block ml-1">{color.label}</span>
															<div class="relative h-10 group p-1 bg-stone-950 border border-stone-800">
																<input 
																	type="color" 
																	value={getHex(currentVal)} 
																	oninput={e => {
																		const hex = e.currentTarget.value;
																		const alpha = getAlpha(currentVal);
																		const alphaHex = Math.round(alpha * 255).toString(16).padStart(2, '0');
																		siteSettings.update(s => ({
																			...s,
																			aesthetic: {
																				...s.aesthetic,
																				buttons: {
																					...s.aesthetic.buttons,
																					[variant]: {
																						...s.aesthetic.buttons[variant],
																						[color.key]: hex + alphaHex
																					}
																				}
																			}
																		}));
																	}} 
																	class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10" 
																/>
																<div class="absolute inset-1 flex items-center px-3 gap-3 bg-stone-900 overflow-hidden industrial-frame">
																	<div class="w-4 h-4 border border-white/5 relative shrink-0">
																		<div class="absolute inset-0" style="background-color: {currentVal}"></div>
																	</div>
																	<span class="text-[9px] font-mono text-stone-500 uppercase">{getHex(currentVal)}</span>
																</div>
															</div>
														</div>
													{/each}
												</div>
											</div>
										{/each}
									</div>
								</div>

								<div class="space-y-8 bg-stone-950/30 p-8 border border-stone-800 industrial-frame">
									<h4 class="text-[10px] font-black text-rust uppercase tracking-widest italic mb-4">Button_Stress_Test</h4>
									<div class="space-y-10">
										<div class="space-y-4">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Variant_Manifest</span>
											<div class="grid grid-cols-1 gap-4">
												<Button variant="primary">Primary_Action</Button>
												<Button variant="secondary">Secondary_Action</Button>
												<Button variant="danger">Danger_Protocol</Button>
												<Button variant="ghost">Ghost_Interface</Button>
											</div>
										</div>

										<div class="space-y-4">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Scale_Rule</span>
											<div class="flex flex-wrap items-end gap-4">
												<Button size="xs">Small</Button>
												<Button size="sm">Medium</Button>
												<Button size="md">Large</Button>
												<Button size="lg">Extra</Button>
											</div>
										</div>

										<div class="space-y-4">
											<span class="text-[8px] text-stone-700 uppercase font-mono">State_Validation</span>
											<div class="grid grid-cols-2 gap-4">
												<Button disabled>Disabled_State</Button>
												<Button loading>Processing...</Button>
											</div>
										</div>
									</div>
								</div>
							</div>
						</div>
					{:else if activeSubsystem === 'interface'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Dashboard Visibility</h4>
									
									<div class="space-y-4">
										{#each [
											{ key: 'show_topology', label: 'System Topology', desc: 'Visual network node map' },
											{ key: 'show_stats_cards', label: 'Primary Stats Cards', desc: 'Uptime, Nodes, Requests, Errors' },
											{ key: 'show_traffic_card', label: 'Network Traffic', desc: 'Inbound/Outbound bandwidth telemetry' },
											                                            { key: 'show_db_card', label: 'Database Status', desc: 'Active connections & pool health' },
											                                            { key: 'show_nodes_table', label: 'Node Management', desc: 'List of synchronized nodes' },
											                                            { key: 'compact_mode', label: 'Compact Interface', desc: 'Reduce padding and font sizes' }
											                                        ] as toggle}
											<div class="flex items-center justify-between p-5 bg-stone-900/30 border border-stone-800 hover:border-rust/30 transition-all group">
												<div class="space-y-1">
													<span class="text-[10px] font-black text-white uppercase tracking-widest block">{toggle.label}</span>
													<span class="text-[8px] text-stone-600 uppercase tracking-tight italic">{toggle.desc}</span>
												</div>
												<label class="relative inline-flex items-center cursor-pointer">
													<input type="checkbox" checked={$siteSettings.dashboard[toggle.key as keyof typeof $siteSettings.dashboard]} onchange={e => siteSettings.update(s => ({ ...s, dashboard: { ...s.dashboard, [toggle.key]: e.currentTarget.checked } }))} class="sr-only peer">
													<div class="w-12 h-6 bg-stone-950 border border-stone-800 peer-focus:outline-none rounded-none peer-checked:after:translate-x-6 peer-checked:after:bg-rust after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-stone-700 after:rounded-none after:h-4 after:w-4 after:transition-all peer-checked:bg-rust/10 peer-checked:border-rust"></div>
												</label>
											</div>
										{/each}
									</div>
								</div>

								<div class="space-y-8 bg-stone-950/30 p-8 border border-stone-800 industrial-frame flex flex-col items-center justify-center">
									<div class="text-center space-y-4">
										<LayoutDashboard class="w-16 h-16 text-rust/20 mx-auto" />
										<p class="text-[10px] text-stone-600 uppercase tracking-widest leading-relaxed max-w-xs">
											Visibility changes are applied instantly. Use these toggles to declutter your tactical workspace.
										</p>
									</div>
								</div>
							</div>
						</div>
					{:else if activeSubsystem === 'dynamics'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Global Animation Speeds</h4>
									
									{#each [
										{ key: 'typing_speed', label: 'Text Typing Speed', min: 0.1, max: 5, step: 0.1, unit: 'x' },
										{ key: 'glitch_frequency', label: 'UI Glitch Rate', min: 0.1, max: 10, step: 0.1, unit: 'hz' },
										{ key: 'header_pulse_speed', label: 'Pulse Speed', min: 0.1, max: 5, step: 0.1, unit: 'x' },
										{ key: 'navbar_entry_delay', label: 'Menu Entry Delay', min: 0, max: 1000, step: 50, unit: 'ms' },
										{ key: 'progress_fill_speed', label: 'Loading Bar Speed', min: 0.1, max: 5, step: 0.1, unit: 'x' },
										{ key: 'global_animation_scale', label: 'Overall Speed Multiplier', min: 0, max: 2, step: 0.1, unit: 'x' }
									] as slider}
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{slider.label}</span>
												<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[slider.key])}{slider.unit}</span>
											</div>
											<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
												<input 
													type="range" 
													min={slider.min} 
													max={slider.max} 
													step={slider.step} 
													value={($siteSettings.aesthetic as any)[slider.key]} 
													oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [slider.key]: parseFloat(e.currentTarget.value) } }))} 
													class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
												/>
												<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic as any)[slider.key] - slider.min) / (slider.max - slider.min)) * 100}%"></div>
											</div>
										</div>
									{/each}
								</div>

								<div class="space-y-8 bg-stone-950/30 p-8 border border-stone-800 industrial-frame">
									<h4 class="text-[10px] font-black text-rust uppercase tracking-widest italic mb-4">Signal_Dynamics_Test</h4>
									<div class="space-y-8">
										<div class="space-y-2">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Typing_Reveal_Test</span>
											<h1 class="text-2xl text-white animate-typing overflow-hidden">ESTABLISHING_LINK...</h1>
										</div>
										
										<div class="space-y-2">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Luminous_Pulse_Test</span>
											<div class="h-1 w-full bg-rust animate-pulse shadow-[0_0_15px_var(--color-rust)]"></div>
										</div>

										<div class="space-y-2">
											<span class="text-[8px] text-stone-700 uppercase font-mono">Data_Stream_Test</span>
											<div class="h-4 w-full bg-black border border-stone-800 p-0.5 overflow-hidden">
												<div class="h-full bg-rust-light animate-[progress-width_var(--progress-speed)_ease-in-out_infinite]"></div>
											</div>
										</div>
									</div>
								</div>
							</div>
						</div>
					{:else if activeSubsystem === 'kinetic'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Interaction Sizing & Movement</h4>
									
									{#each [
										{ key: 'global_transition_speed', label: 'Transition Duration', min: 0, max: 1000, step: 50, unit: 'ms' },
										{ key: 'hover_scale_factor', label: 'Hover Expansion Amount', min: 1, max: 1.2, step: 0.01, unit: 'x' },
										{ key: 'button_press_depth', label: 'Button Click Depth', min: 0, max: 10, step: 1, unit: 'px' },
										{ key: 'ui_animation_intensity', label: 'Distortion Intensity', min: 0, max: 2, step: 0.1, unit: 'α' },
										{ key: 'spin_velocity', label: 'Spinner Rotation Speed', min: 0.1, max: 5, step: 0.1, unit: 'x' },
										{ key: 'scanline_speed_hz', label: 'Scanline Sweep Speed', min: 0.1, max: 5, step: 0.1, unit: 'hz' },
										{ key: 'modal_entry_speed', label: 'Window Opening Speed', min: 100, max: 1000, step: 50, unit: 'ms' },
										{ key: 'glow_pulse_depth', label: 'Glow Brightness', min: 0, max: 1, step: 0.05, unit: 'α' }
									] as slider}
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{slider.label}</span>
												<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[slider.key])}{slider.unit}</span>
											</div>
											<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
												<input 
													type="range" 
													min={slider.min} 
													max={slider.max} 
													step={slider.step} 
													value={($siteSettings.aesthetic as any)[slider.key]} 
													oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [slider.key]: parseFloat(e.currentTarget.value) } }))} 
													class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
												/>
												<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic as any)[slider.key] - slider.min) / (slider.max - slider.min)) * 100}%"></div>
											</div>
										</div>
									{/each}
								</div>

								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Kinetic_Stress_Test</h4>
									<div class="grid grid-cols-1 gap-6">
										<button class="modern-industrial-card glass-panel p-8 text-center group">
											<span class="text-sm font-black uppercase group-hover:text-rust transition-colors">Hover_Motion_Test</span>
										</button>
										<div class="flex gap-4">
											{#each [1,2,3] as i}
												<button class="flex-1 py-4 bg-rust text-white font-black text-[10px] uppercase shadow-lg shadow-rust/20 active:translate-y-[var(--press-depth)]">Press_0{i}</button>
											{/each}
										</div>
									</div>
								</div>
							</div>
						</div>
					{:else if activeSubsystem === 'atmospheric'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<!-- Signal Quality -->
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Screen Filters & Overlays</h4>
									
									{#each [
										{ key: 'crt_effect', label: 'CRT Screen Effect', desc: 'Scanline & phosphor simulation' },
										{ key: 'crt_curve', label: 'Curved Screen', desc: 'Slight spherical distortion' },
										{ key: 'text_glow', label: 'Text Radiance', desc: 'Soft glow around characters' },
										{ key: 'panic_mode', label: 'Emergency Mode', desc: 'Global red alert flashing' }
									] as toggle}
										<div class="flex items-center justify-between p-5 bg-stone-900/30 border border-stone-800 hover:border-rust/30 transition-all group">
											<div class="space-y-1">
												<span class="text-[10px] font-black text-white uppercase tracking-widest block">{toggle.label}</span>
												<span class="text-[8px] text-stone-600 uppercase tracking-tight italic">{toggle.desc}</span>
											</div>
											<label class="relative inline-flex items-center cursor-pointer">
												<input type="checkbox" checked={($siteSettings.aesthetic as any)[toggle.key]} onchange={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [toggle.key]: e.currentTarget.checked } }))} class="sr-only peer">
												<div class="w-12 h-6 bg-stone-950 border border-stone-800 peer-focus:outline-none rounded-none peer-checked:after:translate-x-6 peer-checked:after:bg-rust after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-stone-700 after:rounded-none after:h-4 after:w-4 after:transition-all peer-checked:bg-rust/10 peer-checked:border-rust"></div>
											</label>
										</div>
									{/each}
								</div>

								<!-- Analog Interference -->
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Noise & Distortion Settings</h4>
									
									{#each [
										{ key: 'glitch_intensity', label: 'Glitch Intensity', min: 0, max: 1, step: 0.01, unit: 'Hz' },
										{ key: 'chromatic_aberration', label: 'Color Fringing', min: 0, max: 10, step: 0.5, unit: 'px' },
										{ key: 'scanlines_opacity', label: 'Scanline Density', min: 0, max: 0.5, step: 0.01, unit: 'α' },
										{ key: 'noise_opacity', label: 'Film Grain Amount', min: 0, max: 0.2, step: 0.01, unit: 'α' },
										{ key: 'vignette_intensity', label: 'Edge Shadow Darkening', min: 0, max: 1, step: 0.05, unit: 'α' }
									] as slider}
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{slider.label}</span>
												<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[slider.key])}{slider.unit}</span>
											</div>
											<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
												<input 
													type="range" 
													min={slider.min} 
													max={slider.max} 
													step={slider.step} 
													value={($siteSettings.aesthetic as any)[slider.key]} 
													oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [slider.key]: parseFloat(e.currentTarget.value) } }))} 
													class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
												/>
												<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic as any)[slider.key] - slider.min) / (slider.max - slider.min)) * 100}%"></div>
											</div>
										</div>
									{/each}
								</div>
							</div>
						</div>
					{:else if activeSubsystem === 'geometric'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<!-- Frame Geometry -->
								<div class="space-y-8">
									<div class="space-y-6">
										<span class="text-[9px] font-black text-stone-600 uppercase tracking-[0.3em] ml-1">Icon Style Pack</span>
										<div class="grid grid-cols-5 gap-2">
											{#each ['lucide', 'mdi', 'ph', 'ri', 'tabler'] as pack}
												<button
													onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, icon_pack: pack as any } }))}
													class="py-3 border-2 flex flex-col items-center justify-center gap-2 transition-all {$siteSettings.aesthetic.icon_pack === pack ? 'bg-rust/10 border-rust text-white' : 'bg-stone-950/50 border-stone-800 text-stone-600 hover:border-stone-700'}"
												>
													<span class="text-[9px] font-black uppercase tracking-widest">{pack}</span>
												</button>
											{/each}
										</div>
									</div>

									<div class="space-y-4">
										<div class="flex justify-between items-center">
											<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">Icon Line Weight</span>
											<span class="text-[10px] font-mono text-rust-light">{$siteSettings.aesthetic.icon_stroke}px</span>
										</div>
										<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
											<input 
												type="range" 
												min="0.5" 
												max="3" 
												step="0.1" 
												value={$siteSettings.aesthetic.icon_stroke} 
												oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, icon_stroke: parseFloat(e.currentTarget.value) } }))} 
												class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
											/>
											<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {(($siteSettings.aesthetic.icon_stroke - 0.5) / 2.5) * 100}%"></div>
										</div>
									</div>

									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4 mt-12">Global Layout Rules</h4>
									
									<div class="space-y-6">
										<span class="text-[9px] font-black text-stone-600 uppercase tracking-[0.3em] ml-1">Corner Style</span>
										<div class="grid grid-cols-3 gap-3">
											{#each ['clipped', 'rounded', 'sharp'] as style}
												<button
													onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, corner_style: style as any, industrial_styling: style === 'clipped' || style === 'sharp' } }))}
													class="py-4 border-2 flex flex-col items-center justify-center gap-2 transition-all {$siteSettings.aesthetic.corner_style === style ? 'bg-rust/10 border-rust text-white' : 'bg-stone-950 border-stone-800 text-stone-600 hover:border-stone-700'}"
												>
													<div class="w-6 h-6 border-2 border-current {style === 'rounded' ? 'rounded-md' : style === 'clipped' ? 'industrial-frame border-0 !bg-current !w-4 !h-4' : ''}"></div>
													<span class="text-[9px] font-black uppercase tracking-widest">{style}</span>
												</button>
											{/each}
										</div>
									</div>

									{#each [
										{ key: 'border_radius_lg', label: 'Corner Roundness', max: 40 },
										{ key: 'card_border_width', label: 'Border Thickness', max: 5 },
										{ key: 'card_shadow_size', label: 'Shadow Depth', max: 100 },
										{ key: 'sidebar_width', label: 'Sidebar Width', min: 200, max: 400 },
										{ key: 'grid_opacity', label: 'Background Grid Opacity', max: 0.2, step: 0.01 }
									] as radius}
										<div class="space-y-4">
											<div class="flex justify-between items-center">
												<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{radius.label}</span>
												<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[radius.key])}</span>
											</div>
											<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
												<input 
													type="range" 
													min={radius.min || 0} 
													max={radius.max} 
													step={radius.step || 1} 
													value={($siteSettings.aesthetic as any)[radius.key]} 
													oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [radius.key]: parseFloat(e.currentTarget.value) } }))} 
													class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
												/>
												<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic as any)[radius.key] - (radius.min || 0)) / (radius.max - (radius.min || 0))) * 100}%"></div>
											</div>
										</div>
									{/each}
								</div>
							</div>
						</div>
					{:else}
						<!-- Backgrounds Subsystem -->
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<!-- Sub-navigation -->
							<div class="flex gap-4 border-b border-stone-800 pb-4">
								<button 
									onclick={() => bgSubTab = 'global'}
									class="px-6 py-2 text-[10px] font-black uppercase tracking-widest transition-all {bgSubTab === 'global' ? 'bg-rust text-white shadow-lg shadow-rust/20' : 'bg-stone-900/50 text-stone-600 hover:text-stone-400 hover:bg-stone-900'}"
								>Global Background</button>
								<button 
									onclick={() => bgSubTab = 'cards'}
									class="px-6 py-2 text-[10px] font-black uppercase tracking-widest transition-all {bgSubTab === 'cards' ? 'bg-rust text-white shadow-lg shadow-rust/20' : 'bg-stone-900/50 text-stone-600 hover:text-stone-400 hover:bg-stone-900'}"
								>Card Hover Effects</button>
							</div>

							{#if bgSubTab === 'global'}
								<div in:fade={{ duration: 200 }} class="space-y-12">
									<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
										<div class="space-y-8">
											<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Background Visual Styles</h4>
											<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
												{#each backgroundEngines as engine}
													{@const isSelected = $backgroundConfig.global_type === engine.id}
													<button 
														onclick={() => backgroundConfig.update(b => ({ ...b, global_type: engine.id as any }))}
														class="p-6 border transition-all flex flex-col items-center gap-4 relative overflow-hidden {isSelected ? 'bg-rust/10 border-rust shadow-xl' : 'bg-stone-950/50 border-stone-800 text-stone-600 hover:border-stone-700'}"
													>
														<engine.icon class="w-8 h-8 {isSelected ? 'text-rust-light' : 'text-stone-700'}" />
														<span class="text-[10px] font-black uppercase tracking-[0.2em] {isSelected ? 'text-white' : ''}">{engine.label}</span>
														{#if isSelected}
															<div class="absolute top-0 right-0 w-8 h-8 bg-rust flex items-center justify-center translate-x-4 translate-y-[-16px] rotate-45">
																<Check class="w-3 h-3 text-white -rotate-45 translate-y-2" />
															</div>
														{/if}
													</button>
												{/each}
											</div>
										</div>

										<div class="space-y-8">
											<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Style Calibration</h4>
											
											<div class="space-y-10">
												<!-- Global Overrides -->
												<div class="space-y-6">
													<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block ml-1 italic">Global_Directives</span>
													{#each [
														{ key: 'bg_anim_speed', label: 'Master Animation Velocity', min: 0.1, max: 5, step: 0.1, unit: 'x' },
														{ key: 'bg_anim_opacity', label: 'Master Background Density', min: 0, max: 1, step: 0.01, unit: 'α' }
													] as slider}
														<div class="space-y-4">
															<div class="flex justify-between items-center">
																<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{slider.label}</span>
																<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[slider.key])}{slider.unit}</span>
															</div>
															<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
																<input 
																	type="range" 
																	min={slider.min} 
																	max={slider.max} 
																	step={slider.step} 
																	value={($siteSettings.aesthetic as any)[slider.key]} 
																	oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [slider.key]: parseFloat(e.currentTarget.value) } }))} 
																	class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
																/>
																<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic as any)[slider.key] - slider.min) / (slider.max - slider.min)) * 100}%"></div>
															</div>
														</div>
													{/each}
												</div>

												<!-- Atmospheric Overlays -->
												<div class="space-y-6 pt-6 border-t border-stone-800/50">
													<span class="text-[9px] font-black text-stone-600 uppercase tracking-widest block ml-1 italic">Atmospheric_Layers</span>
													<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
														<div class="flex items-center justify-between p-4 bg-rust/5 border border-rust/30 col-span-full">
															<div class="flex flex-col">
																<span class="text-[9px] font-black text-white uppercase tracking-widest">Master Background Display</span>
																<span class="text-[7px] text-stone-500 uppercase">Enable or disable all engine-driven backgrounds</span>
															</div>
															<label class="relative inline-flex items-center cursor-pointer">
																<input type="checkbox" checked={$backgroundConfig.show_global_background} onchange={e => backgroundConfig.update(b => ({ ...b, show_global_background: e.currentTarget.checked }))} class="sr-only peer">
																<div class="w-10 h-5 bg-stone-950 border border-stone-800 rounded-none peer-checked:bg-rust/20 peer-checked:border-rust after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-stone-700 after:h-3 after:w-3 after:transition-all peer-checked:after:translate-x-5 peer-checked:after:bg-rust"></div>
															</label>
														</div>
														{#each [
															{ key: 'show_smoke', label: 'Volumetric Smoke' },
															{ key: 'show_rain', label: 'Particle Precipitation' },
															{ key: 'show_clouds', label: 'Nebula Drifting' },
															{ key: 'show_vignette', label: 'Peripheral Shading' },
															{ key: 'show_navbar_particles', label: 'Navbar Ash Particles' }
														] as toggle}
															<div class="flex items-center justify-between p-4 bg-stone-900/30 border border-stone-800">
																<span class="text-[9px] font-black text-stone-400 uppercase tracking-widest">{toggle.label}</span>
																<label class="relative inline-flex items-center cursor-pointer">
																	<input type="checkbox" checked={($backgroundConfig as any)[toggle.key]} onchange={e => backgroundConfig.update(b => ({ ...b, [toggle.key]: e.currentTarget.checked }))} class="sr-only peer">
																	<div class="w-10 h-5 bg-stone-950 border border-stone-800 rounded-none peer-checked:bg-rust/20 peer-checked:border-rust after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-stone-700 after:h-3 after:w-3 after:transition-all peer-checked:after:translate-x-5 peer-checked:after:bg-rust"></div>
																</label>
															</div>
														{/each}
													</div>
												</div>

												<!-- Active Engine Tuning -->
												{#if $backgroundConfig.global_type !== 'none'}
													{@const type = $backgroundConfig.global_type}
													{@const engineSettings = ($backgroundConfig.settings as any)[type]}
													<div class="space-y-6 pt-6 border-t border-stone-800/50" in:slide>
														<span class="text-[9px] font-black text-rust uppercase tracking-widest block ml-1 italic">Engine_Tuning: {type.toUpperCase()}</span>
														
														<div class="space-y-6">
															<!-- Color Tuning -->
															<div class="flex justify-between items-center">
																<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">Active_Chrome</span>
																<div class="flex items-center gap-3">
																	<span class="text-[10px] font-mono text-stone-600 uppercase">{engineSettings.color}</span>
																	<input 
																		type="color" 
																		value={engineSettings.color} 
																		oninput={e => backgroundConfig.update(b => {
																			const s = { ...b.settings };
																			(s as any)[type].color = e.currentTarget.value;
																			return { ...b, settings: s };
																		})}
																		class="w-8 h-8 bg-transparent border-0 cursor-pointer"
																	/>
																</div>
															</div>

															<!-- Numeric Parameters -->
															{#each [
																{ key: 'speed', label: 'Phase Velocity', min: 0.1, max: 10, step: 0.1, unit: 'x' },
																{ key: 'opacity', label: 'Substrate Density', min: 0, max: 1, step: 0.05, unit: 'α' },
																{ key: 'scale', label: 'Geometric Scale', min: 0.5, max: 3, step: 0.1, unit: 'x' }
															] as param}
																<div class="space-y-4">
																	<div class="flex justify-between items-center">
																		<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{param.label}</span>
																		<span class="text-[10px] font-mono text-rust-light">{engineSettings[param.key]}{param.unit}</span>
																	</div>
																	<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
																		<input 
																			type="range" 
																			min={param.min} 
																			max={param.max} 
																			step={param.step} 
																			value={engineSettings[param.key]} 
																			oninput={e => backgroundConfig.update(b => {
																				const s = { ...b.settings };
																				(s as any)[type][param.key] = parseFloat(e.currentTarget.value);
																				return { ...b, settings: s };
																			})} 
																			class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
																		/>
																		<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((engineSettings[param.key] - param.min) / (param.max - param.min)) * 100}%"></div>
																	</div>
																</div>
															{/each}
														</div>
													</div>
												{/if}
											</div>
										</div>
									</div>
								</div>
							{:else}
								<div in:fade={{ duration: 200 }} class="space-y-12">
									<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
										<div class="space-y-8">
											<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Hover Effect Toggle</h4>
											<div class="flex items-center justify-between p-5 bg-stone-900/30 border border-stone-800 hover:border-rust/30 transition-all group">
												<div class="space-y-1">
													<span class="text-[10px] font-black text-white uppercase tracking-widest block">Enable Data Ghosting</span>
													<span class="text-[8px] text-text-dim uppercase tracking-tight italic">Reveal hidden telemetry on card hover</span>
												</div>
												<label class="relative inline-flex items-center cursor-pointer">
													<input type="checkbox" 
														checked={$backgroundConfig.card_hover_effect} 
														onchange={e => {
															const val = e.currentTarget.checked;
															backgroundConfig.update(b => ({ ...b, card_hover_effect: val }));
														}} 
														class="sr-only peer">
													<div class="w-12 h-6 bg-stone-950 border border-stone-800 peer-focus:outline-none rounded-none peer-checked:after:translate-x-6 peer-checked:after:bg-rust after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-stone-700 after:rounded-none after:h-4 after:w-4 after:transition-all peer-checked:bg-rust/10 peer-checked:border-rust"></div>
												</label>
											</div>

											<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4 mt-12">Telemetry Calibration</h4>
											{#each [
												{ key: 'card_hover_intensity', label: 'Effect Opacity', min: 0, max: 1, step: 0.05, unit: 'α' },
												{ key: 'card_hover_data_speed', label: 'Scroll Velocity', min: 0.1, max: 5, step: 0.1, unit: 'x' },
												{ key: 'card_hover_spark_density', label: 'Spark Concentration', min: 0, max: 30, step: 1, unit: 'qty' }
											] as slider}
												<div class="space-y-4">
													<div class="flex justify-between items-center">
														<span class="text-[10px] font-black text-stone-400 uppercase tracking-widest">{slider.label}</span>
														<span class="text-[10px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[slider.key])}{slider.unit}</span>
													</div>
													<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
														<input 
															type="range" 
															min={slider.min} 
															max={slider.max} 
															step={slider.step} 
															value={($siteSettings.aesthetic as any)[slider.key]} 
															oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [slider.key]: parseFloat(e.currentTarget.value) } }))} 
															class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
														/>
														<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {((($siteSettings.aesthetic as any)[slider.key] - slider.min) / (slider.max - slider.min)) * 100}%"></div>
													</div>
												</div>
											{/each}
										</div>

										<div class="space-y-8 bg-stone-950/30 p-12 border border-stone-800 industrial-frame flex flex-col items-center justify-center min-h-[300px]">
											<div class="modern-industrial-card glass-panel group p-10 w-full max-w-sm text-center relative cursor-help">
												<div class="corner-tl"></div><div class="corner-tr"></div><div class="corner-bl"></div><div class="corner-br"></div>
												<CardHoverOverlay />
												<div class="relative z-20 space-y-4">
													<div class="w-12 h-12 bg-rust/10 border border-rust/30 mx-auto flex items-center justify-center">
														<Icon name="ph:cursor-click-bold" size="1.5rem" class="text-rust" />
													</div>
													<span class="text-xs font-black uppercase tracking-widest block text-white">Hover to Preview Effect</span>
												</div>
											</div>
										</div>
									</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>

				<div class="p-6 bg-stone-950 border-t border-stone-800 flex items-center justify-between">
					<div class="flex items-center gap-6">
						<div class="flex items-center gap-2">
							<div class="w-1 h-1 bg-rust animate-ping"></div>
							<span class="text-[8px] text-stone-600 uppercase tracking-[0.4em]">STABLE CONNECTION</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="w-1 h-1 bg-emerald-500"></div>
							<span class="text-[8px] text-stone-600 uppercase tracking-[0.4em]">SYNC READY</span>
						</div>
					</div>
					<button
						onclick={() => notifications.add({ type: 'info', message: 'THEME SETTINGS UPDATED' })}
						class="px-10 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-[0.3em] shadow-xl shadow-rust/20 transition-all active:translate-y-px flex items-center gap-3"
					>
						<Icon name="ph:check-bold" />
						Save Theme Settings
					</button>
				</div>
			</div>
		</div>

		<!-- Right: Live Signal Preview & Presets -->
		<div class="2xl:col-span-4 sticky top-10 space-y-8">
			
			<!-- Presets Rack -->
			<div class="modern-industrial-card glass-panel !rounded-none border-stone-800">
				<div class="p-6 border-b border-stone-800 bg-stone-950/40 flex items-center gap-4">
					<Radio class="w-5 h-5 text-rust" />
					<h3 class="font-heading font-black text-sm text-white uppercase tracking-widest">Calibration_Presets</h3>
				</div>
				<div class="p-6 space-y-4">
					{#each presets as p}
						<button
							onclick={() => applyPreset(p)}
							class="w-full p-4 flex items-center justify-between border transition-all group { $siteSettings.aesthetic.theme_preset === p.id ? 'bg-rust/10 border-rust' : 'bg-stone-950/50 border-stone-800 hover:border-stone-700' }"
						>
							<div class="flex items-center gap-4">
								<div class="w-10 h-10 border border-stone-800 p-1 flex items-center justify-center bg-black">
									<div class="w-full h-full shadow-inner" style="background-color: {p.accent}"></div>
								</div>
								<div class="text-left">
									<span class="text-[10px] font-black uppercase tracking-[0.2em] block { $siteSettings.aesthetic.theme_preset === p.id ? 'text-white' : 'text-stone-500' }">{p.name}</span>
									<span class="text-[8px] text-stone-700 uppercase tracking-tight">{p.desc}</span>
								</div>
							</div>
							<ChevronRight class="w-4 h-4 { $siteSettings.aesthetic.theme_preset === p.id ? 'text-white translate-x-1' : 'text-stone-800' } transition-all" />
						</button>
					{/each}
				</div>
			</div>

			<!-- Real-time Signal Preview -->
			<div class="space-y-4">
				<div class="flex items-center gap-4 px-2">
					<div class="h-0.5 w-10 bg-rust"></div>
					<span class="font-jetbrains text-[10px] font-black text-rust uppercase tracking-[0.4em]">Live_Signal_Preview</span>
				</div>

				<div class="bg-[var(--bg-color)] border border-stone-800 industrial-frame overflow-hidden flex flex-col h-[500px] shadow-[0_0_100px_rgba(0,0,0,0.5)] relative">
					<!-- Mock UI Layout -->
					<div class="bg-black/80 border-b border-stone-800 p-4 flex justify-between items-center z-20">
						<div class="flex items-center gap-3">
							<div class="w-2 h-2 bg-rust shadow-[0_0_5px_var(--color-rust)]"></div>
							<span class="text-xs font-black military-label text-white uppercase tracking-tighter">PREVIEW_OS</span>
						</div>
						<div class="flex gap-2">
							<div class="w-2 h-2 rounded-full bg-stone-800"></div>
							<div class="w-2 h-2 rounded-full bg-stone-800"></div>
						</div>
					</div>

					<div class="flex-1 p-6 space-y-6 overflow-hidden bg-transparent z-10 custom-scrollbar">
						<div class="grid grid-cols-2 gap-4">
							<StatsCard title="CPU_LOAD" value="42.8%" iconName="ph:cpu-bold" color="rust" />
							<StatsCard title="DB_UPLINK" value="STABLE" iconName="ph:check-circle-bold" color="emerald" />
						</div>

						<div class="modern-industrial-card !bg-black/40 p-5 space-y-4">
							<div class="flex items-center gap-3">
								<Icon name="ph:terminal-window-bold" size="1rem" class="text-rust" />
								<span class="text-[9px] font-black text-white uppercase tracking-widest">Active_Processes</span>
							</div>
							<div class="space-y-2">
								{#each [1,2] as i}
									<div class="h-8 bg-stone-950/50 border border-stone-800 flex items-center px-4 justify-between transition-transform hover:scale-[var(--hover-scale)]">
										<span class="text-[8px] uppercase font-mono" style="color: var(--text-dim)">Kernel_Module_0{i}</span>
										<span class="text-[8px] text-success font-mono font-bold">RUNNING</span>
									</div>
								{/each}
							</div>
						</div>

						<button class="w-full py-3 bg-rust text-white font-black text-[10px] uppercase active:translate-y-[var(--press-depth)] transition-all">Test_Action_Trigger</button>
					</div>

					<!-- Post-processing Simulation Layer -->
					<div class="absolute inset-0 z-0 pointer-events-none opacity-20">
						<div class="w-full h-full bg-[radial-gradient(circle_at_center,var(--primary-color)_0%,transparent_70%)] opacity-10"></div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.custom-scrollbar::-webkit-scrollbar {
		width: 4px;
	}
	.custom-scrollbar::-webkit-scrollbar-track {
		background: transparent;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb {
		background: #1a1a1a;
	}
	.custom-scrollbar::-webkit-scrollbar-thumb:hover {
		background: var(--color-rust);
	}

	input[type="range"] {
		transition: all 0.3s;
		cursor: pointer;
	}
	input[type="range"]:hover {
		filter: brightness(1.1);
	}
</style>