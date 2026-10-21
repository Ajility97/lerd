<script lang="ts">
  import TuningEditor from '$components/TuningEditor.svelte';
  import ConfigToolbar from '$components/ConfigToolbar.svelte';
  import ConfigFileBody from '$components/ConfigFileBody.svelte';
  import { ConfigFile } from '$lib/configFile.svelte';
  import {
    getServiceConfig,
    loadServiceTuningBackups,
    loadServiceTuningBackupContent,
    type Service
  } from '$stores/services';
  import {
    openTuningSaveModal,
    openTuningRestoreModal,
    openTuningResetModal
  } from '$stores/modals';
  import { onMount } from 'svelte';

  interface Props {
    svc: Service;
  }
  let { svc }: Props = $props();

  const file = new ConfigFile();

  // Pin the loader's reactive input to the service name: every services
  // broadcast passes a fresh svc object, which would otherwise clobber edits.
  const currentName = $derived(svc.name);

  $effect(() => {
    const name = currentName;
    void file.load({
      get: async () => {
        const cfg = await getServiceConfig(name);
        return { content: cfg.content, path: cfg.target, exists: cfg.exists };
      },
      backups: () => loadServiceTuningBackups(name),
      backup: (backupName) => loadServiceTuningBackupContent(name, backupName)
    });
  });

  onMount(() => () => file.dispose());

  function restore() {
    const name = currentName;
    void file.restore((backupName, backup, done) =>
      openTuningRestoreModal({ name, current: file.original, backupName, backup }, done)
    );
  }

  function reset() {
    openTuningResetModal({ name: currentName, path: file.path }, () => file.refresh());
  }

  function save() {
    openTuningSaveModal(
      { name: currentName, content: file.text, original: file.original, exists: file.exists },
      () => file.refresh()
    );
  }
</script>

<div class="flex flex-col h-full">
  <ConfigToolbar {file} onReset={reset} onRestore={restore} onSave={save} />

  <ConfigFileBody {file}>
    <TuningEditor bind:value={file.text} />
  </ConfigFileBody>
</div>
