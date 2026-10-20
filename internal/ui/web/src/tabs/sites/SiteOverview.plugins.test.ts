import { render, waitFor } from '@testing-library/svelte';
import { it, expect, beforeEach, vi } from 'vitest';
import SiteOverview from './SiteOverview.svelte';
import { plugins } from '$stores/plugins';
import type { Site } from '$stores/sites';

let switchFetches = 0;

beforeEach(() => {
  switchFetches = 0;
  plugins.set([
    { name: 'stats', overview_card: { title: 'Requests, last 24h', url: '/plugins/stats/ui/card.html' } },
    { name: 'quiet', overview_card: { title: 'Quiet card', url: '/plugins/quiet/ui/card.html' } }
  ]);
  globalThis.fetch = vi.fn(async (url: string) => {
    if (String(url).includes('/api/page-plugins')) {
      switchFetches++;
      return new Response(JSON.stringify([
        { name: 'stats', kind: 'overview_card', on: true },
        { name: 'quiet', kind: 'overview_card', on: false }
      ]));
    }
    return new Response('{}');
  }) as unknown as typeof fetch;
});

const site = (extra: Partial<Site> = {}) => ({ name: 'shop', domain: 'shop.test', path: '/srv/shop', services: [], ...extra }) as unknown as Site;

// Only a card switched on for the site is drawn, and a live update that hands
// over a fresh site object with the same name does not ask again.
it('draws the cards switched on for the site, once', async () => {
  const { findByText, queryByText, container, rerender } = render(SiteOverview, { props: { site: site() } });
  await findByText('Requests, last 24h');
  expect(queryByText('Quiet card')).toBeNull();
  const frame = container.querySelector('iframe[title="Requests, last 24h"]');
  expect(frame?.getAttribute('src')).toContain('site=shop');

  await rerender({ site: site({ php_version: '8.4' }) });
  await waitFor(() => expect(switchFetches).toBe(1));
});
