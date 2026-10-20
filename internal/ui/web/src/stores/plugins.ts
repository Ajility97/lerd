import { writable, derived } from 'svelte/store';
import { apiFetch, apiJson } from '$lib/api';

export interface PluginPage {
  title: string;
  url: string;
  icon?: string;
}

export interface PluginInfo {
  name: string;
  description?: string;
  page?: PluginPage;
  site_tab?: PluginPage;
  overview_card?: PluginPage;
}

export const plugins = writable<PluginInfo[]>([]);

// Plugins that draw a page on the rail, and those that add a tab to every site.
export const railPlugins = derived(plugins, ($p) => $p.filter((p) => p.page));
export const siteTabPlugins = derived(plugins, ($p) => $p.filter((p) => p.site_tab));

export async function loadPlugins() {
  // A lerd-ui from before plugins answers this path with something else.
  try {
    const list = await apiJson<PluginInfo[]>('/api/plugins');
    plugins.set(Array.isArray(list) ? list : []);
  } catch {
    plugins.set([]);
  }
}

// One thing a plugin can put on a site, its page script or its Overview card,
// and whether the site has it switched on.
export type PagePluginKind = 'page_script' | 'overview_card';
export interface PagePlugin {
  name: string;
  description?: string;
  kind: PagePluginKind;
  on: boolean;
}

// The switches per site name, shared so the Overview shows a card the moment
// the header's menu switches it on.
export const siteSwitches = writable<Record<string, PagePlugin[]>>({});
const keep = (site: string, list: PagePlugin[]) => {
  siteSwitches.update((all) => ({ ...all, [site]: list }));
  return list;
};

export async function loadPagePlugins(site: string): Promise<PagePlugin[]> {
  const list = await apiJson<PagePlugin[]>('/api/page-plugins?site=' + encodeURIComponent(site));
  // A lerd-ui from before Overview cards sends no kind: those were page scripts.
  return keep(site, Array.isArray(list) ? list.map((p) => ({ ...p, kind: p.kind ?? 'page_script' })) : []);
}

// setPagePlugin resolves to the site's switches after this one, or throws
// with lerd's reason.
export async function setPagePlugin(site: string, plugin: string, kind: PagePluginKind, on: boolean): Promise<PagePlugin[]> {
  const res = await apiFetch('/api/page-plugins', {
    method: 'POST',
    body: JSON.stringify({ site, plugin, kind, on }),
    headers: { 'Content-Type': 'application/json' }
  });
  if (!res.ok) throw new Error((await res.text()).trim() || res.statusText);
  return keep(site, await res.json());
}

// A row on the plugin manager. A plugin that failed to load has an error and
// nothing else, so it can still be removed.
export interface ManagedPlugin {
  name: string;
  version?: string;
  description?: string;
  adds?: string[];
  requires?: string[];
  error?: string;
  // Its rail icon, when it has a page.
  icon?: string;
  // Where the catalog installed it from; absent when not listed.
  repo?: string;
  commit?: string;
  // A newer approved commit, when there is one.
  update?: string;
  // Its latest failure, with a filled-in issue on its repository when known.
  failure?: PluginFailure;
}

export interface PluginFailure {
  surface: string;
  what: string;
  error: string;
  at: string;
  report_url?: string;
}

// An approved plugin in the catalog, as the Browse view draws it.
export interface CatalogEntry {
  name: string;
  description?: string;
  repo: string;
  commit: string;
  author?: string;
  category?: string;
  tags?: string[];
  images?: string[];
  lerd?: string;
  requires?: string[];
  platforms?: string[];
  added_at?: string;
  updated_at?: string;
  stars?: number;
  installed: boolean;
  compatible: boolean;
  runs_here?: boolean;
  missing?: string[];
}

// loadCatalog throws with lerd's reason when the catalog cannot be reached.
export async function loadCatalog(): Promise<CatalogEntry[]> {
  const res = await apiFetch('/api/plugin-manager/catalog');
  if (!res.ok) throw new Error((await res.text()).trim() || res.statusText);
  const list = await res.json();
  return Array.isArray(list) ? list : [];
}

export async function installListedPlugin(name: string): Promise<string> {
  return mutate('/api/plugin-manager', {
    method: 'POST',
    body: JSON.stringify({ name }),
    headers: { 'Content-Type': 'application/json' }
  });
}

export async function updatePlugin(name: string): Promise<string> {
  return mutate('/api/plugin-manager/' + encodeURIComponent(name) + '/update', { method: 'POST' });
}

export const managedPlugins = writable<ManagedPlugin[]>([]);

export async function loadManagedPlugins() {
  const list = await apiJson<ManagedPlugin[]>('/api/plugin-manager');
  managedPlugins.set(Array.isArray(list) ? list : []);
}

// installPlugin and removePlugin resolve to an error message, or '' on success,
// and refresh both lists so the rail and site tabs follow at once.
export async function installPlugin(source: string): Promise<string> {
  return mutate('/api/plugin-manager', {
    method: 'POST',
    body: JSON.stringify({ source }),
    headers: { 'Content-Type': 'application/json' }
  });
}

export async function removePlugin(name: string): Promise<string> {
  return mutate('/api/plugin-manager/' + encodeURIComponent(name), { method: 'DELETE' });
}

async function mutate(path: string, init: RequestInit): Promise<string> {
  const res = await apiFetch(path, init);
  const error = res.ok ? '' : (await res.text()).trim() || res.statusText;
  await Promise.all([loadManagedPlugins(), loadPlugins()]);
  return error;
}
