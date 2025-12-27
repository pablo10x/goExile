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
		Sparkles
	} from 'lucide-svelte';
	import StatsCard from '$lib/components/StatsCard.svelte';
	import { fade, slide, scale } from 'svelte/transition';

	// Component Preview States
	let previewSearchQuery = $state('');
	let isExpanded = $state(true);
	let isGeometryExpanded = $state(false);
	let isAtmosphereExpanded = $state(false);

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
			siteSettings.set({
				site_name: 'EXILE',
				version_tag: 'v0.9.4-PROTOTYPE',
				aesthetic: {
					crt_effect: true,
					scanlines_opacity: 0.05,
					noise_opacity: 0.03,
					industrial_styling: true,
					glassmorphism: true,
					glow_effects: true,
					animations_enabled: true,
					topology_blobs: true,
					card_alpha: 0.6,
					backdrop_blur: 20,
					accent_color: '#d97706',
					sidebar_alpha: 0.7,
					bg_opacity: 1.0,
					bg_color: '#020617',
					primary_color: '#d97706',
					secondary_color: '#1e293b',
					card_bg_color: '#0f172a',
					hover_color: '#1e293b',
					text_color_primary: '#e2e2e2',
					text_color_secondary: '#888888',
					border_color: '#1e293b',
					font_primary: 'Inter',
					font_header: 'Kanit',
					font_body: 'Inter',
					font_mono: 'JetBrains Mono',
					border_radius_sm: 2,
					border_radius_md: 4,
					border_radius_lg: 8,
					glass_strength: 0.1,
					modal_animation: 'scale',
					card_border_width: 1,
					card_shadow_size: 4,
					scanline_speed: 4,
					scanline_density: 3,
					noise_intensity: 0.03,
					terminal_line_height: 1.5,
					panic_mode: false,
					industrial_border_color: '#1e293b',
					crt_curve: true,
					vignette_intensity: 0.5,
					border_glow_intensity: 0.4,
					text_glow: true,
					reduced_motion: false,
					mobile_optimized: true,
					sidebar_width: 280,
					card_glow_color: '#d97706',
					font_size_base: 14,
					theme_preset: 'deep_command'
				},
				performance: {
					high_quality_smoke: false,
					particle_density: 0.5,
					low_power_mode: false,
					disable_expensive_animations: false
				},
				site_notice: {
					enabled: false,
					message: 'SYSTEM MAINTENANCE SCHEDULED FOR 0200 HOURS',
					type: 'info'
				}
			});
			notifications.add({ type: 'info', message: 'SETTINGS_RESET_SUCCESSFUL' });
		}
	}
</script>

