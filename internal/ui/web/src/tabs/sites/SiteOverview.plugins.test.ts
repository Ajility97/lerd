import { render } from '@testing-library/svelte';
import { it, expect, beforeEach, vi } from 'vitest';
import SiteOverview from './SiteOverview.svelte';
import { plugins } from '$stores/plugins';
import type { Site } from '$stores/sites';

beforeEach(() => {
  plugins.set([
    { name: 'stats', overview_card: { title: 'Requests, last 24h', url: '/plugins/stats/ui/card.html' } },
    { name: 'quiet', overview_card: { title: 'Quiet card', url: '/plugins/quiet/ui/card.html' } }
  ]);
  globalThis.fetch = vi.fn(async () => new Response('{}')) as unknown as typeof fetch;
});

const site = (extra: Partial<Site> = {}) => ({ name: 'shop', domain: 'shop.test', path: '/srv/shop', services: [], ...extra }) as unknown as Site;

// The cards come with the site, so a switch made in any dashboard reaches an
// Overview already open with the next live update, and nothing is fetched for it.
it('draws the cards the site has switched on, and follows a switch made elsewhere', async () => {
  const { findByText, queryByText, container, rerender } = render(SiteOverview, { props: { site: site({ overview_cards: ['stats'] }) } });
  await findByText('Requests, last 24h');
  expect(queryByText('Quiet card')).toBeNull();
  expect(container.querySelector('iframe[title="Requests, last 24h"]')?.getAttribute('src')).toContain('site=shop');

  await rerender({ site: site({ overview_cards: ['quiet'] }) });
  await findByText('Quiet card');
  expect(queryByText('Requests, last 24h')).toBeNull();
  const calls = (globalThis.fetch as unknown as { mock: { calls: unknown[][] } }).mock.calls.map((c) => String(c[0]));
  expect(calls.some((u) => u.includes('/api/page-plugins'))).toBe(false);
});
