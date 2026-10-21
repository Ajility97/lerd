<script lang="ts">
  import { onMount } from 'svelte';
  import Popover from './Popover.svelte';
  import Toggle from './Toggle.svelte';
  import Icon from './Icon.svelte';
  import { tooltip } from '$lib/tooltip';
  import { notifyLocalFailure } from '$lib/notify';
  import { loadPagePlugins, setPagePlugin, siteSwitches, type PagePlugin } from '$stores/plugins';
  import { m } from '../paraglide/messages.js';

  // Switches what plugins put on one site: scripts on its pages and cards on
  // its Overview. Draws nothing until an installed plugin has either, so the
  // header stays as it was.

  interface Props {
    site: string;
    class?: string;
  }
  let { site, class: cls = '' }: Props = $props();

  let busy = $state('');
  const list = $derived($siteSwitches[site] ?? []);
  const anyOn = $derived(list.some((p) => p.on));
  const groups = $derived(
    [
      { title: m.sites_pagePlugins_scripts(), hint: m.sites_pagePlugins_scriptsHint(), items: list.filter((p) => p.kind === 'page_script') },
      { title: m.sites_pagePlugins_cards(), hint: '', items: list.filter((p) => p.kind === 'overview_card') }
    ].filter((g) => g.items.length > 0)
  );

  onMount(() => {
    loadPagePlugins(site).catch(() => {});
  });

  async function flip(p: PagePlugin) {
    busy = p.kind + ':' + p.name;
    try {
      await setPagePlugin(site, p.name, p.kind, !p.on);
    } catch (err) {
      notifyLocalFailure('page_plugin', m.sites_pagePlugins_label(), err instanceof Error ? err.message : String(err));
    } finally {
      busy = '';
    }
  }
</script>

{#if list.length > 0}
  <div class="{cls} items-center">
    <Popover label={m.sites_pagePlugins_label()} align="right" width={320}>
      {#snippet triggerButton(toggle, open)}
        <button
          type="button"
          onclick={() => {
            // Read again on every open, so a switch made in another dashboard shows.
            if (!open) loadPagePlugins(site).catch(() => {});
            toggle();
          }}
          aria-label={m.sites_pagePlugins_label()}
          aria-expanded={open}
          use:tooltip={m.sites_pagePlugins_label()}
          class="w-8 h-8 flex items-center justify-center rounded-md transition-colors {anyOn
            ? 'text-emerald-500 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/20'
            : 'text-gray-500 dark:text-gray-400 hover:text-lerd-red hover:bg-gray-100 dark:hover:bg-white/5'}"
        >
          <Icon name="puzzle" class="w-4 h-4" />
        </button>
      {/snippet}
      {#snippet children()}
        <div class="py-1.5 divide-y divide-gray-100 dark:divide-lerd-border">
          {#each groups as g (g.title)}
            <section class="px-3 py-2">
              <h4 class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">{g.title}</h4>
              {#if g.hint}<p class="mt-0.5 text-[11px] leading-snug text-gray-500 dark:text-gray-400">{g.hint}</p>{/if}
              <ul class="mt-1.5 space-y-0.5">
                {#each g.items as p (p.name)}
                  <li class="flex items-start gap-3 rounded-md py-1.5">
                    <div class="flex-1 min-w-0">
                      <div class="text-sm font-medium text-gray-900 dark:text-white">{p.name}</div>
                      {#if p.description}<div class="text-xs leading-snug text-gray-500 dark:text-gray-400 line-clamp-2">{p.description}</div>{/if}
                    </div>
                    <span class="pt-1">
                      <Toggle on={p.on} loading={busy === p.kind + ':' + p.name} disabled={busy !== ''} title={p.name} onclick={() => flip(p)} />
                    </span>
                  </li>
                {/each}
              </ul>
            </section>
          {/each}
        </div>
      {/snippet}
    </Popover>
  </div>
{/if}
