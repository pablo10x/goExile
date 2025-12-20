export interface FirewallRule {
    id: number;
    name: string;
    cidr: string;
    port: string;
    protocol: string;
    action: 'ALLOW' | 'DENY';
    enabled: boolean;
    created_at: string;
}

export interface FirewallLog {
    id: number;
    rule_id: number | null;
    source_ip: string;
    dest_port: number;
    protocol: string;
    action: string;
    timestamp: string;
}
