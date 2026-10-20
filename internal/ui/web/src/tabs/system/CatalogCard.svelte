<script lang="ts">
  import Icon from '$components/Icon.svelte';
  import type { CatalogEntry } from '$stores/plugins';
  import { m } from '../../paraglide/messages.js';

  // A catalog listing as a card: its first screenshot, what it is and who made
  // it. The whole card opens the detail, where it is installed from.
  interface Props {
    entry: CatalogEntry;
    focused?: boolean;
    onopen: () => void;
  }
  let { entry: e, focused = false, onopen }: Props = $props();
</script>

<button
  type="button"
  id="plugin-card-{e.name}"
  onclick={onopen}
  class="group flex flex-col text-left rounded-xl border overflow-hidden bg-white dark:bg-lerd-card transition duration-150 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-black/5 {focused
    ? 'border-lerd-red ring-1 ring-lerd-red'
    : 'border-gray-200/80 dark:border-lerd-border hover:border-gray-300 dark:hover:border-white/15'}"
>
  <div class="relative aspect-[16/10] overflow-hidden border-b border-gray-200/80 dark:border-lerd-border bg-gray-100 dark:bg-lerd-bg">
    {#if e.images?.length}
      <img src={e.images[0]} alt="" loading="lazy" class="absolute left-4 top-4 w-[calc(100%-1rem)] h-[calc(100%-1rem)] object-cover object-left-top rounded-tl-md border-l border-t border-gray-200 dark:border-white/10 shadow-xl" />
    {:else}
      <span class="absolute inset-0 flex items-center justify-center text-gray-300 dark:text-gray-600"><Icon name="cube" class="w-8 h-8" /></span>
    {/if}
  </div>
  <div class="flex-1 flex flex-col gap-1 p-3">
    <div class="flex items-baseline justify-between gap-2">
      <span class="truncate text-sm font-semibold text-gray-900 dark:text-white">{e.name}</span>
      {#if e.stars}<span class="shrink-0 text-[10px] tabular-nums text-gray-400">★ {e.stars.toLocaleString()}</span>{/if}
    </div>
    {#if e.description}<p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{e.description}</p>{/if}
    <div class="mt-auto pt-2 flex items-center justify-between gap-2 text-[10px] text-gray-400">
      <span class="truncate">{e.author ? m.system_plugins_by({ author: e.author }) : ''}</span>
      {#if e.installed}
        <span class="shrink-0 font-medium text-emerald-600 dark:text-emerald-400">{m.system_plugins_installedBadge()}</span>
      {:else if e.category}
        <span class="shrink-0">{e.category}</span>
      {/if}
    </div>
  </div>
</button>
