import { writable, derived } from 'svelte/store';

export type TabId = 'dashboard' | 'sites' | 'services' | 'system' | 'plugins';

// The rail's tabs. Plugins is a page too, reached from the rail's plugin group
// and the palette rather than a tab of its own.
export type RailTab = Exclude<TabId, 'plugins'>;
export const TABS: RailTab[] = ['dashboard', 'sites', 'services', 'system'];
const PAGES: TabId[] = [...TABS, 'plugins'];

// Plugins lived under System before it had a page; its old links still land.
const LEGACY_PLUGINS = 'system/plugins';

export function parseHash(hash: string): { tab: TabId; rest: string } {
  let h = hash.startsWith('#') ? hash.slice(1) : hash;
  if (h === LEGACY_PLUGINS || h.startsWith(LEGACY_PLUGINS + '/')) h = h.slice('system/'.length);
  for (const t of PAGES) {
    if (h === t) return { tab: t, rest: '' };
    if (h.startsWith(t + '/')) return { tab: t, rest: h.slice(t.length + 1) };
  }
  return { tab: 'dashboard', rest: '' };
}

const initial = parseHash(location.hash);
export const tab = writable<TabId>(initial.tab);
export const routeRest = writable<string>(initial.rest);

export const route = derived([tab, routeRest], ([$tab, $rest]) => ({ tab: $tab, rest: $rest }));

window.addEventListener('hashchange', () => {
  const p = parseHash(location.hash);
  tab.set(p.tab);
  routeRest.set(p.rest);
});

export function goToTab(t: TabId, rest = '') {
  const h = rest ? `${t}/${rest}` : t;
  location.hash = h;
}
