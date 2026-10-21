<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import BackupCheckbox from '$components/BackupCheckbox.svelte';
  import { modal } from '$stores/modals';
  import { saveSiteEnv } from '$stores/sites';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.envSave);
  let backup = $state(false);

  async function run(t: NonNullable<typeof target>) {
    const res = await saveSiteEnv(t.domain, t.branch, t.content, backup, t.file);
    if (!res.ok) return res.error || m.envEditor_saveFailed();
  }
</script>

<ConfigConfirmModal
  title={m.envEditor_confirmTitle()}
  size="md"
  {target}
  acceptLabel={m.common_save()}
  busyLabel={m.envEditor_saving()}
  {run}
>
  {#snippet children(t, busy)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.envEditor_confirmBody({ domain: t.domain, file: t.file })}
    </p>
    <BackupCheckbox bind:checked={backup} disabled={busy} label={m.envEditor_backupLabel()} file={t.file} />
  {/snippet}
</ConfigConfirmModal>
