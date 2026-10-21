<script lang="ts">
  import ToggleButton from '$components/ToggleButton.svelte';
  import LogsButton from '$components/LogsButton.svelte';
  import GearButton from '$components/GearButton.svelte';
  import StripeConfigModal from './StripeConfigModal.svelte';
  import { m } from '../../paraglide/messages.js';

  // Stripe listener on/off plus a gear that edits the webhook path and secret key,
  // rendered as one segmented control mirroring HorizonControl. The gear is
  // always available (even while stopped) so the route can be set before the
  // listener is first enabled, which is what non-Laravel projects need.
  interface Props {
    running: boolean;
    asleep?: boolean;
    loading?: boolean;
    webhookPath?: string;
    secretEnvKey?: string;
    onToggle: () => void;
    onSaveConfig: (path: string, secretEnvKey: string) => Promise<{ ok: boolean; error?: string }>;
    onLogs?: () => void;
  }
  let {
    running,
    asleep = false,
    loading = false,
    webhookPath = '',
    secretEnvKey = '',
    onToggle,
    onSaveConfig,
    onLogs
  }: Props = $props();

  let modalOpen = $state(false);
</script>

<div class="inline-flex items-center">
  <ToggleButton
    label={m.sites_controls_stripe()}
    on={running}
    {asleep}
    {loading}
    disabled={loading}
    rounding="rounded-l-md border-r-0"
    title={running ? m.sites_controls_stripeToggle_on() : m.sites_controls_stripeToggle_off()}
    onclick={onToggle}
  />
  <GearButton
    title={m.sites_controls_stripeConfig()}
    rounding={onLogs ? 'border-r-0' : 'rounded-r-md'}
    onclick={() => (modalOpen = true)}
  />
  {#if onLogs}
    <LogsButton label={m.sites_controls_stripe()} onclick={onLogs} />
  {/if}
</div>

<StripeConfigModal
  open={modalOpen}
  path={webhookPath}
  {secretEnvKey}
  onclose={() => (modalOpen = false)}
  onsave={onSaveConfig}
/>
