<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { ConfigFile } from '$lib/configFile.svelte';
  import { m } from '../paraglide/messages.js';

  interface Props {
    file: ConfigFile;
    // The editor itself, bound to file.text.
    children: Snippet;
  }
  let { file, children }: Props = $props();
</script>

<div class="flex-1 min-h-0 overflow-hidden bg-gray-50 dark:bg-black/40">
  {#if file.loading}
    <p class="text-xs text-gray-400 px-3 py-2.5">{m.common_loading()}</p>
  {:else if file.error}
    <p class="text-xs text-red-500 dark:text-red-400 px-3 py-2.5">{file.error}</p>
  {:else}
    <div class="h-full min-h-64">
      {@render children()}
    </div>
  {/if}
</div>
