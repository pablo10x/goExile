export function parseVersion(v: string): number[] {
    if (!v) return [0, 0, 0];
    // Remove 'v' prefix if present
    const clean = v.replace(/^v/, '');
    const parts = clean.split('.').map(Number);
    while (parts.length < 3) parts.push(0);
    return parts;
}

/**
 * Compare two semantic versions.
 * Returns:
 *  1 if v1 > v2
 * -1 if v1 < v2
 *  0 if v1 == v2
 */
export function compareVersions(v1: string, v2: string): number {
    const p1 = parseVersion(v1);
    const p2 = parseVersion(v2);

    for (let i = 0; i < 3; i++) {
        if (p1[i] > p2[i]) return 1;
        if (p1[i] < p2[i]) return -1;
    }
    return 0;
}

export function isNewer(current: string, target: string): boolean {
    return compareVersions(target, current) > 0;
}

export function isOlder(current: string, target: string): boolean {
    return compareVersions(target, current) < 0;
}
