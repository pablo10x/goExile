<script lang="ts">
	import Iconify from '@iconify/svelte';
	import { siteSettings } from '$lib/stores';

	interface Props {
		name: string;
		class?: string;
		size?: string | number;
		strokeWidth?: number;
		style?: string;
	}

	let { name, class: className = '', size = '1em', strokeWidth, style = '' }: Props = $props();

	// Map generic names to specific icon names in different packs
	// This ensures that when the user switches packs, the right icon is still shown
	const iconMap: Record<string, Record<string, string>> = {
		'dashboard': {
			'lucide': 'lucide:layout-dashboard',
			'mdi': 'mdi:view-dashboard',
			'ph': 'ph:squares-four-bold',
			'ri': 'ri:dashboard-line',
			'tabler': 'tabler:layout-dashboard'
		},
		'gauge': {
			'lucide': 'lucide:gauge',
			'mdi': 'mdi:gauge',
			'ph': 'ph:gauge-bold',
			'ri': 'ri:speed-up-line',
			'tabler': 'tabler:gauge'
		},
		'server': {
			'lucide': 'lucide:server',
			'mdi': 'mdi:server',
			'ph': 'ph:server-bold',
			'ri': 'ri:server-line',
			'tabler': 'tabler:server'
		},
		'cpu': {
			'lucide': 'lucide:cpu',
			'mdi': 'mdi:cpu-64-bit',
			'ph': 'ph:cpu-bold',
			'ri': 'ri:cpu-line',
			'tabler': 'tabler:cpu'
		},
		'database': {
			'lucide': 'lucide:database',
			'mdi': 'mdi:database',
			'ph': 'ph:database-bold',
			'ri': 'ri:database-2-line',
			'tabler': 'tabler:database'
		},
		'users': {
			'lucide': 'lucide:users',
			'mdi': 'mdi:account-group',
			'ph': 'ph:users-bold',
			'ri': 'ri:group-line',
			'tabler': 'tabler:users'
		},
		'settings': {
			'lucide': 'lucide:settings',
			'mdi': 'mdi:cog',
			'ph': 'ph:gear-bold',
			'ri': 'ri:settings-3-line',
			'tabler': 'tabler:settings'
		},
		'sliders': {
			'lucide': 'lucide:sliders',
			'mdi': 'mdi:tune-vertical',
			'ph': 'ph:sliders-bold',
			'ri': 'ri:equalizer-line',
			'tabler': 'tabler:sliders'
		},
		'palette': {
			'lucide': 'lucide:palette',
			'mdi': 'mdi:palette',
			'ph': 'ph:palette-bold',
			'ri': 'ri:palette-line',
			'tabler': 'tabler:palette'
		},
		'file-text': {
			'lucide': 'lucide:file-text',
			'mdi': 'mdi:file-document',
			'ph': 'ph:file-text-bold',
			'ri': 'ri:file-text-line',
			'tabler': 'tabler:file-text'
		},
		'shield': {
			'lucide': 'lucide:shield-check',
			'mdi': 'mdi:shield-check',
			'ph': 'ph:shield-check-bold',
			'ri': 'ri:shield-check-line',
			'tabler': 'tabler:shield-check'
		},
		'activity': {
			'lucide': 'lucide:activity',
			'mdi': 'mdi:pulse',
			'ph': 'ph:activity-bold',
			'ri': 'ri:pulse-line',
			'tabler': 'tabler:activity'
		},
		'plus': {
			'lucide': 'lucide:plus',
			'mdi': 'mdi:plus',
			'ph': 'ph:plus-bold',
			'ri': 'ri:add-line',
			'tabler': 'tabler:plus'
		},
		'upload': {
			'lucide': 'lucide:upload',
			'mdi': 'mdi:upload',
			'ph': 'ph:upload-simple-bold',
			'ri': 'ri:upload-line',
			'tabler': 'tabler:upload'
		},
		'hard-drive': {
			'lucide': 'lucide:hard-drive',
			'mdi': 'mdi:harddisk',
			'ph': 'ph:hard-drive-bold',
			'ri': 'ri:hard-drive-2-line',
			'tabler': 'tabler:device-floppy'
		},
		'clock': {
			'lucide': 'lucide:clock',
			'mdi': 'mdi:clock-outline',
			'ph': 'ph:clock-bold',
			'ri': 'ri:time-line',
			'tabler': 'tabler:clock'
		},
		'radio': {
			'lucide': 'lucide:radio',
			'mdi': 'mdi:radiobox-marked',
			'ph': 'ph:radio-button-bold',
			'ri': 'ri:radio-button-line',
			'tabler': 'tabler:radio'
		},
		'globe': {
			'lucide': 'lucide:globe',
			'mdi': 'mdi:earth',
			'ph': 'ph:globe-bold',
			'ri': 'ri:global-line',
			'tabler': 'tabler:globe'
		},
		'alert': {
			'lucide': 'lucide:alert-circle',
			'mdi': 'mdi:alert-circle-outline',
			'ph': 'ph:warning-circle-bold',
			'ri': 'ri:error-warning-line',
			'tabler': 'tabler:alert-circle'
		}
	};

	let resolvedIcon = $derived(() => {
		const pack = $siteSettings.aesthetic.icon_pack || 'lucide';
		// If name is already a full iconify name (contains :), use it directly
		if (name.includes(':')) return name;
		
		const entry = iconMap[name.toLowerCase()];
		if (entry && entry[pack]) return entry[pack];
		
		// Fallback to lucide or just the name if no mapping found
		return name.includes(':') ? name : `lucide:${name}`;
	});

	// Handle stroke width for icon sets that support it (Lucide, Tabler)
	let computedStyle = $derived.by(() => {
		const currentStroke = strokeWidth ?? $siteSettings.aesthetic.icon_stroke ?? 2;
		return `${style}; --icon-stroke: ${currentStroke}px;`;
	});
</script>

<Iconify 
	icon={resolvedIcon()} 
	class={className} 
	width={size} 
	height={size}
	style={computedStyle}
/>
