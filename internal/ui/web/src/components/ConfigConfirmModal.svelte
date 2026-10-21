<script lang="ts" generics="T">
  import type { Snippet } from 'svelte';
  import Modal from './Modal.svelte';
  import DetailButton from './DetailButton.svelte';
  import { closeModal, modal } from '$stores/modals';
  import { m } from '../paraglide/messages.js';

  // The shell every config editor's save, restore and reset confirmation
  // shares: busy and error state, and the refresh once the change lands.
  interface Props {
    title: string;
    size: 'md' | 'lg';
    target: T | undefined;
    acceptLabel: string;
    busyLabel: string;
    // Resolves to the error to show, or nothing when the change landed.
    run: (target: T) => Promise<string | undefined>;
    children: Snippet<[T, boolean]>;
    // Under the error, for what a failed run leaves behind.
    notice?: Snippet;
  }
  let { title, size, target, acceptLabel, busyLabel, run, children, notice }: Props = $props();

  let busy = $state(false);
  let error = $state('');

  function safeClose() {
    if (busy) return;
    closeModal();
  }

  async function accept() {
    if (!target) return;
    // Snapshot onSuccess before the await: a modal opened during the request
    // would clear $modal and the editor would keep stale dirty/backup state.
    const onSuccess = $modal.onSuccess;
    busy = true;
    error = '';
    try {
      const failure = await run(target);
      if (failure) {
        error = failure;
        return;
      }
      closeModal();
      try {
        await onSuccess?.();
      } catch {
        /* a failed refresh is surfaced in the caller's tab state */
      }
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      busy = false;
    }
  }
</script>

<Modal open {title} onclose={safeClose} {size}>
  <div class="px-5 py-4 space-y-3">
    {#if !target}
      <p class="text-sm text-gray-500 dark:text-gray-400">{m.common_loading()}</p>
    {:else}
      {@render children(target, busy)}
      {#if error}
        <p class="text-xs text-red-500">{error}</p>
      {/if}
      {@render notice?.()}
    {/if}
  </div>

  {#snippet footer()}
    <DetailButton onclick={safeClose} disabled={busy}>{m.common_cancel()}</DetailButton>
    {#if target}
      <DetailButton tone="primary" onclick={accept} loading={busy} disabled={busy}>
        {busy ? busyLabel : acceptLabel}
      </DetailButton>
    {/if}
  {/snippet}
</Modal>
