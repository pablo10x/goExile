<script lang="ts">
	import Iconify from '@iconify/svelte';

	interface Props {
		name: string;
		class?: string;
		size?: string | number;
		strokeWidth?: number;
		style?: string;
	}

	let { name, class: className = '', size = '1em', strokeWidth, style = '' }: Props = $props();

	// Map generic names to specific icon names in Phosphor (ph) pack as default
	const iconMap: Record<string, string> = {
		'dashboard': 'ph:squares-four-bold',
		'gauge': 'ph:gauge-bold',
		'server': 'ph:server-bold',
		'cpu': 'ph:cpu-bold',
		'database': 'ph:database-bold',
		'users': 'ph:users-bold',
		'settings': 'ph:gear-bold',
		'sliders': 'ph:sliders-bold',
		'palette': 'ph:palette-bold',
		'file-text': 'ph:file-text-bold',
		'shield': 'ph:shield-check-bold',
		'activity': 'ph:activity-bold',
		'plus': 'ph:plus-bold',
		'upload': 'ph:upload-simple-bold',
		'hard-drive': 'ph:hard-drive-bold',
		'clock': 'ph:clock-bold',
		'radio': 'ph:radio-button-bold',
		'globe': 'ph:globe-bold',
		'alert': 'ph:warning-circle-bold'
	};

	let resolvedIcon = $derived(() => {
		// If name is already a full iconify name (contains :), use it directly
		if (name.includes(':')) return name;
		
		const entry = iconMap[name.toLowerCase()];
		if (entry) return entry;
		
		// Fallback to ph or just the name if no mapping found
		return `ph:${name}`;
	});

	let computedStyle = $derived.by(() => {
		const currentStroke = strokeWidth ?? 2;
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