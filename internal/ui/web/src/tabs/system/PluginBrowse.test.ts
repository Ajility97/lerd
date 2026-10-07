import { render, fireEvent, waitFor, within } from '@testing-library/svelte';
import { it, expect, vi, beforeEach } from 'vitest';
import PluginBrowse from './PluginBrowse.svelte';

const A = 'a'.repeat(40);
let catalog: object[];
let posts: unknown[];

beforeEach(() => {
  catalog = [
    { name: 'old', repo: 'r', commit: A, stars: 50, updated_at: '2026-01-01', installed: false },
    { name: 'fresh', repo: 'r', commit: 'b'.repeat(40), stars: 2, updated_at: '2026-10-01', installed: false, images: ['https://plugins.lerd.sh/images/fresh/1.png'] },
    { name: 'mine', repo: 'r', commit: A, installed: true }
  ];
  posts = [];
  globalThis.fetch = vi.fn(async (url: string, init?: RequestInit) => {
    if (init?.method === 'POST') {
      posts.push(JSON.parse(init.body as string));
      return new Response('{}');
    }
    if (String(url).endsWith('/catalog')) return new Response(JSON.stringify(catalog));
    return new Response('[]');
  }) as unknown as typeof fetch;
});

const names = (c: HTMLElement) => [...c.querySelectorAll('article')].map((a) => a.id.replace('plugin-card-', ''));

it('lists the catalog newest first, with its preview, and marks what is installed', async () => {
  const { container, findByText, getByText } = render(PluginBrowse);
  await findByText('fresh');
  expect(names(container)).toEqual(['fresh', 'old', 'mine']);
  expect(container.querySelector('#plugin-card-fresh img')?.getAttribute('src')).toBe('https://plugins.lerd.sh/images/fresh/1.png');
  expect(within(container.querySelector('#plugin-card-mine') as HTMLElement).getByText('Installed')).toBeTruthy();
  expect(getByText('★ 50')).toBeTruthy();
});

// The commit is in the question because it is the one thing the user is
// trusting: lerd installs that and nothing newer.
it('installs by name after confirming the approved commit', async () => {
  const { findByText, getAllByRole, getByRole } = render(PluginBrowse);
  await findByText('old');
  await fireEvent.click(getAllByRole('button', { name: 'Install' })[1]);
  const dialog = getByRole('dialog');
  expect(dialog.textContent).toContain('aaaaaaa');
  await fireEvent.click(within(dialog).getByRole('button', { name: 'Install' }));
  await waitFor(() => expect(posts).toEqual([{ name: 'old' }]));
});

it('says why when the catalog cannot be reached', async () => {
  globalThis.fetch = vi.fn(async () => new Response('fetching the plugin catalog: no route', { status: 502 })) as unknown as typeof fetch;
  const { findByText } = render(PluginBrowse);
  await findByText('The plugin catalog could not be loaded');
  await findByText(/no route/);
});
