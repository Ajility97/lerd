<script lang="ts">
  import { revokedPlugins, keepRevokedPlugin, removePlugin } from '$stores/plugins';
  import { notifyLocalFailure } from '$lib/notify';
  import { m } from '../paraglide/messages.js';

  // A plugin pulled for a security problem stays stopped until the user decides,
  // so this has no dismiss: it goes when the plugin is removed or kept.
  let busy = $state('');

  async function act(name: string, keep: boolean) {
    busy = name;
    try {
      const err = keep ? await keepRevokedPlugin(name) : await removePlugin(name);
      if (err) notifyLocalFailure('plugin_revoked', m.plugins_revoked_title({ plugin: name }), err);
    } finally {
      busy = '';
    }
  }
</script>

{#each $revokedPlugins as r (r.name)}
  <div role="alert" class="rounded-lg border-l-4 border border-red-400 dark:border-red-500/50 border-l-red-600 bg-white/90 dark:bg-lerd-card/85 backdrop-blur-md shadow-2xl px-3 py-2.5 flex items-start gap-3">
    <svg class="w-5 h-5 shrink-0 mt-0.5 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
    </svg>
    <div class="flex-1 min-w-0">
      <p class="text-xs font-semibold text-red-900 dark:text-red-200">{m.plugins_revoked_title({ plugin: r.name })}</p>
      <p class="text-[11px] text-red-800 dark:text-red-300/90 mt-0.5">{r.reason}</p>
      <p class="text-[11px] text-gray-600 dark:text-gray-400 mt-0.5">{m.plugins_revoked_body()}</p>
    </div>
    <button
      onclick={() => act(r.name, true)}
      disabled={busy !== ''}
      class="shrink-0 text-xs font-medium text-red-900/80 dark:text-red-200/80 hover:text-red-900 dark:hover:text-red-100 border border-red-500/40 hover:border-red-500/70 rounded-sm px-2.5 py-1.5 transition-colors disabled:opacity-50"
    >{m.plugins_revoked_keep()}</button>
    <button
      onclick={() => act(r.name, false)}
      disabled={busy !== ''}
      class="shrink-0 text-xs font-medium bg-red-600 hover:bg-red-700 text-white rounded-sm px-3 py-1.5 transition-colors disabled:opacity-50"
    >{m.common_remove()}</button>
  </div>
{/each}
