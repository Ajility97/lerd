<script lang="ts">
  import IconButton from './IconButton.svelte';
  import type { TooltipPlacement } from '$lib/tooltip';

  interface Props {
    title: string;
    active: boolean;
    onclick?: (e: MouseEvent) => void;
    placement?: TooltipPlacement;
  }
  let { title, active, onclick, placement }: Props = $props();
</script>

<IconButton {title} {active} {onclick} {placement}>X</IconButton>
