<script lang="ts">
  import { machineMemory, loadMachineMemory } from '$stores/machineMemory';
  import {
    lerdResizeMachine,
    lerdStarting,
    lerdStopping,
    lerdStartStep,
    lerdStartUnit,
    lerdStartError
  } from '$stores/lerdLifecycle';
  import { accessMode } from '$stores/accessMode';
  import { startStepLabel } from '$lib/startStep';
  import { m } from '../paraglide/messages.js';
  import BottomBanner from './BottomBanner.svelte';
  import ConfirmModal from './ConfirmModal.svelte';

  // Dismissed for this many kills; a further kill brings the banner back.
  let dismissedAt = $state(0);
  let confirming = $state(false);
  let resizing = $state(false);
  let failure = $state('');

  const kills = $derived($machineMemory.kills ?? []);
  const suggestedGB = $derived(($machineMemory.suggested_mib ?? 0) / 1024);
  // A start or stop already owns the progress stores and the machine.
  const busy = $derived(resizing || $lerdStarting || $lerdStopping);
  const canResize = $derived(suggestedGB > 0 && $accessMode.localControl && !busy);
  // No current size means it could not be read, not that the Mac is full.
  const full = $derived(($machineMemory.current_mib ?? 0) > 0 && suggestedGB === 0);
  const subtitle = $derived.by(() => {
    if (resizing && $lerdStarting) return $lerdStartUnit || startStepLabel($lerdStartStep);
    if (failure) return m.machine_memory_banner_failed({ error: failure });
    if (full) return m.machine_memory_banner_noRoom();
    return m.machine_memory_banner_killed({ processes: [...new Set(kills)].join(', ') });
  });

  // The count starts over on every boot, so a lower one is a new boot.
  $effect(() => {
    if (kills.length < dismissedAt) dismissedAt = 0;
  });

  // The check reads a log on the host, so a minute's poll costs nothing in the VM.
  $effect(() => {
    loadMachineMemory();
    const t = setInterval(loadMachineMemory, 60_000);
    return () => clearInterval(t);
  });

  async function resize() {
    confirming = false;
    if (busy) return;
    resizing = true;
    failure = '';
    if (!(await lerdResizeMachine(suggestedGB * 1024))) failure = $lerdStartError || '?';
    resizing = false;
    await loadMachineMemory();
  }
</script>

{#if kills.length > dismissedAt || resizing || failure}
  <BottomBanner
    title={m.machine_memory_banner_title({ count: kills.length })}
    {subtitle}
    actionLabel={canResize ? m.machine_memory_banner_restart({ gb: suggestedGB }) : undefined}
    dismissLabel={m.machine_memory_banner_dismiss()}
    onAction={() => (confirming = true)}
    onDismiss={() => {
      dismissedAt = kills.length;
      failure = '';
    }}
  >
    {#snippet icon()}
      <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
      </svg>
    {/snippet}
  </BottomBanner>
{/if}

<ConfirmModal
  open={confirming}
  title={m.machine_memory_confirm_title()}
  body={m.machine_memory_confirm_body({ gb: suggestedGB })}
  confirmLabel={m.machine_memory_banner_restart({ gb: suggestedGB })}
  onconfirm={resize}
  onclose={() => (confirming = false)}
/>
