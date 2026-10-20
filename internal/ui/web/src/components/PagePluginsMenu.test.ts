import { render, fireEvent, waitFor } from '@testing-library/svelte';
import { it, expect, vi } from 'vitest';
import PagePluginsMenu from './PagePluginsMenu.svelte';

function serve(list: object[]) {
  const posts: unknown[] = [];
  globalThis.fetch = vi.fn(async (_url: string, init?: RequestInit) => {
    if (init?.method === 'POST') {
      const body = JSON.parse(init.body as string);
      posts.push(body);
      list = list.map((p) => {
        const q = p as { name: string; kind?: string };
        return q.name === body.plugin && (q.kind ?? 'page_script') === body.kind ? { ...p, on: body.on } : p;
      });
    }
    return new Response(JSON.stringify(list));
  }) as unknown as typeof fetch;
  return posts;
}

// The header keeps its icons for what the site uses; with no page plugin
// installed this menu has nothing to offer and must not take a slot.
it('draws nothing when no installed plugin has a page script', async () => {
  serve([]);
  const { container } = render(PagePluginsMenu, { props: { site: 'shop' } });
  await Promise.resolve();
  expect(container.querySelector('button')).toBeNull();
});

it('switches a plugin on for the site', async () => {
  const posts = serve([{ name: 'debug', description: 'Debug bar', on: false }]);
  const { findByLabelText, findByText, getByTitle } = render(PagePluginsMenu, { props: { site: 'shop' } });
  await fireEvent.click(await findByLabelText('Plugins on this site'));
  await findByText('Debug bar');
  await fireEvent.click(getByTitle('debug'));
  await waitFor(() => expect(posts).toEqual([{ site: 'shop', plugin: 'debug', kind: 'page_script', on: true }]));
});

// One plugin can bring both, and each is switched on its own.
it('groups page scripts apart from Overview cards and switches each kind on its own', async () => {
  const posts = serve([
    { name: 'stats', description: 'Request numbers', kind: 'page_script', on: true },
    { name: 'stats', description: 'Request numbers', kind: 'overview_card', on: false }
  ]);
  const { findByLabelText, findByText, getAllByTitle } = render(PagePluginsMenu, { props: { site: 'shop2' } });
  await fireEvent.click(await findByLabelText('Plugins on this site'));
  await findByText('On every page');
  await findByText('On its Overview');
  const [script, card] = getAllByTitle('stats');
  expect(script.getAttribute('aria-checked')).toBe('true');
  await fireEvent.click(card);
  await waitFor(() => expect(posts).toEqual([{ site: 'shop2', plugin: 'stats', kind: 'overview_card', on: true }]));
  await waitFor(() => expect(getAllByTitle('stats')[1].getAttribute('aria-checked')).toBe('true'));
});
