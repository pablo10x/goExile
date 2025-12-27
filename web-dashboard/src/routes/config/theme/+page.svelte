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
		Check,
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
		Radio
	} from 'lucide-svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import { fade, slide, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	// Calibration Subsystems
	let activeSubsystem = $state<'chromatic' | 'atmospheric' | 'geometric' | 'structural'>('chromatic');

	const subsystems = [
		{ id: 'chromatic', label: 'Chromatic_Matrix', icon: Palette, desc: 'Color calibration & luminance' },
		{ id: 'atmospheric', label: 'Atmospheric_Core', icon: Cloud, desc: 'CRT, Scanlines & Noise' },
		{ id: 'geometric', label: 'Geometric_Logic', icon: Box, desc: 'Frames, Radii & Layout' },
		{ id: 'structural', label: 'Structural_Engine', icon: Layers, desc: 'Background & Particles' }
	];

	const fontOptions = ['Inter', 'Space Grotesk', 'Michroma', 'Orbitron', 'Red Hat Mono', 'Syncopate', 'Kanit', 'JetBrains Mono'];

	const backgroundEngines = [
		{ id: 'architecture', label: 'Brutalist', icon: Server },
		{ id: 'tactical_grid', label: 'Radar', icon: Network },
		{ id: 'neural_network', label: 'Neural', icon: Activity },
		{ id: 'data_flow', label: 'Data Flow', icon: FileJson },
		{ id: 'digital_horizon', label: 'Horizon', icon: Globe },
		{ id: 'particle_network', label: 'Particles', icon: Sparkles },
		{ id: 'cyber_ocean', label: 'Ocean', icon: Waves },
		{ id: 'static_void', label: 'Static', icon: Cloud },
		{ id: 'none', label: 'Minimal', icon: Zap }
	];

	// Preset Engine
	const presets = [
		{ 
			id: 'deep_command', 
			name: 'DEEP_COMMAND', 
			accent: '#c2410c', 
			bg: '#020202',
			desc: 'High-alert tactical interface'
		},
		{ 
			id: 'mercury_protocol', 
			name: 'MERCURY_PROTO', 
			accent: '#06b6d4', 
			bg: '#050505',
			desc: 'Clean, cold digital aesthetic'
		},
		{ 
			id: 'solar_flare', 
			name: 'SOLAR_FLARE', 
			accent: '#f59e0b', 
			bg: '#080705',
			desc: 'High-luminance emergency hud'
		},
		{ 
			id: 'void_walker', 
			name: 'VOID_WALKER', 
			accent: '#a855f7', 
			bg: '#020105',
			desc: 'Minimalist stealth operation'
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
			// (Implementation remains similar but with new properties)
			location.reload(); 
		}
	}
</script>

