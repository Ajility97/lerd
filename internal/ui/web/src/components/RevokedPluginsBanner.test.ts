import { render, fireEvent, waitFor } from '@testing-library/svelte';
import { it, expect, beforeEach, vi } from 'vitest';
import RevokedPluginsBanner from './RevokedPluginsBanner.svelte';
import { revokedPlugins } from '$stores/plugins';

const calls: { method: string; url: string }[] = [];

beforeEach(() => {
  calls.length = 0;
  globalThis.fetch = vi.fn(async (url: string, init?: RequestInit) => {
    calls.push({ method: init?.method ?? 'GET', url: String(url) });
    return new Response('[]');
  }) as unknown as typeof fetch;
});

// A revoked plugin is named with the catalog's reason, and stays on screen
// until the user removes it or keeps it: there is nothing to dismiss.
it('names a revoked plugin with its reason and keeps it on request', async () => {
  revokedPlugins.set([{ name: 'bad', reason: "Sends the site's .env away", at: '2026-10-24' }]);
  const { getByText, getByRole, queryByLabelText } = render(RevokedPluginsBanner);
  expect(getByText('bad was stopped')).toBeTruthy();
  expect(getByText("Sends the site's .env away")).toBeTruthy();
  expect(queryByLabelText(/dismiss/i)).toBeNull();

  await fireEvent.click(getByRole('button', { name: 'Keep using it' }));
  await waitFor(() => expect(calls.some((c) => c.method === 'POST' && c.url.endsWith('/api/plugin-manager/bad/keep'))).toBe(true));
});
