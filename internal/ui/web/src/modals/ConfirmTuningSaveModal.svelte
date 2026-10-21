<script lang="ts">
  import ConfigConfirmModal from '$components/ConfigConfirmModal.svelte';
  import BackupCheckbox from '$components/BackupCheckbox.svelte';
  import { modal } from '$stores/modals';
  import { saveServiceConfig } from '$stores/services';
  import { m } from '../paraglide/messages.js';

  const target = $derived($modal.tuningSave);
  let backup = $state(false);
  let rolledBack = $state(false);

  async function run(t: NonNullable<typeof target>) {
    rolledBack = false;
    const res = await saveServiceConfig(t.name, t.content, backup);
    if (res.ok) return;
    // No onSuccess even when rolled back: it would reload the editor from the
    // server and throw away what the user typed while this modal stays open.
    rolledBack = res.rolledBack ?? false;
    return res.error || m.tuningEditor_saveFailed();
  }
</script>

<ConfigConfirmModal
  title={m.tuningEditor_confirmTitle()}
  size="md"
  {target}
  acceptLabel={m.common_save()}
  busyLabel={m.tuningEditor_saving()}
  {run}
>
  {#snippet children(t, busy)}
    <p class="text-sm text-gray-700 dark:text-gray-300">
      {m.tuningEditor_confirmBody({ name: t.name })}
    </p>
    {#if t.exists}
      <BackupCheckbox bind:checked={backup} disabled={busy} label={m.tuningEditor_backupLabel()} file="{t.name}.conf" />
    {/if}
  {/snippet}
  {#snippet notice()}
    {#if rolledBack}
      <p class="text-xs text-emerald-600 dark:text-emerald-400">{m.tuningEditor_rolledBack()}</p>
    {/if}
  {/snippet}
</ConfigConfirmModal>
