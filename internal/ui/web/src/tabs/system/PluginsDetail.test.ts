import { render, fireEvent, waitFor, within } from '@testing-library/svelte';
import { it, expect, beforeEach, vi } from 'vitest';
import PluginsDetail from './PluginsDetail.svelte';

let installed: object[];
const calls: { method: string; url: string; body?: string }[] = [];

beforeEach(() => {
  location.hash = '#system/plugins';
  installed = [
    { name: 'broken', error: 'api_version 9 needs a newer lerd' },
    { name: 'hello', version: '0.1.0', description: 'Says hello', adds: ['commands', 'page'], requires: ['python3', 'jq'], icon: '/plugins/hello/ui/icon.svg' }
  ];
  calls.length = 0;
  globalThis.fetch = vi.fn(async (url: string, init?: RequestInit) => {
    const method = init?.method ?? 'GET';
    calls.push({ method, url: String(url), body: init?.body as string | undefined });
    if (method === 'DELETE') installed = installed.filter((p) => !String(url).endsWith('/' + (p as { name: string }).name));
    if (method === 'POST') return new Response('not a lerd plugin', { status: 400 });
    if (String(url).endsWith('/api/plugin-manager')) return new Response(JSON.stringify(installed));
    return new Response('[]');
  }) as unknown as typeof fetch;
});

// A card carries the name, the version and whether it loaded; what it adds and
// requires waits in its detail.
it('lists plugins as cards and opens one to show what it adds and requires', async () => {
  const { findByText, getByText, queryByText, getByRole } = render(PluginsDetail);
  await findByText('hello');
  expect(getByText('Says hello')).toBeTruthy();
  expect(getByText('Failed to load')).toBeTruthy();
  expect(queryByText('python3')).toBeNull();

  await fireEvent.click(getByRole('button', { name: /^hello/ }));
  const dialog = within(getByRole('dialog'));
  expect(dialog.getByText('page')).toBeTruthy();
  expect(dialog.getByText('python3')).toBeTruthy();
  expect(dialog.getByText('jq')).toBeTruthy();
  expect(dialog.getByText('Not listed in the catalog')).toBeTruthy();
});

it('shows why a broken plugin failed in its detail', async () => {
  const { findByText, getByRole } = render(PluginsDetail);
  await fireEvent.click(await findByText('broken'));
  expect(await within(getByRole('dialog')).findByText(/api_version 9 needs a newer lerd/)).toBeTruthy();
});

// The author can only fix what they hear about, so a failure links straight to
// a filled-in issue on the plugin's repository.
it('shows a plugin\'s last failure with a link to report it', async () => {
  installed = [
    {
      name: 'hello',
      failure: { plugin: 'hello', surface: 'hook', what: 'site.linked', error: 'boom\nmore', at: '2026-10-15T10:00:00Z', report_url: 'https://github.com/acme/hello/issues/new?title=x' }
    }
  ];
  const { findByText, getByRole } = render(PluginsDetail);
  // The list is a store the other tests filled too, so wait for this one's
  // fetch before opening the card, and let the dialog catch up after.
  await waitFor(() => expect(calls.some((c) => c.url.endsWith('/api/plugin-manager'))).toBe(true));
  await fireEvent.click(await findByText('hello'));
  const dialog = within(getByRole('dialog'));
  expect(await dialog.findByText(/boom/)).toBeTruthy();
  expect(dialog.getByRole('link', { name: 'Report to the author' }).getAttribute('href')).toBe('https://github.com/acme/hello/issues/new?title=x');
});

// A broken plugin is exactly the one a user most needs to get rid of.
it('removes a plugin from its detail after confirming, broken ones included', async () => {
  const { findByText, getAllByRole } = render(PluginsDetail);
  await fireEvent.click(await findByText('broken'));
  await fireEvent.click(within(getAllByRole('dialog')[0]).getByRole('button', { name: 'Remove' }));
  const confirm = getAllByRole('dialog').at(-1)!;
  await fireEvent.click(within(confirm).getByRole('button', { name: 'Remove' }));
  await waitFor(() => expect(calls.some((c) => c.method === 'DELETE' && c.url.endsWith('/api/plugin-manager/broken'))).toBe(true));
});

it('installs from a folder or git in a modal and keeps the field when the install fails', async () => {
  const { getByRole, getByLabelText } = render(PluginsDetail);
  await fireEvent.click(getByRole('button', { name: 'Install from a folder or git' }));
  const input = getByLabelText('Folder path or git URL') as HTMLInputElement;
  await fireEvent.input(input, { target: { value: ' https://example.test/plugin.git ' } });
  await fireEvent.click(within(getByRole('dialog')).getByRole('button', { name: 'Install' }));
  await waitFor(() => expect(calls.some((c) => c.method === 'POST')).toBe(true));
  const post = calls.find((c) => c.method === 'POST')!;
  expect(JSON.parse(post.body!)).toEqual({ source: 'https://example.test/plugin.git' });
  expect(input.value.trim()).toBe('https://example.test/plugin.git');
});

// The two views are icon buttons at the header's far end, the current one
// marked, so the header carries only the title and its actions.
it('switches between installed and browse with icon buttons in the header', async () => {
  const { getByRole } = render(PluginsDetail);
  const installedBtn = getByRole('button', { name: 'Installed' });
  const browseBtn = getByRole('button', { name: 'Browse' });
  expect(installedBtn.textContent?.trim()).toBe('');
  expect(installedBtn.getAttribute('aria-current')).toBe('page');
  expect(browseBtn.getAttribute('aria-current')).toBeNull();
  await fireEvent.click(browseBtn);
  expect(location.hash).toBe('#system/plugins/browse');
});

// With nothing installed the page says what plugins add and offers both ways
// to get one, rather than a bare "none".
it('shows what plugins add and both ways to get one when none is installed', async () => {
  installed = [];
  const { findByText, getByText, getByRole } = render(PluginsDetail);
  await findByText('Make Lerd do more');
  expect(getByText('Overview cards')).toBeTruthy();
  expect(getByText('lerd plugin new my-plugin')).toBeTruthy();
  await fireEvent.click(getByRole('button', { name: 'Browse plugins' }));
  expect(location.hash).toBe('#system/plugins/browse');
});
