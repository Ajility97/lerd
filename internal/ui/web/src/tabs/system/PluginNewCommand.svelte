<script lang="ts">
  import CopyButton from '$components/CopyButton.svelte';
  import { m } from '../../paraglide/messages.js';

  // The one command that starts a plugin, ready to paste into a terminal.
  const COMMAND = 'lerd plugin new my-plugin';
</script>

<div class="flex items-center gap-2.5 whitespace-nowrap rounded-lg border border-gray-200 dark:border-lerd-border bg-gray-50 dark:bg-lerd-bg px-3 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">
  <span class="text-lerd-red" aria-hidden="true">$</span>
  <code class="min-w-0 flex-1 truncate">{COMMAND}</code>
  <CopyButton text={COMMAND} label={m.common_copy()} />
</div>
