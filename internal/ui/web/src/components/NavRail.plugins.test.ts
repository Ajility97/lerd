import { render, fireEvent, waitFor } from '@testing-library/svelte';
import { it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import NavRail from './NavRail.svelte';
import { services } from '$stores/services';
import { accessMode } from '$stores/accessMode';
import { dashboardOpen } from '$stores/dashboard';

const PLUGINS = [
  { name: 'hello', page: { title: 'Hello', url: '/plugins/hello/ui/index.html', icon: '/plugins/hello/ui/icon.svg' } },
  { name: 'quiet', site_tab: { title: 'Quiet', url: '/plugins/quiet/ui/site.html' } }
];

beforeEach(() => {
  globalThis.fetch = (async (url: string) =>
    new Response(JSON.stringify(String(url).endsWith('/api/plugins') ? PLUGINS : {}), {
      status: 200
    })) as unknown as typeof fetch;
  accessMode.set({ localControl: true, local: true, lanExposed: false, checked: true });
  dashboardOpen.set(null);
  services.set([]);
});

// Only a plugin with a page earns a rail button, and it opens in the overlay as
// a plugin page rather than a service dashboard.
it('draws a launcher for each plugin page and opens it', async () => {
  const { findByLabelText, queryByLabelText } = render(NavRail);
  const button = await findByLabelText('Hello');
  expect(queryByLabelText('Quiet')).toBeNull();

  await fireEvent.click(button);
  await waitFor(() => expect(get(dashboardOpen)?.plugin).toBe('hello'));
  expect(get(dashboardOpen)?.dashboard).toBe('/plugins/hello/ui/index.html');
  expect(location.hash).toBe('#plugin/hello');
});

// The plugin group always ends with the way to the Plugins page, so plugins can
// be found from the rail, and it opens the installed list, empty or not.
it('ends the plugin group with a Plugins button under the rule', async () => {
  const { findByLabelText, container, unmount } = render(NavRail);
  const page = await findByLabelText('Hello');
  const button = await findByLabelText('Plugins');
  const divider = container.querySelector('[data-rail-divider]')!;
  expect(divider.compareDocumentPosition(page) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  expect(page.compareDocumentPosition(button) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  await fireEvent.click(button);
  expect(location.hash).toBe('#plugins');
  unmount();

  globalThis.fetch = (async () => new Response('[]')) as unknown as typeof fetch;
  const again = render(NavRail);
  await waitFor(() => expect(again.queryByLabelText('Hello')).toBeNull());
  expect(again.container.querySelector('[data-rail-divider]')).not.toBeNull();
  location.hash = '';
  await fireEvent.click(again.getByLabelText('Plugins'));
  expect(location.hash).toBe('#plugins');
});
