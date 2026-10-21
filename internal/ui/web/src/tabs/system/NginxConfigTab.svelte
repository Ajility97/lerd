<script lang="ts">
  import NginxEditor from '$components/NginxEditor.svelte';
  import ConfigToolbar from '$components/ConfigToolbar.svelte';
  import ConfigFileBody from '$components/ConfigFileBody.svelte';
  import { ConfigFile } from '$lib/configFile.svelte';
  import {
    getNginxConfig,
    loadNginxConfigBackups,
    loadNginxConfigBackupContent
  } from '$stores/nginx';
  import {
    openNginxGlobalSaveModal,
    openNginxGlobalRestoreModal,
    openNginxGlobalResetModal
  } from '$stores/modals';
  import { onMount } from 'svelte';

  const file = new ConfigFile();

  onMount(() => {
    void file.load({
      get: getNginxConfig,
      backups: loadNginxConfigBackups,
      backup: loadNginxConfigBackupContent
    });
    return () => file.dispose();
  });

  function restore() {
    void file.restore((backupName, backup, done) =>
      openNginxGlobalRestoreModal({ current: file.original, backupName, backup }, done)
    );
  }

  function reset() {
    openNginxGlobalResetModal({ path: file.path }, file.refreshLater());
  }

  function save() {
    openNginxGlobalSaveModal(
      { content: file.text, original: file.original, exists: file.exists },
      file.refreshLater()
    );
  }
</script>

<div class="flex flex-col h-full">
  <ConfigToolbar {file} onReset={reset} onRestore={restore} onSave={save} />

  <ConfigFileBody {file}>
    <NginxEditor bind:value={file.text} />
  </ConfigFileBody>
</div>
