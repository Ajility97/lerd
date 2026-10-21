<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import BackupDiff from '$components/BackupDiff.svelte';
  import { modal } from '$stores/modals';
  import { restorePhpIni } from '$stores/phpVersions';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.phpIniRestore);

  async function run(t: NonNullable<typeof target>) {
    const res = await restorePhpIni(t.version, t.backupName);
    if (!res.ok) return res.error || m.nginxEditor_restoreFailed();
  }
</script>

<ConfigConfirmModal
  title={m.phpIniEditor_restoreTitle()}
  size="lg"
  {target}
  acceptLabel={m.nginxEditor_restoreAccept()}
  busyLabel={m.nginxEditor_restoring()}
  {run}
>
  {#snippet children(t)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.phpIniEditor_restoreBody({ scope: t.label, name: t.backupName })}
    </p>
    <BackupDiff current={t.current} backup={t.backup} emptyText={m.nginxEditor_restoreNoDiff()} />
  {/snippet}
</ConfigConfirmModal>
