<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import BackupCheckbox from '$components/BackupCheckbox.svelte';
  import { modal } from '$stores/modals';
  import { savePhpIni } from '$stores/phpVersions';
  import { notifyLocalFailure } from '$lib/notify';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.phpIniSave);
  let backup = $state(false);

  async function run(t: NonNullable<typeof target>) {
    const res = await savePhpIni(t.version, t.content, backup);
    if (res.ok) return;
    const error = res.error || m.nginxEditor_saveFailed();
    // The restart is the slow, failure-prone half, and the error otherwise
    // lives only in a modal the user is about to close.
    notifyLocalFailure('php_ini', m.phpIniEditor_restartFailed({ scope: t.label }), error);
    return error;
  }
</script>

<ConfigConfirmModal
  title={m.phpIniEditor_confirmTitle()}
  size="md"
  {target}
  acceptLabel={m.common_save()}
  busyLabel={m.phpIniEditor_restarting()}
  {run}
>
  {#snippet children(t, busy)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.phpIniEditor_confirmBody({ scope: t.label })}
    </p>
    {#if t.exists}
      <BackupCheckbox bind:checked={backup} disabled={busy} label={m.nginxEditor_backupLabel()} file="98-user.ini" />
    {/if}
  {/snippet}
</ConfigConfirmModal>
