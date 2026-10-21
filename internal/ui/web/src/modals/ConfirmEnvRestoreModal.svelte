<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import BackupDiff from '$components/BackupDiff.svelte';
  import { modal } from '$stores/modals';
  import { restoreSiteEnv } from '$stores/sites';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.envRestore);

  async function run(t: NonNullable<typeof target>) {
    // Pass the exact backup previewed in the diff, so a save landing a newer
    // backup since the modal opened cannot swap a different file into place.
    const res = await restoreSiteEnv(t.domain, t.branch, t.file, t.backupName);
    if (!res.ok) return res.error || m.envEditor_restoreFailed();
  }
</script>

<ConfigConfirmModal
  title={m.envEditor_restoreTitle()}
  size="lg"
  {target}
  acceptLabel={m.envEditor_restoreAccept()}
  busyLabel={m.envEditor_restoring()}
  {run}
>
  {#snippet children(t)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.envEditor_restoreBody({ file: t.file, name: t.backupName })}
    </p>
    <BackupDiff current={t.current} backup={t.backup} emptyText={m.envEditor_restoreNoDiff()} />
  {/snippet}
</ConfigConfirmModal>
