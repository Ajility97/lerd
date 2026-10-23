<script lang="ts">
  import type { Snippet } from 'svelte';
  import PluginGlyph, { type GlyphName } from './PluginGlyph.svelte';

  // A slim invitation under a list that already has plugins: to the catalog
  // from the installed ones, to building one from the catalog.
  interface Props {
    glyph: GlyphName;
    title: string;
    body: string;
    children: Snippet;
  }
  let { glyph, title, body, children }: Props = $props();
</script>

<section class="flex flex-col sm:flex-row sm:items-center gap-3 rounded-xl border border-gray-200/80 dark:border-lerd-border bg-white dark:bg-lerd-card p-4">
  <div class="flex items-start gap-3 min-w-0 flex-1">
    <span class="shrink-0 flex items-center justify-center w-9 h-9 rounded-lg bg-lerd-red/10 text-lerd-red"><PluginGlyph name={glyph} class="w-4.5 h-4.5" /></span>
    <div class="min-w-0">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{title}</h3>
      <p class="text-xs text-gray-500 dark:text-gray-400">{body}</p>
    </div>
  </div>
  <div class="shrink-0 flex flex-wrap items-center gap-2">{@render children()}</div>
</section>
