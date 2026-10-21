<script lang="ts">
  import ToggleButton from '$components/ToggleButton.svelte';
  import LogsButton from '$components/LogsButton.svelte';
  import GearButton from '$components/GearButton.svelte';
  import WorkerOptionsModal from './WorkerOptionsModal.svelte';
  import { m } from '../../paraglide/messages.js';
  import type { WorkerOption } from '$stores/sites';

  // A worker toggle plus the segments its definition earns: a gear with one
  // field per tune_command placeholder, and a shortcut to its journal. A worker
  // that earns neither renders as the plain toggle, so every worker row can go
  // through this one component.
  interface Props {
    label: string;
    running: boolean;
    asleep?: boolean;
    failing?: boolean;
    unreachable?: boolean;
    loading?: boolean;
    disabled?: boolean;
    title?: string;
    options?: WorkerOption[];
    onToggle: () => void;
    onSaveOptions?: (values: Record<string, string>) => void;
    onLogs?: () => void;
  }
  let {
    label,
    running,
    asleep = false,
    failing = false,
    unreachable = false,
    loading = false,
    disabled = false,
    title = '',
    options = [],
    onToggle,
    onSaveOptions = () => {},
    onLogs
  }: Props = $props();

  let modalOpen = $state(false);

  const hasGear = $derived(options.length > 0);
  const segmented = $derived(hasGear || Boolean(onLogs));
</script>

<div class="inline-flex shrink-0 items-center">
  <ToggleButton
    {label}
    on={running}
    {asleep}
    {failing}
    {unreachable}
    {loading}
    {disabled}
    {title}
    rounding={segmented ? 'rounded-l-md border-r-0' : 'rounded-md'}
    onclick={onToggle}
  />
  {#if hasGear}
    <GearButton
      title={m.sites_controls_workerOptions()}
      rounding={onLogs ? 'border-r-0' : 'rounded-r-md'}
      onclick={() => (modalOpen = true)}
    />
  {/if}
  {#if onLogs}
    <LogsButton {label} onclick={onLogs} />
  {/if}
</div>

{#if hasGear}
  <WorkerOptionsModal
    open={modalOpen}
    {label}
    {options}
    onclose={() => (modalOpen = false)}
    onsave={onSaveOptions}
  />
{/if}
