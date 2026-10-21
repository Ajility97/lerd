<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import BackupDiff from '$components/BackupDiff.svelte';
  import { modal } from '$stores/modals';
  import { restoreServiceTuning } from '$stores/services';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.tuningRestore);
  let rolledBack = $state(false);

  async function run(t: NonNullable<typeof target>) {
    rolledBack = false;
    // Pass the exact backup previewed in the diff, so a save landing a newer
    // backup since the modal opened cannot swap a different file into place.
    const res = await restoreServiceTuning(t.name, t.backupName);
    if (res.ok) return;
    rolledBack = res.rolledBack ?? false;
    return res.error || m.tuningEditor_restoreFailed();
  }
</script>

<ConfigConfirmModal
  title={m.tuningEditor_restoreTitle()}
  size="lg"
  {target}
  acceptLabel={m.tuningEditor_restoreAccept()}
  busyLabel={m.tuningEditor_restoring()}
  {run}
>
  {#snippet children(t)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.tuningEditor_restoreBody({ name: t.name, backup: t.backupName })}
    </p>
    <BackupDiff current={t.current} backup={t.backup} emptyText={m.tuningEditor_restoreNoDiff()} />
  {/snippet}
  {#snippet notice()}
    {#if rolledBack}
      <p class="text-xs text-emerald-600 dark:text-emerald-400">{m.tuningEditor_rolledBack()}</p>
    {/if}
  {/snippet}
</ConfigConfirmModal>
