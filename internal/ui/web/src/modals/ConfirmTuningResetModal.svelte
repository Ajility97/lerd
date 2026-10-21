<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import { modal } from '$stores/modals';
  import { resetServiceTuning } from '$stores/services';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.tuningReset);
  let rolledBack = $state(false);
  let autoBackupName = $state('');

  async function run(t: NonNullable<typeof target>) {
    rolledBack = false;
    autoBackupName = '';
    const res = await resetServiceTuning(t.name);
    if (res.ok) return;
    rolledBack = res.rolledBack ?? false;
    autoBackupName = res.autoBackupName ?? '';
    return res.error || m.tuningEditor_resetFailed();
  }
</script>

<ConfigConfirmModal
  title={m.tuningEditor_resetTitle()}
  size="md"
  {target}
  acceptLabel={m.tuningEditor_resetAccept()}
  busyLabel={m.tuningEditor_resetting()}
  {run}
>
  {#snippet children(t)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.tuningEditor_resetBody({ name: t.name })}
    </p>
    <p class="text-[10px] text-gray-400 dark:text-gray-600 font-mono break-all">{t.path}</p>
  {/snippet}
  {#snippet notice()}
    {#if rolledBack}
      <p class="text-xs text-emerald-600 dark:text-emerald-400">
        {autoBackupName ? m.tuningEditor_resetRolledBack({ name: autoBackupName }) : m.tuningEditor_rolledBack()}
      </p>
    {/if}
  {/snippet}
</ConfigConfirmModal>
