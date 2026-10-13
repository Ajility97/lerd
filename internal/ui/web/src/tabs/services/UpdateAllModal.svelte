<script lang="ts">
  import ConfirmModal from '$components/ConfirmModal.svelte';
  import ServiceIcon from '$components/ServiceIcon.svelte';
  import Toggle from '$components/Toggle.svelte';
  import {
    serviceLabel,
    updateServices,
    updateProgress,
    type Service,
    type ServiceUpdateResult
  } from '$stores/services';
  import { m } from '../../paraglide/messages.js';

  interface Props {
    // The services that had an update when the dialog opened. Kept fixed so a
    // row does not vanish the moment its own update clears the badge.
    services: Service[];
    onclose: () => void;
    ondone: (results: ServiceUpdateResult[]) => void;
  }
  let { services, onclose, ondone }: Props = $props();

  let on = $state<Record<string, boolean>>(Object.fromEntries(services.map((s) => [s.name, true])));
  let updating = $state(false);
  const selected = $derived(services.filter((s) => on[s.name]).map((s) => s.name));

  function target(s: Service): string {
    if (!s.latest_version || s.latest_version === s.version) return s.version ?? '';
    return s.version ? s.version + ' → ' + s.latest_version : '→ ' + s.latest_version;
  }

  async function run() {
    updating = true;
    try {
      ondone(await updateServices(selected));
    } finally {
      updating = false;
      onclose();
    }
  }
</script>

<ConfirmModal
  open
  title={m.services_updateAll()}
  body={m.services_updateAll_body()}
  confirmLabel={m.services_update()}
  loading={updating}
  disabled={selected.length === 0}
  onconfirm={run}
  {onclose}
>
  {#snippet extra()}
    <div class="rounded-lg border border-gray-200 dark:border-lerd-border divide-y divide-gray-100 dark:divide-lerd-border">
      {#each services as s (s.name)}
        {@const progress = $updateProgress[s.name]}
        <div class="flex items-center gap-3 px-3 py-2.5">
          <ServiceIcon name={s.name} category={s.category} icon={s.icon} color={s.color} preset={s.preset} compact />
          <div class="flex-1 min-w-0">
            <div class="text-sm text-gray-800 dark:text-gray-100 truncate">{serviceLabel(s.name)}</div>
            <div class="truncate text-xs {progress?.error ? 'text-red-500' : progress ? 'text-gray-500 dark:text-gray-400' : 'font-mono text-emerald-600 dark:text-emerald-400'}">
              {progress ? progress.message : target(s)}
            </div>
          </div>
          <Toggle on={on[s.name]} title={serviceLabel(s.name)} disabled={updating} onclick={() => (on[s.name] = !on[s.name])} />
        </div>
      {/each}
    </div>
  {/snippet}
</ConfirmModal>
