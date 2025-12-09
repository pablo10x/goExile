import { describe, it, expect } from 'vitest';
import { formatBytes, formatUptime } from './utils';

describe('Utility Functions', () => {
    describe('formatBytes', () => {
        it('formats 0 bytes correctly', () => {
            expect(formatBytes(0)).toBe('0 B');
        });

        it('formats KB correctly', () => {
            expect(formatBytes(1024)).toBe('1 KB');
        });

        it('formats MB correctly', () => {
            expect(formatBytes(1024 * 1024 * 2.5)).toBe('2.5 MB');
        });
    });

    describe('formatUptime', () => {
        it('formats 0ms correctly', () => {
            expect(formatUptime(0)).toBe('0s');
        });

        it('formats seconds correctly', () => {
            expect(formatUptime(10000)).toBe('0h 0m 10s');
        });

        it('formats minutes and seconds correctly', () => {
            expect(formatUptime(65000)).toBe('0h 1m 5s');
        });

        it('formats hours correctly', () => {
            expect(formatUptime(3665000)).toBe('1h 1m 5s');
        });
    });
});