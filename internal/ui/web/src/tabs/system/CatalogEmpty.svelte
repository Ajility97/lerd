<script lang="ts">
  import PluginArt from './PluginArt.svelte';
  import PluginGlyph from './PluginGlyph.svelte';
  import PluginNewCommand from './PluginNewCommand.svelte';
  import { PLUGIN_SUBMIT_GUIDE } from '$stores/plugins';
  import { m } from '../../paraglide/messages.js';

  // How a plugin gets into the catalog, from the first command to the pull
  // request a maintainer reviews: the whole of Browse while the catalog lists
  // nothing, and its foot, with more set, once it lists some.
  let { more = false }: { more?: boolean } = $props();
  const EXAMPLE = 'https://github.com/lerd-env/plugins/pull/1';
  const steps = [
    [m.system_plugins_step1Title, m.system_plugins_step1Body],
    [m.system_plugins_step2Title, m.system_plugins_step2Body],
    [m.system_plugins_step3Title, m.system_plugins_step3Body]
  ];
  const LINK = 'inline-flex items-center gap-1.5 h-8 px-3 rounded-md text-xs font-semibold';
</script>

<div class="w-full mx-auto max-w-4xl pt-4 pb-6 text-center">
  <div class="flex justify-center"><PluginArt kind="catalog" /></div>
  <h2 class="mt-4 text-xl font-semibold tracking-tight text-gray-900 dark:text-white">{more ? m.system_plugins_catalogMoreTitle() : m.system_plugins_catalogEmpty()}</h2>
  <p class="mx-auto mt-1.5 max-w-lg text-sm text-gray-500 dark:text-gray-400">{more ? m.system_plugins_catalogMoreLead() : m.system_plugins_catalogEmptyLead()}</p>
  <ol class="mt-6 grid gap-3 sm:grid-cols-3 text-left">
    {#each steps as [title, body], i (i)}
      <li class="flex flex-col gap-2 rounded-xl border border-gray-200/80 dark:border-lerd-border bg-white dark:bg-lerd-card p-4">
        <div class="flex items-center gap-2.5">
          <span class="flex items-center justify-center w-6 h-6 rounded-full bg-lerd-red/10 text-lerd-red text-xs font-bold">{i + 1}</span>
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{title()}</h3>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400">{body()}</p>
        {#if i === 0}<PluginNewCommand />{/if}
      </li>
    {/each}
  </ol>
  <div class="mt-5 flex flex-wrap justify-center gap-2">
    <a href={PLUGIN_SUBMIT_GUIDE} target="_blank" rel="noopener noreferrer" class="{LINK} bg-lerd-red hover:bg-lerd-redhov text-lerd-onred">
      <PluginGlyph name="page" class="w-3.5 h-3.5" />{m.system_plugins_submitGuide()}
    </a>
    <a href={EXAMPLE} target="_blank" rel="noopener noreferrer" class="{LINK} border border-gray-200 dark:border-lerd-border bg-white dark:bg-lerd-card text-gray-700 dark:text-gray-200 hover:border-gray-300 dark:hover:border-white/15">
      <PluginGlyph name="code" class="w-3.5 h-3.5" />{m.system_plugins_submitExample()}
    </a>
  </div>
</div>
