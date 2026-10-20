<script lang="ts">
  import Icon from './Icon.svelte';
  import { tooltip } from '$lib/tooltip';
  import { m } from '../paraglide/messages.js';

  // Which slice of a long list is on screen, and the two steps either side of
  // it. The caller slices; it draws nothing when everything fits on one page.
  interface Props {
    page: number;
    total: number;
    size: number;
    onchange: (page: number) => void;
  }
  let { page, total, size, onchange }: Props = $props();

  const last = $derived(Math.max(0, Math.ceil(total / size) - 1));
  const BUTTON =
    'flex items-center justify-center w-7 h-7 rounded-md text-gray-500 dark:text-gray-400 enabled:hover:text-gray-700 dark:enabled:hover:text-gray-200 enabled:hover:bg-gray-100 dark:enabled:hover:bg-white/5 disabled:opacity-40 transition-colors';
</script>

{#if total > size}
  <div class="flex items-center justify-between gap-2 pt-3 text-xs text-gray-500 dark:text-gray-400">
    <span class="tabular-nums">
      {m.common_pageRange({ from: page * size + 1, to: Math.min(total, (page + 1) * size), total })}
    </span>
    <div class="flex items-center gap-1">
      <button type="button" onclick={() => onchange(Math.max(0, page - 1))} disabled={page === 0} use:tooltip={m.common_pagePrev()} aria-label={m.common_pagePrev()} class={BUTTON}>
        <Icon name="back" class="w-3.5 h-3.5" />
      </button>
      <button type="button" onclick={() => onchange(Math.min(last, page + 1))} disabled={page >= last} use:tooltip={m.common_pageNext()} aria-label={m.common_pageNext()} class={BUTTON}>
        <Icon name="chevron" class="w-3.5 h-3.5 -rotate-90" />
      </button>
    </div>
  </div>
{/if}
