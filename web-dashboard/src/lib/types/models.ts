// Core models for goExile dashboard
// These should match the backend structures in server/models/models.go

export interface Node {
	id: number;
	name: string;
	region: string;
	host: string;
	port: number;
	max_instances: number;
	current_instances: number;
	status: string;
	last_seen: string;
	is_draining: boolean;
	tags?: string;
	maintenance_window?: string;
	resource_limits?: string;
	public_ip?: string;
	cpu_usage: number;
	mem_used: number;
	mem_total: number;
	disk_used: number;
	disk_total: number;
	game_version: string;
}

export interface GameServerVersion {
	id: number;
	filename: string;
	version: string;
	comment: string;
	uploaded_at: string;
	is_active: boolean;
}

export interface ServerConfig {
	id: number;
	key: string;
	value: string;
	type: string;
	category: string;
	description: string;
	is_read_only: boolean;
	requires_restart: boolean;
	updated_at: string;
	updated_by: string;
}

export interface Note {
	id: number;
	title: string;
	content: string;
	color: string;
	status: string;
	rotation: number;
	created_at: string;
	updated_at: string;
}

export interface Todo {
	id: number;
	parent_id?: number;
	content: string;
	done: boolean;
	in_progress: boolean;
	created_at: string;
	deadline?: string;
	sub_tasks?: Todo[];
	comments?: TodoComment[];
}

export interface TodoComment {
	id: number;
	todo_id: number;
	content: string;
	author: string;
	created_at: string;
}

export interface SystemLog {
	id: number;
	timestamp: string;
	level: string;
	category: string;
	source: string;
	message: string;
	details: string;
	client_ip: string;
	path: string;
	method: string;
}

export interface RedEyeRule {
	id: number;
	name: string;
	cidr: string;
	port: string;
	path_pattern: string;
	protocol: string;
	action: 'ALLOW' | 'DENY' | 'RATE_LIMIT';
	rate_limit: number;
	burst: number;
	enabled: boolean;
	created_at: string;
}

export interface RedEyeLog {
	id: number;
	rule_id?: number;
	source_ip: string;
	dest_port: number;
	protocol: string;
	action: string;
	details: string;
	timestamp: string;
}
