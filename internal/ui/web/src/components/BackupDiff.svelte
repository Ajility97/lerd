<script lang="ts">
  import { diffLines } from '$lib/diff';

  interface Props {
    current: string;
    backup: string;
    emptyText: string;
  }
  let { current, backup, emptyText }: Props = $props();

  const lines = $derived(diffLines(current, backup));
</script>

<div class="border border-gray-200 dark:border-lerd-border rounded-sm bg-gray-50 dark:bg-black/40 overflow-auto max-h-96">
  {#if lines.length === 0}
    <p class="text-xs text-gray-400 px-3 py-2">{emptyText}</p>
  {:else}
    <pre class="text-[11px] leading-relaxed font-mono"><!--
      -->{#each lines as l, i (i)}<!--
        -->{#if l.op === '+'}<!--
          --><span class="block px-3 bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-300">+ {l.line}</span><!--
        -->{:else if l.op === '-'}<!--
          --><span class="block px-3 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300">- {l.line}</span><!--
        -->{:else}<!--
          --><span class="block px-3 text-gray-600 dark:text-gray-400">  {l.line}</span><!--
        -->{/if}<!--
      -->{/each}<!--
    --></pre>
  {/if}
</div>
