<script lang="ts">
  import { onMount } from 'svelte';
  import DetailButton from '$components/DetailButton.svelte';
  import DetailHeader from '$components/DetailHeader.svelte';
  import IconButton from '$components/IconButton.svelte';
  import Icon from '$components/Icon.svelte';
  import Modal from '$components/Modal.svelte';
  import ConfirmModal from '$components/ConfirmModal.svelte';
  import PluginBrowse from './PluginBrowse.svelte';
  import InstalledPluginTile from './InstalledPluginTile.svelte';
  import PluginDetailModal from './PluginDetailModal.svelte';
  import PluginsEmpty from './PluginsEmpty.svelte';
  import { managedPlugins, loadManagedPlugins, installPlugin, removePlugin, updatePlugin } from '$stores/plugins';
  import { routeRest, goToTab } from '$stores/route';
  import { notifyLocalFailure, notifyLocalInfo } from '$lib/notify';
  import { m } from '../../paraglide/messages.js';

  let source = $state('');
  let adding = $state(false);
  let installing = $state(false);
  let openName = $state('');
  const opened = $derived($managedPlugins.find((p) => p.name === openName));
  let removing = $state('');
  let updating = $state('');
  let confirmName = $state('');

  // #plugins is the installed list; #plugins/browse[/<name>] the catalog, which
  // plugins.lerd.sh links straight into.
  const parts = $derived($routeRest.split('/'));
  const browsing = $derived(parts[0] === 'browse');
  const focus = $derived(browsing ? decodeURIComponent(parts[1] ?? '') : '');

  async function update(name: string) {
    updating = name;
    try {
      const err = await updatePlugin(name);
      if (err) notifyLocalFailure('plugin_update', m.system_plugins_title(), err);
    } finally {
      updating = '';
    }
  }

  onMount(() => {
    void loadManagedPlugins();
  });

  async function install(e?: Event) {
    e?.preventDefault();
    const src = source.trim();
    if (!src) return;
    installing = true;
    try {
      const err = await installPlugin(src);
      if (err) {
        notifyLocalFailure('plugin_install', m.system_plugins_installFailed(), err);
        return;
      }
      source = '';
      adding = false;
      notifyLocalInfo('plugin_install', m.system_plugins_title(), m.system_plugins_installed());
    } finally {
      installing = false;
    }
  }

  async function remove() {
    const name = confirmName;
    removing = name;
    try {
      const err = await removePlugin(name);
      if (err) notifyLocalFailure('plugin_remove', m.system_plugins_title(), err);
      else openName = '';
    } finally {
      removing = '';
      confirmName = '';
    }
  }
</script>

{#snippet views()}
  <div class="flex items-center gap-1">
    <IconButton size="sm" placement="bottom" title={m.system_plugins_tabInstalled()} active={!browsing} onclick={() => goToTab('plugins')}>
      <Icon name="puzzle" class="w-4 h-4" />
    </IconButton>
    <IconButton size="sm" placement="bottom" title={m.system_plugins_tabBrowse()} active={browsing} onclick={() => goToTab('plugins', 'browse')}>
      <Icon name="download" class="w-4 h-4" />
    </IconButton>
  </div>
{/snippet}

<div class="flex-1 flex flex-col min-h-0">
  <DetailHeader title={m.system_plugins_title()} trailing={views} />

  <div class="flex-1 min-h-0 overflow-y-auto p-4 flex flex-col gap-4">
    <p class="text-xs text-gray-500 dark:text-gray-400">{m.system_plugins_description()}</p>

    {#if browsing}
      {#key focus}<PluginBrowse {focus} />{/key}
    {:else}
      <section class="flex-1 flex flex-col gap-2.5">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">{m.system_plugins_tabInstalled()}</h2>
          <button type="button" onclick={() => (adding = true)} class="inline-flex items-center gap-1 text-xs font-medium text-lerd-red hover:text-lerd-redhov">
            <Icon name="plus" class="w-3.5 h-3.5" />
            {m.system_plugins_installFrom()}
          </button>
        </div>
        {#if $managedPlugins.length === 0}
          <div class="flex-1 flex items-center justify-center"><PluginsEmpty /></div>
        {:else}
          <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-2">
            {#each $managedPlugins as p (p.name)}
              <InstalledPluginTile plugin={p} onopen={() => (openName = p.name)} />
            {/each}
          </div>
        {/if}
      </section>
    {/if}
  </div>
</div>

{#if opened}
  {@const p = opened}
  <PluginDetailModal
    name={p.name}
    description={p.description}
    version={p.version}
    icon={p.icon}
    repo={p.repo}
    commit={p.commit}
    adds={p.adds}
    requires={p.requires}
    error={p.error}
    failure={p.failure}
    onclose={() => (openName = '')}
  >
    {#snippet actions()}
      {#if p.update}
        <DetailButton tone="primary" onclick={() => update(p.name)} loading={updating === p.name} disabled={updating !== ''} title={p.update.slice(0, 7)}>
          {m.system_plugins_update()}
        </DetailButton>
      {/if}
      <DetailButton tone="danger" onclick={() => (confirmName = p.name)} loading={removing === p.name} disabled={removing !== ''}>
        {m.common_remove()}
      </DetailButton>
    {/snippet}
  </PluginDetailModal>
{/if}

<Modal open={adding} title={m.system_plugins_installFrom()} onclose={() => (adding = false)}>
  <form class="p-5 space-y-3" onsubmit={install}>
    <p class="text-xs text-gray-500 dark:text-gray-400">{m.system_plugins_installFromBody()}</p>
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="text"
      bind:value={source}
      placeholder={m.system_plugins_sourcePlaceholder()}
      aria-label={m.system_plugins_sourcePlaceholder()}
      disabled={installing}
      autofocus
      class="w-full h-8 text-sm font-mono bg-white dark:bg-lerd-bg border border-gray-200 dark:border-lerd-border rounded-md px-2.5 text-gray-700 dark:text-gray-200 focus:outline-hidden focus:border-lerd-red/50"
    />
  </form>
  {#snippet footer()}
    <DetailButton onclick={() => (adding = false)}>{m.common_cancel()}</DetailButton>
    <DetailButton tone="primary" onclick={() => install()} loading={installing} disabled={installing || !source.trim()}>
      {m.common_install()}
    </DetailButton>
  {/snippet}
</Modal>

<ConfirmModal
  open={confirmName !== ''}
  title={m.system_plugins_removeTitle({ name: confirmName })}
  body={m.system_plugins_removeBody()}
  confirmLabel={m.common_remove()}
  danger
  loading={removing !== ''}
  onconfirm={remove}
  onclose={() => (confirmName = '')}
/>
