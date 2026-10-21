<script lang="ts">
  import TuningEditor from '$components/TuningEditor.svelte';
  import ConfigToolbar from '$components/ConfigToolbar.svelte';
  import ConfigFileBody from '$components/ConfigFileBody.svelte';
  import { ConfigFile } from '$lib/configFile.svelte';
  import {
    getPhpIni,
    loadPhpIniBackups,
    loadPhpIniBackupContent
  } from '$stores/phpVersions';
  import {
    openPhpIniSaveModal,
    openPhpIniRestoreModal,
    openPhpIniResetModal
  } from '$stores/modals';
  import { onMount } from 'svelte';

  interface Props {
    // The API scope token: a bare PHP version, or "shared" for the file every
    // version loads. One pane owns exactly one file, so its whole editing
    // state is local and saving here cannot disturb the sibling pane.
    scope: string;
    // Pane heading and the human-readable scope the confirm modals name.
    title: string;
    hint: string;
    label: string;
  }
  let { scope, title, hint, label }: Props = $props();

  const file = new ConfigFile();

  // Pin the loader's reactive input to the pane's scope so an unrelated store
  // push can't clobber unsaved edits, and a version switch reloads the file.
  const currentScope = $derived(scope);

  $effect(() => {
    const v = currentScope;
    void file.load({
      get: () => getPhpIni(v),
      backups: () => loadPhpIniBackups(v),
      backup: (name) => loadPhpIniBackupContent(v, name)
    });
  });

  onMount(() => () => file.dispose());

  function restore() {
    const v = currentScope;
    void file.restore((backupName, backup, done) =>
      openPhpIniRestoreModal({ version: v, label, current: file.original, backupName, backup }, done)
    );
  }

  function reset() {
    openPhpIniResetModal({ version: currentScope, label, path: file.path }, file.refreshLater());
  }

  function save() {
    openPhpIniSaveModal(
      { version: currentScope, label, content: file.text, original: file.original, exists: file.exists },
      file.refreshLater()
    );
  }
</script>

<section class="flex flex-col flex-1 min-w-0 min-h-64">
  <div
    class="flex items-baseline gap-2 px-3 sm:px-5 py-2 border-b border-gray-100 dark:border-lerd-border shrink-0"
  >
    <span class="text-[11px] uppercase tracking-wide font-semibold text-gray-500 dark:text-gray-400"
      >{title}</span
    >
    <span class="text-[11px] text-gray-500 dark:text-gray-400 truncate">{hint}</span>
  </div>
  <ConfigToolbar {file} onReset={reset} onRestore={restore} onSave={save} />

  <ConfigFileBody {file}>
    <TuningEditor bind:value={file.text} />
  </ConfigFileBody>
</section>
