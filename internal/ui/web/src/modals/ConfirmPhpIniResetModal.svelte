<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import { modal } from '$stores/modals';
  import { resetPhpIni } from '$stores/phpVersions';
  import { notifyLocalFailure } from '$lib/notify';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.phpIniReset);

  async function run(t: NonNullable<typeof target>) {
    const res = await resetPhpIni(t.version);
    if (res.ok) return;
    const error = res.error || m.nginxEditor_resetFailed();
    // A reset that cannot restart FPM leaves the version down, so say so
    // somewhere that outlives this modal.
    notifyLocalFailure('php_ini', m.phpIniEditor_restartFailed({ scope: t.label }), error);
    return error;
  }
</script>

<ConfigConfirmModal
  title={m.phpIniEditor_resetTitle()}
  size="md"
  {target}
  acceptLabel={m.nginxEditor_resetAccept()}
  busyLabel={m.phpIniEditor_restarting()}
  {run}
>
  {#snippet children(t)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.phpIniEditor_resetBody({ scope: t.label })}
    </p>
    <p class="text-[10px] text-gray-400 dark:text-gray-600 font-mono break-all">{t.path}</p>
  {/snippet}
</ConfigConfirmModal>
