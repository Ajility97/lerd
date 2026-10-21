import { describe, it, expect, vi, beforeEach } from 'vitest';

const { loadSites } = vi.hoisted(() => ({ loadSites: vi.fn() }));
vi.mock('$stores/sites', () => ({ loadSites }));

import { SiteAction } from './siteAction.svelte';

describe('SiteAction', () => {
  beforeEach(() => loadSites.mockReset().mockResolvedValue(undefined));

  it('reloads the sites and flashes the success message', async () => {
    const a = new SiteAction();
    await a.run(async () => ({ ok: true }), 'Added');
    expect(loadSites).toHaveBeenCalled();
    expect(a.flash).toBe('Added');
    expect(a.loading).toBe(false);
  });

  it('keeps the error and skips the reload when the action fails', async () => {
    const a = new SiteAction();
    await a.run(async () => ({ ok: false, error: 'taken' }), 'Added');
    expect(a.error).toBe('taken');
    expect(a.flash).toBe('');
    expect(loadSites).not.toHaveBeenCalled();
  });

  it('clears the flash after a few seconds', async () => {
    vi.useFakeTimers();
    const a = new SiteAction();
    await a.run(async () => ({ ok: true }), 'Added');
    vi.advanceTimersByTime(3000);
    expect(a.flash).toBe('');
    vi.useRealTimers();
  });
});
