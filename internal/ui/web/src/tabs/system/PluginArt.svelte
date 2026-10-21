<script lang="ts">
  // The two illustrations of the empty plugin views: the dashboard with a plugin
  // snapping in (nothing installed) and a shelf with an open slot (an empty
  // catalog). Drawn in the theme's own colours, so they follow its accent.
  interface Props {
    kind: 'installed' | 'catalog';
  }
  let { kind }: Props = $props();

  const CARD = 'fill-white dark:fill-lerd-card stroke-gray-200 dark:stroke-lerd-border';
  const FAINT = 'fill-gray-300 dark:fill-gray-600';
</script>

{#if kind === 'installed'}
  <svg width="260" height="150" viewBox="0 0 260 150" fill="none" aria-hidden="true">
    <rect x="40" y="18" width="180" height="114" rx="12" class={CARD} stroke-width="1.5" />
    <rect x="40" y="18" width="30" height="114" rx="12" class="fill-gray-100 dark:fill-white/5" />
    <circle cx="55" cy="38" r="6" class="fill-lerd-red" />
    <rect x="51" y="54" width="8" height="8" rx="2" class={FAINT} />
    <rect x="51" y="70" width="8" height="8" rx="2" class={FAINT} />
    <rect x="51" y="88" width="8" height="8" rx="2" class="stroke-lerd-red" stroke-width="1.5" stroke-dasharray="2 2" />
    <rect x="82" y="32" width="70" height="7" rx="3.5" class={FAINT} />
    <rect x="82" y="48" width="124" height="30" rx="7" class="fill-gray-50 dark:fill-lerd-bg stroke-gray-200 dark:stroke-lerd-border" />
    <rect x="82" y="86" width="58" height="34" rx="7" class="fill-gray-50 dark:fill-lerd-bg stroke-gray-200 dark:stroke-lerd-border" />
    <rect x="148" y="86" width="58" height="34" rx="7" class="stroke-lerd-red" stroke-width="1.5" stroke-dasharray="4 3" />
    <g transform="translate(166 66) rotate(-8)">
      <rect width="56" height="44" rx="9" class="fill-lerd-red" />
      <path d="M28 10 17 16v12l11 6 11-6V16zM17 16l11 6 11-6M28 22v12" stroke="#fff" stroke-width="1.8" stroke-linejoin="round" />
    </g>
    <g transform="translate(14 98) rotate(10)">
      <rect width="36" height="30" rx="7" class={CARD} stroke-width="1.5" />
      <path d="m10 12-4 3 4 3M26 12l4 3-4 3M20 9l-4 12" class="stroke-lerd-red" stroke-width="1.6" stroke-linecap="round" />
    </g>
    <g transform="translate(206 8) rotate(12)">
      <rect width="34" height="30" rx="7" class={CARD} stroke-width="1.5" />
      <path d="M17 7l1.6 4.4L23 13l-4.4 1.6L17 19l-1.6-4.4L11 13l4.4-1.6z" class="fill-lerd-red" />
    </g>
  </svg>
{:else}
  <svg width="280" height="150" viewBox="0 0 280 150" fill="none" aria-hidden="true">
    <rect x="30" y="20" width="220" height="112" rx="12" class={CARD} stroke-width="1.5" />
    <rect x="46" y="34" width="90" height="8" rx="4" class={FAINT} />
    <rect x="46" y="54" width="56" height="62" rx="8" class="stroke-gray-300 dark:stroke-lerd-border" stroke-width="1.5" stroke-dasharray="4 4" />
    <rect x="112" y="54" width="56" height="62" rx="8" class="stroke-gray-300 dark:stroke-lerd-border" stroke-width="1.5" stroke-dasharray="4 4" />
    <rect x="178" y="54" width="56" height="62" rx="8" class="fill-lerd-red/10 stroke-lerd-red" stroke-width="1.5" stroke-dasharray="4 4" />
    <path d="M206 75v20M196 85h20" class="stroke-lerd-red" stroke-width="2.2" stroke-linecap="round" />
    <g transform="translate(214 100) rotate(-10)">
      <rect width="48" height="38" rx="8" class="fill-lerd-red" />
      <path d="M24 9 15 14v10l9 5 9-5V14zM15 14l9 5 9-5M24 19v10" stroke="#fff" stroke-width="1.7" stroke-linejoin="round" />
    </g>
    <g transform="translate(8 12) rotate(-12)">
      <rect width="36" height="30" rx="7" class={CARD} stroke-width="1.5" />
      <path d="M18 7l1.6 4.4L24 13l-4.4 1.6L18 19l-1.6-4.4L12 13l4.4-1.6z" class="fill-lerd-red" />
    </g>
  </svg>
{/if}
