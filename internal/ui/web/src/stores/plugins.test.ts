import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { get } from 'svelte/store';

describe('plugins store', () => {
  const realFetch = globalThis.fetch;

  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    globalThis.fetch = realFetch;
  });

  it('reloads both plugin lists after the catalog check', async () => {
    const fetchSpy = vi.fn(
      async () =>
        new Response(JSON.stringify([{ name: 'notes', update: true }]), { status: 200, headers: { 'Content-Type': 'application/json' } })
    );
    globalThis.fetch = fetchSpy as unknown as typeof fetch;
    const { pluginUpdates, managedPlugins } = await import('./plugins');
    const { wsMessage } = await import('$lib/ws');

    wsMessage.set({ type: 'plugins' });
    await vi.waitFor(() => expect(get(pluginUpdates)).toBe(1));

    expect(get(managedPlugins)).toHaveLength(1);
  });

  it('leaves the lists alone for other messages', async () => {
    const fetchSpy = vi.fn(async () => new Response('[]', { status: 200 }));
    globalThis.fetch = fetchSpy as unknown as typeof fetch;
    await import('./plugins');
    const { wsMessage } = await import('$lib/ws');

    wsMessage.set({ type: 'notification', notification: { kind: 'update_available', data: { version: '1.38.0' } } });

    expect(fetchSpy).not.toHaveBeenCalled();
  });
});
