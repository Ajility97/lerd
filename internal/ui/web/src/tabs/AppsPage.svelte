<script lang="ts">
  import { dashboardServices, openDashboard } from '$stores/dashboard';
  import ServiceIcon from '../components/ServiceIcon.svelte';
  import { serviceLabel } from '$stores/services';
  import { accessMode } from '$stores/accessMode';
  import { goToTab } from '$stores/route';
  import Icon from '../components/Icon.svelte';
  import { m } from '../paraglide/messages.js';
</script>

{#snippet tile(label: string, onclick: () => void, glyph: import('svelte').Snippet)}
  <button
    {onclick}
    class="flex flex-col items-center justify-center gap-2 aspect-square rounded-xl border border-gray-200 dark:border-lerd-border bg-white dark:bg-lerd-card hover:border-lerd-red/40 hover:bg-lerd-red/5 transition-colors p-3"
  >
    {@render glyph()}
    <span class="text-[11px] font-medium text-gray-700 dark:text-gray-300 text-center truncate max-w-full">{label}</span>
  </button>
{/snippet}
{#snippet puzzle()}<Icon name="puzzle" class="w-6 h-6 text-gray-600 dark:text-gray-300" />{/snippet}

<div class="flex-1 overflow-y-auto p-4">
  {#if $dashboardServices.length === 0 && !$accessMode.localControl}
    <div class="text-center text-sm text-gray-500 dark:text-gray-400 py-16">
      {m.apps_empty()}
    </div>
  {:else}
    <div class="grid grid-cols-3 gap-3">
      {#each $dashboardServices as svc (svc.name)}
        {#snippet mark()}<ServiceIcon name={svc.name} category={svc.category} icon={svc.icon} color={svc.color} preset={svc.preset} bare />{/snippet}
        {@render tile(serviceLabel(svc.name), () => openDashboard(svc), mark)}
      {/each}
      <!-- The phone's way to the Plugins page, which the rail offers on a desktop. -->
      {#if $accessMode.localControl}
        {@render tile(m.system_plugins_title(), () => goToTab('plugins'), puzzle)}
      {/if}
    </div>
  {/if}
</div>