<div class="relative z-10 w-full space-y-10 pb-32">
	<!-- Header -->
	<div class="flex flex-col xl:flex-row xl:items-center justify-between gap-8 mb-10 border-l-4 border-rust pl-6 sm:pl-10 py-2 bg-[#0a0a0a]/60 backdrop-blur-xl industrial-frame">
		<div class="flex items-center gap-6">
			<div class="p-4 bg-rust/10 border border-rust/30 shadow-2xl industrial-frame">
				<Palette class="w-10 h-10 text-rust-light" />
			</div>
			<div>
				<h1 class="text-4xl sm:text-5xl font-heading font-black text-white uppercase tracking-tighter leading-none">
					THEME_<span class="text-rust">LAB</span>
				</h1>
				<p class="font-jetbrains text-[10px] text-stone-500 uppercase tracking-[0.3em] font-black mt-2">
					Interface Geometry & Chromatic Calibration
				</p>
			</div>
		</div>
		<div class="flex items-center gap-4">
			<button
				onclick={resetToDefault}
				class="px-8 py-3 bg-stone-900 hover:bg-stone-800 text-stone-400 font-heading font-black text-[11px] uppercase tracking-widest transition-all border border-stone-800"
			>
				Reset Defaults
			</button>
			<div class="h-10 w-px bg-stone-800 hidden sm:block"></div>
			<div class="flex items-center gap-3 font-jetbrains text-[10px] font-black text-stone-600 uppercase tracking-widest">
				<div class="w-2 h-2 bg-emerald-500 shadow-[0_0_10px_#10b981] animate-pulse"></div>
				ENGINE_ACTIVE
			</div>
		</div>
	</div>

	<div class="grid grid-cols-1 2xl:grid-cols-12 gap-10 items-start">
		<!-- Left: Controls -->
		<div class="2xl:col-span-7 space-y-10">
			<!-- Color Matrix -->
			<div class="modern-industrial-card glass-panel !rounded-none">
				<div class="p-6 border-b border-stone-800 bg-stone-950/40 flex items-center gap-4">
					<Activity class="w-5 h-5 text-rust" />
					<h3 class="font-heading font-black text-sm text-white uppercase tracking-widest">Chromatic Matrix</h3>
				</div>
				<div class="p-8">
					<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
						{#each [
							{ key: 'primary_color', label: 'Primary (Accent)' },
							{ key: 'secondary_color', label: 'Secondary' },
							{ key: 'card_bg_color', label: 'Card Base' },
							{ key: 'hover_color', label: 'Hover State' },
							{ key: 'border_color', label: 'Border Base' },
							{ key: 'card_glow_color', label: 'Glow Vector' },
							{ key: 'text_color_primary', label: 'Primary Text' },
							{ key: 'text_color_secondary', label: 'Secondary Text' },
							{ key: 'bg_color', label: 'Void Color' }
						] as color}
							{@const currentVal = ($siteSettings.aesthetic as any)[color.key]}
							<div class="space-y-4">
								<div class="flex justify-between items-center">
									<h4 class="font-black text-stone-500 uppercase text-[9px] tracking-widest">{color.label}</h4>
									<span class="text-[9px] font-mono text-stone-600 uppercase">{currentVal}</span>
								</div>
								
								<div class="flex flex-col gap-3">
									<div class="relative h-10 group">
										<input 
											type="color" 
											value={getHex(currentVal)} 
											oninput={e => updateColor(color.key, e.currentTarget.value, getAlpha(currentVal))} 
											class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10" 
										/>
										<div class="absolute inset-0 border border-stone-800 group-hover:border-rust transition-colors flex items-center px-3 gap-3 bg-stone-950/50 industrial-frame overflow-hidden">
											<div class="w-5 h-5 border border-white/10 shadow-lg relative shrink-0">
												<!-- Checkerboard background for alpha preview -->
												<div class="absolute inset-0 bg-[url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAYAAACp8Z5+AAAAAXNSR0IArs4c6QAAACBJREFUGFdjZEADJEYGhj9MDIyMDAwMDAwMDAwMDAwMDAwMABYDAwMAAAAASUVORK5CYII=')] opacity-20"></div>
												<div class="absolute inset-0" style="background-color: {currentVal}"></div>
											</div>
											<span class="text-[9px] font-mono text-stone-700 truncate">HEX: {getHex(currentVal)}</span>
										</div>
									</div>

									<div class="space-y-2">
										<div class="flex justify-between items-center">
											<span class="text-[8px] font-mono text-stone-700 uppercase">Alpha_Opacity</span>
											<span class="text-[8px] font-mono text-rust-light">{(getAlpha(currentVal) * 100).toFixed(0)}%</span>
										</div>
										<div class="relative flex items-center h-1.5 bg-stone-950 border border-stone-800">
											<input 
												type="range" 
												min="0" 
												max="1" 
												step="0.01" 
												value={getAlpha(currentVal)} 
												oninput={e => updateColor(color.key, getHex(currentVal), parseFloat(e.currentTarget.value))} 
												class="w-full h-full appearance-none cursor-pointer bg-transparent accent-rust z-10" 
											/>
											<div class="absolute top-0 left-0 h-full bg-rust/30" style="width: {getAlpha(currentVal) * 100}%"></div>
										</div>
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Geometry & Effects -->
			<div class="space-y-6">
				<!-- System Geometry -->
				<div class="modern-industrial-card glass-panel !rounded-none">
					<button 
						onclick={() => isGeometryExpanded = !isGeometryExpanded}
						class="w-full p-6 border-b border-stone-800 bg-stone-950/40 flex items-center justify-between hover:bg-stone-900/40 transition-colors"
					>
						<div class="flex items-center gap-4">
							<Monitor class="w-5 h-5 text-rust" />
							<h3 class="font-heading font-black text-sm text-white uppercase tracking-widest">System Geometry</h3>
						</div>
						<ChevronDown class="w-5 h-5 text-stone-600 transition-transform duration-300 {isGeometryExpanded ? 'rotate-180 text-rust' : ''}" />
					</button>
					
					{#if isGeometryExpanded}
						<div class="p-8" transition:slide>
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
								{#each [
									{ key: 'crt_effect', label: 'CRT Overlay', desc: 'Cathode-ray simulation', icon: Monitor },
									{ key: 'crt_curve', label: 'Lens Distortion', desc: 'Curved viewport logic', icon: Monitor },
									{ key: 'industrial_styling', label: 'Angular Frames', desc: 'Brutalist clip-paths', icon: Shield },
									{ key: 'text_glow', label: 'Neural Text Glow', desc: 'Heading radiance', icon: Zap },
									{ key: 'reduced_motion', label: 'Reduced Motion', desc: 'Disable heavy loops', icon: Activity }
								] as toggle}
									<div class="flex items-center justify-between p-4 bg-stone-950/30 border border-stone-800 hover:border-rust/30 transition-all industrial-frame">
										<div class="flex items-center gap-3">
											<toggle.icon class="w-4 h-4 text-rust" />
											<div>
												<h4 class="font-black text-stone-300 uppercase text-[10px] tracking-widest">{toggle.label}</h4>
												<p class="text-[8px] text-stone-600 font-jetbrains uppercase">{toggle.desc}</p>
											</div>
										</div>
										<label class="relative inline-flex items-center cursor-pointer">
											<input type="checkbox" checked={($siteSettings.aesthetic as any)[toggle.key]} onchange={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [toggle.key]: e.currentTarget.checked } }))} class="sr-only peer">
											<div class="w-10 h-5 bg-stone-900 border border-stone-800 peer-focus:outline-none rounded-none peer-checked:after:translate-x-5 peer-checked:after:bg-rust after:content-[''] after:absolute after:top-[3px] after:left-[3px] after:bg-stone-700 after:rounded-none after:h-3.5 after:w-3.5 after:transition-all peer-checked:bg-rust/10 peer-checked:border-rust"></div>
										</label>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>

				<!-- Atmosphere -->
				<div class="modern-industrial-card glass-panel !rounded-none">
					<button 
						onclick={() => isAtmosphereExpanded = !isAtmosphereExpanded}
						class="w-full p-6 border-b border-stone-800 bg-stone-950/40 flex items-center justify-between hover:bg-stone-900/40 transition-colors"
					>
						<div class="flex items-center gap-4">
							<Zap class="w-5 h-5 text-rust" />
							<h3 class="font-heading font-black text-sm text-white uppercase tracking-widest">Atmosphere</h3>
						</div>
						<ChevronDown class="w-5 h-5 text-stone-600 transition-transform duration-300 {isAtmosphereExpanded ? 'rotate-180 text-rust' : ''}" />
					</button>

					{#if isAtmosphereExpanded}
						<div class="p-8 space-y-10" transition:slide>
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4 pb-8 border-b border-stone-800">
								{#each [
									{ key: 'show_smoke', label: 'Volumetric Smoke', desc: 'Interactive fluid simulation', icon: Cloud },
									{ key: 'show_rain', label: 'Digital Rain', desc: 'Vertical data streams', icon: CloudRain },
									{ key: 'show_clouds', label: 'Nebula Layer', desc: 'Background parallax clouds', icon: Wind },
									{ key: 'show_navbar_particles', label: 'Nav Particles', desc: 'Sidebar ambient dust', icon: Sparkles }
								] as toggle}
									<div class="flex items-center justify-between p-4 bg-stone-950/30 border border-stone-800 hover:border-rust/30 transition-all industrial-frame">
										<div class="flex items-center gap-3">
											<toggle.icon class="w-4 h-4 text-rust" />
											<div>
												<h4 class="font-black text-stone-300 uppercase text-[10px] tracking-widest">{toggle.label}</h4>
												<p class="text-[8px] text-stone-600 font-jetbrains uppercase">{toggle.desc}</p>
											</div>
										</div>
										<label class="relative inline-flex items-center cursor-pointer">
											<input type="checkbox" checked={($backgroundConfig as any)[toggle.key]} onchange={e => backgroundConfig.update(b => ({ ...b, [toggle.key]: e.currentTarget.checked }))} class="sr-only peer">
											<div class="w-10 h-5 bg-stone-900 border border-stone-800 peer-focus:outline-none rounded-none peer-checked:after:translate-x-5 peer-checked:after:bg-rust after:content-[''] after:absolute after:top-[3px] after:left-[3px] after:bg-stone-700 after:rounded-none after:h-3.5 after:w-3.5 after:transition-all peer-checked:bg-rust/10 peer-checked:border-rust"></div>
										</label>
									</div>
								{/each}
							</div>

							<div class="grid grid-cols-1 md:grid-cols-2 gap-8">
								{#each [
									{ key: 'scanline_speed', label: 'Scanline Velocity', icon: RefreshCw, max: 20, min: 0.5, step: 0.5, unit: 's' },
									{ key: 'scanline_density', label: 'Scanline Pitch', icon: Menu, max: 20, min: 1, step: 1, unit: 'px' },
									{ key: 'vignette_intensity', label: 'Vignette Depth', icon: Monitor, max: 1, min: 0, step: 0.05, unit: '' },
									{ key: 'backdrop_blur', label: 'Glass Diffusion', icon: Cloud, max: 40, min: 0, step: 1, unit: 'px' },
									{ key: 'sidebar_width', label: 'Module Width', icon: Menu, max: 400, min: 200, step: 8, unit: 'px' }
								] as slider}
									<div class="space-y-4">
										<div class="flex justify-between items-center">
											<div class="flex items-center gap-3">
												<slider.icon class="w-3.5 h-3.5 text-rust" />
												<h4 class="font-black text-stone-400 uppercase text-[10px] tracking-widest">{slider.label}</h4>
											</div>
											<span class="text-[10px] font-mono text-rust-light">
												{(($siteSettings.aesthetic as any)[slider.key])}{slider.unit}
											</span>
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
					{/if}
				</div>
			</div>

			<!-- Visual Core Engine -->
			<div class="modern-industrial-card glass-panel !rounded-none">
				<div class="p-6 border-b border-stone-800 bg-stone-950/40 flex items-center gap-4">
					<Activity class="w-5 h-5 text-rust" />
					<h3 class="font-heading font-black text-sm text-white uppercase tracking-widest">Visual Core Engine</h3>
				</div>
				<div class="p-8 space-y-6">
					<div class="grid grid-cols-1 gap-4">
						{#each backgroundEngines as engine}
							{@const isSelected = $backgroundConfig.global_type === engine.id}
							<div class="bg-stone-950/30 border transition-all industrial-frame overflow-hidden {isSelected ? 'border-rust shadow-lg shadow-rust/10' : 'border-stone-800 hover:border-stone-700'}">
								<button 
									onclick={() => backgroundConfig.update(b => ({ ...b, global_type: engine.id as any }))}
									class="w-full flex items-center justify-between p-4"
								>
									<div class="flex items-center gap-4">
										<div class="p-2 bg-stone-900 border border-stone-800 {isSelected ? 'text-rust-light border-rust/30' : 'text-stone-600'}">
											<engine.icon class="w-5 h-5" />
										</div>
										<span class="font-black uppercase text-[10px] tracking-widest {isSelected ? 'text-white' : 'text-stone-500'}">{engine.label}</span>
									</div>
									<div class="w-3 h-3 border border-stone-700 flex items-center justify-center {isSelected ? 'bg-rust border-rust' : 'bg-stone-900'}">
										{#if isSelected}<div class="w-1.5 h-1.5 bg-white"></div>{/if}
									</div>
								</button>

								{#if isSelected && ($backgroundConfig.settings as any)[engine.id]}
									{@const currentEngine = ($backgroundConfig.settings as any)[engine.id]}
									<div class="px-4 pb-6 pt-2 border-t border-stone-800/50 bg-black/20" transition:slide>
										<div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
											<!-- Intensity -->
											{#if currentEngine.intensity !== undefined}
												<div class="space-y-2">
													<div class="flex justify-between items-center">
														<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Intensity</span>
														<span class="text-[9px] font-mono text-rust-light">{currentEngine.intensity.toFixed(2)}</span>
													</div>
													<input 
														type="range" 
														min="0" 
														max="2" 
														step="0.05" 
														value={currentEngine.intensity} 
														oninput={e => backgroundConfig.update(b => {
															const settings = { ...b.settings };
															(settings as any)[b.global_type].intensity = parseFloat(e.currentTarget.value);
															return { ...b, settings };
														})}
														class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
													/>
												</div>
											{/if}

											<!-- Speed -->
											{#if currentEngine.speed !== undefined}
												<div class="space-y-2">
													<div class="flex justify-between items-center">
														<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Velocity</span>
														<span class="text-[9px] font-mono text-rust-light">{currentEngine.speed.toFixed(2)}x</span>
													</div>
													<input 
														type="range" 
														min="0" 
														max="5" 
														step="0.1" 
														value={currentEngine.speed} 
														oninput={e => backgroundConfig.update(b => {
															const settings = { ...b.settings };
															(settings as any)[b.global_type].speed = parseFloat(e.currentTarget.value);
															return { ...b, settings };
														})}
														class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
													/>
												</div>
											{/if}

											<!-- Density -->
											{#if currentEngine.density !== undefined}
												<div class="space-y-2">
													<div class="flex justify-between items-center">
														<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Density</span>
														<span class="text-[9px] font-mono text-rust-light">{currentEngine.density.toFixed(1)}</span>
													</div>
													<input 
														type="range" 
														min="0.1" 
														max="3" 
														step="0.1" 
														value={currentEngine.density} 
														oninput={e => backgroundConfig.update(b => {
															const settings = { ...b.settings };
															(settings as any)[b.global_type].density = parseFloat(e.currentTarget.value);
															return { ...b, settings };
														})}
														class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
													/>
												</div>
											{/if}

											<!-- Size -->
											{#if currentEngine.size !== undefined}
												<div class="space-y-2">
													<div class="flex justify-between items-center">
														<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Particle Size</span>
														<span class="text-[9px] font-mono text-rust-light">{currentEngine.size.toFixed(1)}x</span>
													</div>
													<input 
														type="range" 
														min="0.5" 
														max="3" 
														step="0.1" 
														value={currentEngine.size} 
														oninput={e => backgroundConfig.update(b => {
															const settings = { ...b.settings };
															(settings as any)[b.global_type].size = parseFloat(e.currentTarget.value);
															return { ...b, settings };
														})}
														class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
													/>
												</div>
											{/if}

											<!-- Color -->
											{#if currentEngine.color !== undefined}
												<div class="space-y-2">
													<div class="flex justify-between items-center">
														<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Base Color</span>
														<span class="text-[9px] font-mono text-stone-600 uppercase">{currentEngine.color}</span>
													</div>
													<div class="relative h-8 group">
														<input 
															type="color" 
															value={currentEngine.color} 
															oninput={e => backgroundConfig.update(b => {
																const settings = { ...b.settings };
																(settings as any)[b.global_type].color = e.currentTarget.value;
																return { ...b, settings };
															})}
															class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10" 
														/>
														<div class="absolute inset-0 border border-stone-800 group-hover:border-rust transition-colors flex items-center px-3 gap-3 bg-stone-950/50">
															<div class="w-3 h-3 border border-white/10 shadow-lg" style="background-color: {currentEngine.color}"></div>
															<span class="text-[8px] font-mono text-stone-500">HEX_VALUE</span>
														</div>
													</div>
												</div>
											{/if}
										</div>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<!-- Right: Live Preview -->
		<div class="2xl:col-span-5 sticky top-10 space-y-8">
			<div class="flex items-center gap-4 mb-2">
				<div class="h-0.5 w-8 bg-rust"></div>
				<span class="font-jetbrains text-[10px] font-black text-rust uppercase tracking-[0.3em]">Live_Signal_Preview</span>
			</div>

			<div class="bg-[var(--bg-color)] border border-stone-800/50 industrial-frame overflow-hidden flex flex-col h-[700px] shadow-2xl relative">
				<!-- Mock UI Header -->
				<div class="bg-black/80 backdrop-blur-md border-b border-stone-800 flex justify-between items-center px-4 py-2 relative z-20 h-14">
					<div class="flex items-center gap-3">
						<div class="p-2 -ml-2 text-stone-400">
							<Menu class="w-5 h-5" />
						</div>
						<div class="flex flex-col">
							<div class="flex items-center gap-2">
								<div class="w-1.5 h-1.5 bg-rust shadow-[0_0_5px_var(--color-rust)]"></div>
								<h1 class="text-sm font-black military-label text-white tracking-tighter uppercase">
									EXILE_<span class="text-rust-light">OS</span>
								</h1>
							</div>
							<span class="text-[6px] font-mono text-stone-600 tracking-[0.3em]">MOBILE_INTERFACE_V1</span>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<div class="p-2 border border-stone-800 bg-stone-900/50 text-stone-500">
							<SettingsIcon class="w-4 h-4" />
						</div>
					</div>
				</div>

				<div class="flex flex-1 overflow-hidden">
					<!-- Mock Sidebar -->
					<div class="w-16 bg-[var(--sidebar-bg)] backdrop-blur-xl border-r border-stone-800 flex-col items-center py-4 gap-4 relative z-20 hidden md:flex shadow-2xl">
						{#each [
							{ icon: LayoutDashboard, active: true },
							{ icon: Activity, active: false },
							{ icon: SettingsIcon, active: false },
							{ icon: Shield, active: false }
						] as nav}
							<div class="w-10 h-10 flex items-center justify-center rounded-xl transition-all {nav.active ? 'bg-gradient-to-r from-rust/30 to-rust/20 text-rust shadow-lg border border-rust/40' : 'text-stone-600 hover:bg-white/5'}">
								<nav.icon class="w-5 h-5" />
							</div>
						{/each}
						
						<div class="mt-auto p-2">
							<div class="w-8 h-8 rounded-lg bg-stone-900/50 border border-stone-800 flex items-center justify-center text-stone-600">
								<Eye class="w-4 h-4" />
							</div>
						</div>
					</div>

					<!-- Mock Content -->
					<div class="flex-1 overflow-y-auto p-6 space-y-8 bg-transparent relative z-10 custom-scrollbar">
						<!-- Content Stats -->
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
							<StatsCard title="SIGNAL_LOAD" value="84.2%" Icon={Activity} color="rust" />
							<StatsCard title="UPLINK_SYNC" value="ACTIVE" Icon={CheckCircle} color="emerald" />
						</div>

						<!-- Content List -->
						<div class="modern-industrial-card !bg-black/40 !rounded-none overflow-hidden">
							<div class="p-3 border-b border-stone-800/50 flex justify-between items-center bg-stone-950/20">
								<span class="text-[9px] font-black text-stone-500 uppercase tracking-widest italic">Node Registry</span>
								<Search class="w-3 h-3 text-stone-700" />
							</div>
							<div class="p-0">
								{#each [1, 2, 3] as i}
									<div class="flex items-center justify-between p-3 border-b border-stone-800/30 hover:bg-rust/5 transition-all group cursor-pointer">
										<div class="flex items-center gap-3">
											<div class="w-7 h-7 bg-stone-950 border border-stone-800 flex items-center justify-center group-hover:border-rust transition-all">
												<div class="w-1 h-1 bg-rust shadow-[0_0_5px_var(--color-rust)]"></div>
											</div>
											<div class="flex flex-col">
												<span class="text-[10px] font-black text-stone-200 uppercase tracking-widest">UNIT_0{i}</span>
												<span class="text-[7px] font-mono text-stone-600 uppercase">Sector: Delta-9</span>
											</div>
										</div>
										<ChevronDown class="w-3 h-3 text-stone-700 group-hover:text-rust" />
									</div>
								{/each}
							</div>
						</div>

						<!-- Mock Buttons -->
						<div class="flex flex-wrap gap-3">
							<button class="px-5 py-2.5 bg-rust text-white font-heading font-black text-[9px] uppercase tracking-widest shadow-lg shadow-rust/20">Execute</button>
							<button class="px-5 py-2.5 bg-stone-900 border border-stone-800 text-stone-400 font-heading font-black text-[9px] uppercase tracking-widest hover:border-rust/30 transition-all">Standby</button>
							<button class="px-5 py-2.5 bg-red-600 text-white font-heading font-black text-[9px] uppercase tracking-widest shadow-lg shadow-red-900/20">Terminate</button>
						</div>

						<!-- Mock Description -->
						<div class="p-5 border-l-2 border-rust/30 bg-stone-900/20 space-y-2">
							<span class="text-[8px] font-black text-rust uppercase tracking-widest">[KERNEL_REPORT]</span>
							<p class="text-[11px] text-stone-400 uppercase leading-relaxed font-jetbrains">All decentralized modules are synchronizing with the central controller via encrypted neural link v4.2.</p>
						</div>
					</div>
				</div>

				<!-- Background Canvas Simulation Overlay -->
				{#if $backgroundConfig.global_type !== 'none'}
					<div class="absolute inset-0 z-0 pointer-events-none opacity-20">
						<div class="w-full h-full bg-[radial-gradient(circle_at_center,var(--primary-color)_0%,transparent_70%)] opacity-10"></div>
					</div>
				{/if}
			</div>

			<!-- Typography Matrix -->
			<div class="modern-industrial-card glass-panel !rounded-none">
				<div class="p-6 border-b border-stone-800 bg-stone-950/40 flex items-center gap-4">
					<Edit3 class="w-5 h-5 text-rust" />
					<h3 class="font-heading font-black text-sm text-white uppercase tracking-widest">Typography Matrix</h3>
				</div>
				<div class="p-8 space-y-8">
					<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
						<!-- Header Font -->
						<div class="space-y-4">
							<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest block">Header Typeface</span>
							<div class="grid grid-cols-1 gap-2">
								{#each ['Kanit', 'Orbitron', 'Rajdhani', 'Black Ops One', 'Syncopate', 'Michroma', 'Montserrat', 'Playfair Display'] as font}
									<button 
										onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, font_header: font } }))}
										class="px-3 py-2 border transition-all text-[10px] font-black uppercase text-left industrial-frame {$siteSettings.aesthetic.font_header === font ? 'border-rust text-rust-light bg-rust/5' : 'border-stone-800 text-stone-600 hover:border-stone-700'}"
										style="font-family: '{font}', sans-serif;"
									>
										{font}
									</button>
								{/each}
							</div>
						</div>

						<!-- Body Font -->
						<div class="space-y-4">
							<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest block">Body Typeface</span>
							<div class="grid grid-cols-1 gap-2">
								{#each ['Inter', 'Roboto', 'Lato', 'Open Sans', 'Poppins', 'Space Grotesk', 'Red Hat Mono'] as font}
									<button 
										onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, font_body: font } }))}
										class="px-3 py-2 border transition-all text-[10px] font-medium uppercase text-left industrial-frame {$siteSettings.aesthetic.font_body === font ? 'border-rust text-rust-light bg-rust/5' : 'border-stone-800 text-stone-600 hover:border-stone-700'}"
										style="font-family: '{font}', sans-serif;"
									>
										{font}
									</button>
								{/each}
							</div>
						</div>

						<!-- Mono Font -->
						<div class="space-y-4">
							<span class="text-[10px] font-black text-stone-500 uppercase tracking-widest block">Terminal / Mono</span>
							<div class="grid grid-cols-1 gap-2">
								{#each ['JetBrains Mono', 'Fira Code', 'Red Hat Mono', 'Space Grotesk'] as font}
									<button 
										onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, font_mono: font } }))}
										class="px-3 py-2 border transition-all text-[10px] font-mono uppercase text-left industrial-frame {$siteSettings.aesthetic.font_mono === font ? 'border-rust text-rust-light bg-rust/5' : 'border-stone-800 text-stone-600 hover:border-stone-700'}"
										style="font-family: '{font}', monospace;"
									>
										{font}
									</button>
								{/each}
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Interface Architecture -->
			<div class="modern-industrial-card glass-panel !rounded-none">
				<div class="p-6 border-b border-stone-800 bg-stone-950/40 flex items-center gap-4">
					<Database class="w-5 h-5 text-rust" />
					<h3 class="font-heading font-black text-sm text-white uppercase tracking-widest">Interface Architecture</h3>
				</div>
				<div class="p-8 space-y-8">
					<div class="grid grid-cols-1 md:grid-cols-2 gap-10">
						<!-- Radius Controls -->
						<div class="space-y-6">
							<h4 class="text-[10px] font-black text-stone-400 uppercase tracking-widest border-b border-stone-800 pb-2">Corner Geometry</h4>
							{#each [
								{ key: 'border_radius_sm', label: 'Radius Small (SM)', max: 10 },
								{ key: 'border_radius_md', label: 'Radius Medium (MD)', max: 20 },
								{ key: 'border_radius_lg', label: 'Radius Large (LG)', max: 40 }
							] as radius}
								<div class="space-y-2">
									<div class="flex justify-between items-center">
										<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">{radius.label}</span>
										<span class="text-[9px] font-mono text-rust-light">{(($siteSettings.aesthetic as any)[radius.key])}px</span>
									</div>
									<input 
										type="range" 
										min="0" 
										max={radius.max} 
										step="1" 
										value={($siteSettings.aesthetic as any)[radius.key]} 
										oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, [radius.key]: parseInt(e.currentTarget.value) } }))} 
										class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
									/>
								</div>
							{/each}

							<div class="space-y-4 pt-4 border-t border-stone-800">
								<div class="space-y-2">
									<div class="flex justify-between items-center">
										<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Border Weight</span>
										<span class="text-[9px] font-mono text-rust-light">{$siteSettings.aesthetic.card_border_width}px</span>
									</div>
									<input 
										type="range" 
										min="0" 
										max="5" 
										step="1" 
										value={$siteSettings.aesthetic.card_border_width} 
										oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, card_border_width: parseInt(e.currentTarget.value) } }))} 
										class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
									/>
								</div>

								<div class="space-y-2">
									<div class="flex justify-between items-center">
										<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Shadow Vector</span>
										<span class="text-[9px] font-mono text-rust-light">{$siteSettings.aesthetic.card_shadow_size}px</span>
									</div>
									<input 
										type="range" 
										min="0" 
										max="50" 
										step="2" 
										value={$siteSettings.aesthetic.card_shadow_size} 
										oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, card_shadow_size: parseInt(e.currentTarget.value) } }))} 
										class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
									/>
								</div>
							</div>
						</div>

						<!-- Material Properties -->
						<div class="space-y-6">
							<h4 class="text-[10px] font-black text-stone-400 uppercase tracking-widest border-b border-stone-800 pb-2">Material Physics</h4>
							
							<div class="space-y-2">
								<div class="flex justify-between items-center">
									<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest">Glass Density</span>
									<span class="text-[9px] font-mono text-rust-light">{$siteSettings.aesthetic.glass_strength}</span>
								</div>
								<input 
									type="range" 
									min="0" 
									max="1" 
									step="0.05" 
									value={$siteSettings.aesthetic.glass_strength} 
									oninput={e => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, glass_strength: parseFloat(e.currentTarget.value) } }))} 
									class="w-full h-1 bg-stone-800 rounded-lg appearance-none cursor-pointer accent-rust" 
								/>
							</div>

							<div class="space-y-2">
								<span class="text-[9px] font-bold text-stone-500 uppercase tracking-widest block mb-2">Modal Entry Vector</span>
								<div class="grid grid-cols-3 gap-2">
									{#each ['scale', 'slide', 'fade'] as anim}
										<button
											onclick={() => siteSettings.update(s => ({ ...s, aesthetic: { ...s.aesthetic, modal_animation: anim } }))}
											class="py-2 border text-[9px] font-black uppercase tracking-widest transition-all {$siteSettings.aesthetic.modal_animation === anim ? 'bg-rust/20 border-rust text-white' : 'bg-stone-950 border-stone-800 text-stone-600 hover:border-stone-700'}"
										>
											{anim}
										</button>
									{/each}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	/* Hide scrollbar for Chrome, Safari and Opera */
	.no-scrollbar::-webkit-scrollbar {
		display: none;
	}

	/* Hide scrollbar for IE, Edge and Firefox */
	.no-scrollbar {
		-ms-overflow-style: none; /* IE and Edge */
		scrollbar-width: none; /* Firefox */
	}

	input[type="range"] {
		@apply hover:brightness-110 transition-all duration-300;
	}
</style>