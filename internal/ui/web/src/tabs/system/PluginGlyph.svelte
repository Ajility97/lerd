<script lang="ts" module>
  // The surfaces a plugin can add, drawn where the empty plugin views name them.
  // Kept here rather than in the app's icon set, which nothing else needs them in.
  const PATHS = {
    terminal: 'M4 5h16v14H4zM8 10l3 2-3 2M13 15h3',
    page: 'M4 4h16v16H4zM10 4v16M13 9h4M13 13h4',
    tab: 'M3 9h18v11H3zM3 9V5h7l2 4',
    card: 'M4 5h16v14H4zM8 15v-3M12 15V9M16 15v-5',
    sparkle: 'M12 3l1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8zM19 16v4M17 18h4',
    database: 'M4 6c0-1.7 3.6-3 8-3s8 1.3 8 3-3.6 3-8 3-8-1.3-8-3zM4 6v12c0 1.7 3.6 3 8 3s8-1.3 8-3V6M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3',
    code: 'M8 7l-5 5 5 5M16 7l5 5-5 5M14 4l-4 16',
    download: 'M12 3v12M7 10l5 5 5-5M5 21h14'
  } as const;
  export type GlyphName = keyof typeof PATHS;
</script>

<script lang="ts">
  interface Props {
    name: GlyphName;
    class?: string;
  }
  let { name, class: cls = 'w-4 h-4' }: Props = $props();
</script>

<svg class={cls} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
  <path d={PATHS[name]} />
</svg>
