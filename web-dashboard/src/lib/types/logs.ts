export interface SystemLog {
	id: number;
	timestamp: string;
	level: 'INFO' | 'WARN' | 'ERROR' | 'FATAL';
	category: 'Internal' | 'Node' | 'Security' | 'General';
	source: string;
	message: string;
	details: string;
	client_ip: string;
	path: string;
	method: string;
}

export interface LogsResponse {
	logs: SystemLog[];
	total: number;
	limit: number;
	offset: number;
}
