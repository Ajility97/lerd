<script lang="ts">
  // The empty plugin views' illustrations, drawn as the dashboard itself: one
  // chrome for the header and rail, the lighter content inset beside it, in the
  // theme's own colours so the picture follows whatever theme is on.
  interface Props {
    kind: 'installed' | 'catalog';
  }
  let { kind }: Props = $props();
  const clip = $props.id();

  const CHROME = 'fill-lerd-header-light dark:fill-lerd-header';
  const CONTENT = 'fill-gray-50 dark:fill-lerd-bg';
  const CARD = 'fill-white dark:fill-lerd-card stroke-gray-200 dark:stroke-lerd-border';
  const FAINT = 'fill-gray-300 dark:fill-gray-600';
  const SLOT = 'stroke-lerd-red';
</script>

{#snippet frame()}
  <defs><clipPath id={clip}><rect x="36" y="14" width="196" height="122" rx="12" /></clipPath></defs>
  <g clip-path="url(#{clip})">
    <rect x="36" y="14" width="196" height="122" class={CONTENT} />
    <rect x="36" y="14" width="196" height="18" class={CHROME} />
    <rect x="36" y="14" width="26" height="122" class={CHROME} />
    <path d="M62 32h170" class="stroke-gray-200 dark:stroke-lerd-border" />
  </g>
  <rect x="36" y="14" width="196" height="122" rx="12" class="stroke-gray-200 dark:stroke-lerd-border" stroke-width="1.5" />
  <rect x="44" y="19" width="10" height="9" rx="2.5" fill="#ff2d20" />
  <rect x="68" y="20.5" width="34" height="5" rx="2.5" class={FAINT} />
  <rect x="45" y="42" width="8" height="8" rx="2" class={FAINT} />
  <rect x="45" y="56" width="8" height="8" rx="2" class={FAINT} />
{/snippet}

{#if kind === 'installed'}
  <svg width="268" height="150" viewBox="0 0 268 150" fill="none" aria-hidden="true">
    {@render frame()}
    <rect x="44.5" y="70" width="9" height="9" rx="2" class={SLOT} stroke-width="1.4" stroke-dasharray="2 2" />
    <rect x="70" y="40" width="152" height="26" rx="6" class={CARD} />
    <rect x="70" y="74" width="72" height="50" rx="6" class={CARD} />
    <rect x="150" y="74" width="72" height="50" rx="6" class={SLOT} stroke-width="1.5" stroke-dasharray="4 3" />
    <g transform="translate(170 86) rotate(-8)">
      <rect width="56" height="44" rx="9" class="fill-lerd-red" />
      <path d="M28 10 17 16v12l11 6 11-6V16zM17 16l11 6 11-6M28 22v12" stroke="#fff" stroke-width="1.8" stroke-linejoin="round" />
    </g>
    <g transform="translate(10 96) rotate(10)">
      <rect width="36" height="30" rx="7" class={CARD} stroke-width="1.5" />
      <path d="m10 12-4 3 4 3M26 12l4 3-4 3M20 9l-4 12" class="stroke-lerd-red" stroke-width="1.6" stroke-linecap="round" />
    </g>
    <g transform="translate(220 4) rotate(12)">
      <rect width="34" height="30" rx="7" class={CARD} stroke-width="1.5" />
      <path d="M17 7l1.6 4.4L23 13l-4.4 1.6L17 19l-1.6-4.4L11 13l4.4-1.6z" class="fill-lerd-red" />
    </g>
  </svg>
{:else}
  <svg width="268" height="150" viewBox="0 0 268 150" fill="none" aria-hidden="true">
    {@render frame()}
    <rect x="45" y="70" width="8" height="8" rx="2" class={FAINT} />
    <rect x="70" y="40" width="70" height="5" rx="2.5" class={FAINT} />
    <rect x="70" y="54" width="46" height="70" rx="6" class="stroke-gray-300 dark:stroke-lerd-border" stroke-width="1.5" stroke-dasharray="4 4" />
    <rect x="123" y="54" width="46" height="70" rx="6" class="stroke-gray-300 dark:stroke-lerd-border" stroke-width="1.5" stroke-dasharray="4 4" />
    <rect x="176" y="54" width="46" height="70" rx="6" class="fill-lerd-red/10 {SLOT}" stroke-width="1.5" stroke-dasharray="4 4" />
    <path d="M199 79v20M189 89h20" class="stroke-lerd-red" stroke-width="2.2" stroke-linecap="round" />
    <g transform="translate(206 104) rotate(-10)">
      <rect width="48" height="38" rx="8" class="fill-lerd-red" />
      <path d="M24 9 15 14v10l9 5 9-5V14zM15 14l9 5 9-5M24 19v10" stroke="#fff" stroke-width="1.7" stroke-linejoin="round" />
    </g>
    <g transform="translate(222 2) rotate(12)">
      <rect width="36" height="30" rx="7" class={CARD} stroke-width="1.5" />
      <path d="M18 7l1.6 4.4L24 13l-4.4 1.6L18 19l-1.6-4.4L12 13l4.4-1.6z" class="fill-lerd-red" />
    </g>
  </svg>
{/if}
