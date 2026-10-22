<script lang="ts">
  import Icon from './Icon.svelte';
  import { tooltip } from '$lib/tooltip';
  import { status } from '$stores/status';
  import { setStreamingMode } from '$stores/workspaces';
  import { m } from '../paraglide/messages.js';

  const on = $derived(Boolean($status.streaming_mode));
  const label = $derived(on ? m.nav_streamingOn() : m.nav_streamingOff());

  // The POST returns before the sites snapshot rebuilds, and turning it off
  // can take seconds, so stay busy until the status reports the new mode.
  let target = $state<boolean | null>(null);
  const busy = $derived(target !== null && target !== on);

  $effect(() => {
    if (target === on) target = null;
  });

  async function toggle() {
    if (busy) return;
    target = !on;
    const res = await setStreamingMode(target);
    if (res.ok) return;
    target = null;
    console.error('streaming mode failed:', res.error);
  }
</script>

{#if $status.streaming_enabled}
  <button
    type="button"
    onclick={toggle}
    aria-label={label}
    aria-pressed={on}
    aria-busy={busy}
    disabled={busy}
    use:tooltip={{ label, placement: 'bottom' }}
    class="inline-flex items-center justify-center w-5 h-5 rounded transition-colors {on
      ? 'text-lerd-red'
      : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-200'}"
  >
    {#if busy}
      <Icon name="spinner" class="w-3.5 h-3.5 animate-spin" />
    {:else}
      <Icon name="eyeOff" class="w-3.5 h-3.5" />
    {/if}
  </button>
{/if}
