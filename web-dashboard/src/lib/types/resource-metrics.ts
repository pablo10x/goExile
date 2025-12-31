// Resource metrics types and interfaces for the goExile dashboard

export interface ResourceStats {
	cpu_percent: number;
	memory_usage: number;
	disk_usage: number;
	status: string;
	uptime: number;
}

export interface ResourceHistory {
	timestamp: string;
	cpu: number;
	memory_percent: number;
}

export interface PeakResourceStats {
	peakCpu: number;
	peakMemory: number;
	peakDisk: number;
}

export interface ResourceConsumer {
	id: number;
	nodeId: number;
	instanceId: string;
	port: number;
	cpu_percent: number;
	memory_usage: number;
	disk_usage: number;
	status: string;
	region: string;
	host: string;
}

export interface ResourceMetricsState {
	stats: ResourceStats | null;
	history: ResourceHistory[];
	peakStats: PeakResourceStats;
	loading: boolean;
	error: string | null;
}

export type ResourceType = 'cpu' | 'memory' | 'disk';
export type ResourceColor = 'orange' | 'blue' | 'green' | 'red' | 'purple' | 'teal' | 'emerald';
export type TrendDirection = 'up' | 'down' | 'stable';

export interface ResourceStatsCardProps {
	title: string;
	current: number;
	peak: number;
	unit: string;
	icon: string;
	color: ResourceColor;
	trend?: TrendDirection;
	animated?: boolean;
}

export interface ResourceProgressBarProps {
	label: string;
	value: number;
	max: number;
	color: ResourceColor;
	animated?: boolean;
	showThreshold?: boolean;
	threshold?: number;
	height?: string;
}

export interface ResourceMetricsPanelProps {
	nodeId: number;
	instanceId: number | string;
	memTotal?: number;
	diskTotal?: number;
	height?: number;
	compact?: boolean;
	showTitle?: boolean;
	autoRefresh?: boolean;
}

export interface TopResourceConsumersProps {
	limit?: number;
	refreshInterval?: number;
	resourceType?: ResourceType;
	compact?: boolean;
}