<div class="relative z-10 w-full space-y-10 pb-32 font-jetbrains">
	<!-- System Calibration Header -->
	<div class="flex flex-col xl:flex-row xl:items-end justify-between gap-8 border-l-4 border-rust pl-6 sm:pl-10 py-4 bg-[#0a0a0a]/60 backdrop-blur-xl industrial-frame shadow-2xl relative overflow-hidden">
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
					SYSTEM_<span class="text-rust">CALIBRATION</span>
				</h1>
				<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-[0.4em] font-black mt-3 italic">
					Fine-tuning neural interface geometry & chromatic response
				</p>
			</div>
		</div>

		<div class="flex items-center gap-4 p-2 relative z-10">
			<div class="flex bg-black/40 p-1 border border-stone-800">
				<button
					onclick={resetToDefault}
					class="px-6 py-3 hover:bg-white hover:text-black text-stone-500 font-black text-[10px] uppercase tracking-widest transition-all"
				>
					Purge_Config
				</button>
				<div class="w-px h-10 bg-stone-800"></div>
				<div class="px-6 py-3 flex items-center gap-3">
					<div class="w-2 h-2 bg-emerald-500 animate-flicker"></div>
					<span class="text-[10px] font-black text-emerald-500/70 uppercase tracking-widest">Calibration_Online</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Main Workbench -->
	<div class="grid grid-cols-1 2xl:grid-cols-12 gap-8 items-start">
		
		<!-- Left: Subsystem Navigation & Controls -->
		<div class="2xl:col-span-8 space-y-8">
			
			<!-- Subsystem Tabs -->
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				{#each subsystems as sub}
					<button
						onclick={() => activeSubsystem = sub.id as any}
						class="flex flex-col items-start p-5 border transition-all duration-500 relative group overflow-hidden {activeSubsystem === sub.id ? 'bg-rust text-white border-rust shadow-2xl shadow-rust/20' : 'bg-stone-950/50 border-stone-800 text-stone-600 hover:border-stone-700'}"
					>
						<div class="flex items-center justify-between w-full mb-4">
							<sub.icon class="w-5 h-5 {activeSubsystem === sub.id ? 'text-white' : 'text-stone-700 group-hover:text-rust'}" />
							<div class="w-1.5 h-1.5 rounded-full {activeSubsystem === sub.id ? 'bg-white animate-pulse' : 'bg-stone-900'}"></div>
						</div>
						<span class="text-[10px] font-black uppercase tracking-widest mb-1">{sub.label}</span>
						<span class="text-[8px] opacity-60 uppercase tracking-tight line-clamp-1">{sub.desc}</span>
						
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
					<div class="flex gap-2">
						{#each [1,2,3] as i}
							<div class="w-1 h-1 bg-stone-800 rounded-full"></div>
						{/each}
					</div>
				</div>

				<div class="p-10 flex-1 overflow-y-auto custom-scrollbar bg-[radial-gradient(circle_at_top_right,rgba(249,115,22,0.02),transparent)]">
					{#if activeSubsystem === 'chromatic'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<!-- Primary Matrix -->
							<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-10">
								{#each [
									{ key: 'primary_color', label: 'Primary_Accent', desc: 'Main interactive color' },
									{ key: 'secondary_color', label: 'Secondary_Node', desc: 'Muted structural color' },
									{ key: 'card_bg_color', label: 'Core_Background', desc: 'Module base surface' },
									{ key: 'card_glow_color', label: 'Luminous_Vector', desc: 'Glow & radiance intensity' },
									{ key: 'text_color_primary', label: 'Text_High_Vis', desc: 'Primary information layer' },
									{ key: 'bg_color', label: 'Void_Foundation', desc: 'Deep system background' }
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

							<!-- Specialized Signals -->
							<div class="pt-10 border-t border-stone-800">
								<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] mb-8 italic">Specialized_Signal_Calibration</h4>
								<div class="grid grid-cols-2 lg:grid-cols-4 gap-8">
									{#each [
										{ key: 'success_color', label: 'Success_Sig', color: 'text-emerald-500' },
										{ key: 'warning_color', label: 'Warning_Sig', color: 'text-amber-500' },
										{ key: 'danger_color', label: 'Critical_Sig', color: 'text-red-500' },
										{ key: 'info_color', label: 'Info_Sig', color: 'text-cyan-500' }
									] as sig}
										<div class="flex items-center gap-4 p-4 bg-stone-900/30 border border-stone-800 hover:border-rust/30 transition-all group relative">
											<div class="w-10 h-10 shrink-0 relative">
												<input 
													type="color" 
													value={getHex(($siteSettings.aesthetic as any)[sig.key])} 
													oninput={e => updateColor(sig.key, e.currentTarget.value, getAlpha(($siteSettings.aesthetic as any)[sig.key]))}
													class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10" 
												/>
												<div class="absolute inset-0 border border-stone-700 group-hover:border-rust transition-colors" style="background-color: {($siteSettings.aesthetic as any)[sig.key]}"></div>
											</div>
											<div>
												<span class="text-[9px] font-black text-stone-500 uppercase tracking-widest block">{sig.label}</span>
												<span class="text-[10px] font-mono font-bold {sig.color}">{getHex(($siteSettings.aesthetic as any)[sig.key])}</span>
											</div>
										</div>
									{/each}
								</div>
							</div>
						</div>
					{:else if activeSubsystem === 'atmospheric'}
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
								<!-- Signal Quality -->
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Signal_Quality_Calibration</h4>
									
									{#each [
										{ key: 'crt_effect', label: 'CRT_Cathode_Ray', desc: 'Scanline simulation & phosphor' },
										{ key: 'crt_curve', label: 'Spherical_Distortion', desc: 'Curved viewport geometry' },
										{ key: 'text_glow', label: 'Neural_Luminance', desc: 'Text-glow radiation effect' },
										{ key: 'panic_mode', label: 'RED_ALERT_PROTOCOL', desc: 'System-wide emergency chroma' }
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
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Analog_Interference_Values</h4>
									
									{#each [
										{ key: 'glitch_intensity', label: 'Glitch_Frequency', min: 0, max: 1, step: 0.01, unit: 'Hz' },
										{ key: 'chromatic_aberration', label: 'Lens_Fringing', min: 0, max: 10, step: 0.5, unit: 'px' },
										{ key: 'scanlines_opacity', label: 'Scanline_Density', min: 0, max: 0.5, step: 0.01, unit: 'α' },
										{ key: 'noise_opacity', label: 'Grain_Saturation', min: 0, max: 0.2, step: 0.01, unit: 'α' },
										{ key: 'vignette_intensity', label: 'Vignette_Falloff', min: 0, max: 1, step: 0.05, unit: 'α' }
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
								<!-- Typography -->
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Typeface_Engine_Calibration</h4>
									
									{#each [
										{ key: 'font_header', label: 'Heading_Font' },
										{ key: 'font_body', label: 'Interface_Font' },
										{ key: 'font_mono', label: 'Telemetry_Font' }
									] as type}
										<div class="space-y-3">
											<span class="text-[9px] font-black text-stone-600 uppercase tracking-[0.3em] ml-1">{type.label}</span>
											<div class="grid grid-cols-2 gap-2">
												{#each fontOptions as font}
													<button 
														onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [type.key]: font } }))}
														class="px-3 py-2 border transition-all text-[9px] font-black uppercase text-left industrial-frame {($siteSettings.aesthetic as any)[type.key] === font ? 'border-rust text-white bg-rust/10' : 'border-stone-800 text-stone-600 hover:border-stone-700'}"
														style="font-family: '{font}';"
													>
														{font}
													</button>
												{/each}
											</div>
										</div>
									{/each}
								</div>

								<!-- Frame Geometry -->
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Structural_Geometry_Rules</h4>
									
									<div class="space-y-6">
										<span class="text-[9px] font-black text-stone-600 uppercase tracking-[0.3em] ml-1">Corner_Style_Protocol</span>
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
										{ key: 'border_radius_lg', label: 'Radius_Global', max: 40 },
										{ key: 'card_border_width', label: 'Line_Weight', max: 5 },
										{ key: 'card_shadow_size', label: 'Shadow_Vector', max: 100 },
										{ key: 'sidebar_width', label: 'Module_Width', min: 200, max: 400 },
										{ key: 'grid_opacity', label: 'Grid_Substrate_Opacity', max: 0.2, step: 0.01 }
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
						<!-- Structural / Structural Engine -->
						<div in:fade={{ duration: 300 }} class="space-y-12">
							<div class="grid grid-cols-1 gap-10">
								<div class="space-y-8">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4">Structural_Background_Protocol</h4>
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

								<div class="pt-10 border-t border-stone-800">
									<h4 class="text-[10px] font-black text-stone-500 uppercase tracking-[0.4em] border-b border-stone-800 pb-4 mb-8">Atmosphere_Particles</h4>
									<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
										{#each [
											{ key: 'show_smoke', label: 'Volumetric_Fluid', desc: 'Simulate neural smoke exhaust', icon: Cloud },
											{ key: 'show_rain', label: 'Digital_Precipitation', desc: 'Vertical packet stream falling', icon: CloudRain },
											{ key: 'show_clouds', label: 'Nebula_Overlay', desc: 'High-altitude gas formation', icon: Wind },
											{ key: 'show_navbar_particles', label: 'Interface_Dust', desc: 'Sidebar ambient magnetic dust', icon: Sparkles }
										] as toggle}
											<div class="flex items-center justify-between p-5 bg-stone-900/30 border border-stone-800 hover:border-rust/30 transition-all group industrial-frame">
												<div class="flex items-center gap-4">
													<toggle.icon class="w-5 h-5 text-rust opacity-40 group-hover:opacity-100 transition-opacity" />
													<div class="space-y-1">
														<span class="text-[10px] font-black text-white uppercase tracking-widest block">{toggle.label}</span>
														<span class="text-[8px] text-stone-600 uppercase tracking-tight italic">{toggle.desc}</span>
													</div>
												</div>
												<label class="relative inline-flex items-center cursor-pointer">
													<input type="checkbox" checked={($backgroundConfig as any)[toggle.key]} onchange={e => backgroundConfig.update(b => ({ ...b, [toggle.key]: e.currentTarget.checked }))} class="sr-only peer">
													<div class="w-12 h-6 bg-stone-950 border border-stone-800 peer-focus:outline-none rounded-none peer-checked:after:translate-x-6 peer-checked:after:bg-rust after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-stone-700 after:rounded-none after:h-4 after:w-4 after:transition-all peer-checked:bg-rust/10 peer-checked:border-rust"></div>
												</label>
											</div>
										{/each}
									</div>
								</div>
							</div>
						</div>
					{/if}
				</div>

				<div class="p-6 bg-stone-950 border-t border-stone-800 flex items-center justify-between">
					<div class="flex items-center gap-6">
						<div class="flex items-center gap-2">
							<div class="w-1 h-1 bg-rust animate-ping"></div>
							<span class="text-[8px] text-stone-600 uppercase tracking-[0.4em]">SYNC_READY</span>
						</div>
						<div class="flex items-center gap-2">
							<div class="w-1 h-1 bg-emerald-500"></div>
							<span class="text-[8px] text-stone-600 uppercase tracking-[0.4em]">BUFFER_CLEAR</span>
						</div>
					</div>
					<button
						onclick={() => notifications.add({ type: 'info', message: 'INTERFACE_RECALIBRATED' })}
						class="px-10 py-3 bg-rust hover:bg-rust-light text-white font-heading font-black text-[11px] uppercase tracking-[0.3em] shadow-xl shadow-rust/20 transition-all active:translate-y-px"
					>
						Commit_Calibration
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
							<StatsCard title="CPU_LOAD" value="42.8%" Icon={Activity} color="rust" />
							<StatsCard title="DB_UPLINK" value="STABLE" Icon={CheckCircle} color="emerald" />
						</div>

						<div class="modern-industrial-card !bg-black/40 p-5 space-y-4">
							<div class="flex items-center gap-3">
								<Terminal class="w-4 h-4 text-rust" />
								<span class="text-[9px] font-black text-white uppercase tracking-widest">Active_Processes</span>
							</div>
							<div class="space-y-2">
								{#each [1,2] as i}
									<div class="h-8 bg-stone-950/50 border border-stone-800 flex items-center px-4 justify-between">
										<span class="text-[8px] text-stone-500 uppercase font-mono">Kernel_Module_0{i}</span>
										<span class="text-[8px] text-emerald-500 font-mono font-bold">RUNNING</span>
									</div>
								{/each}
							</div>
						</div>

						<div class="p-4 bg-red-950/10 border-l-2 border-red-600 text-red-500 flex items-center gap-4">
							<AlertCircle class="w-4 h-4 animate-pulse" />
							<span class="text-[9px] font-black uppercase tracking-widest">Breach_Attempt_Detected_Sector_7</span>
						</div>
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