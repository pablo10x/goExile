import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import StatsCard from './StatsCard.svelte';

describe('StatsCard Component', () => {
    it('renders title and value correctly', () => {
        render(StatsCard, { title: 'Test Title', value: '123' });
        expect(screen.getByText('Test Title')).toBeInTheDocument();
        expect(screen.getByText('123')).toBeInTheDocument();
    });

    it('renders subValue when provided', () => {
        render(StatsCard, { title: 'Network', value: '', subValue: '10 MB / 20 MB' });
        expect(screen.getByText('Network')).toBeInTheDocument();
        expect(screen.getByText(/10 MB \/ 20 MB/)).toBeInTheDocument();
    });
});