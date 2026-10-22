import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import { tick } from 'svelte';
import { status } from '$stores/status';
import StreamingToggle from './StreamingToggle.svelte';

vi.mock('$stores/workspaces', () => ({ setStreamingMode: vi.fn(async () => ({ ok: true })) }));

describe('StreamingToggle', () => {
  it('is absent while streaming mode is disabled', () => {
    status.set({ streaming_enabled: false } as never);
    const { queryByRole } = render(StreamingToggle);
    expect(queryByRole('button')).toBeNull();
  });

  it('shows once streaming mode is enabled', () => {
    status.set({ streaming_enabled: true, streaming_mode: false } as never);
    const { getByRole } = render(StreamingToggle);
    expect(getByRole('button', { pressed: false })).toBeTruthy();
  });

  it('stays busy until the status reports the new mode', async () => {
    status.set({ streaming_enabled: true, streaming_mode: true } as never);
    const { getByRole } = render(StreamingToggle);
    await fireEvent.click(getByRole('button'));
    expect(getByRole('button').getAttribute('aria-busy')).toBe('true');

    status.set({ streaming_enabled: true, streaming_mode: false } as never);
    await tick();
    expect(getByRole('button').getAttribute('aria-busy')).toBe('false');
  });

  it('stays free when the mode later changes elsewhere', async () => {
    status.set({ streaming_enabled: true, streaming_mode: true } as never);
    const { getByRole } = render(StreamingToggle);
    await fireEvent.click(getByRole('button'));
    status.set({ streaming_enabled: true, streaming_mode: false } as never);
    await tick();

    status.set({ streaming_enabled: true, streaming_mode: true } as never);
    await tick();
    expect(getByRole('button').getAttribute('aria-busy')).toBe('false');
  });
});
