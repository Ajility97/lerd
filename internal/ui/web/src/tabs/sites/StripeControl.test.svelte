<script lang="ts">
  import StripeControl from './StripeControl.svelte';

  interface Props {
    running?: boolean;
    loading?: boolean;
    webhookPath?: string;
    secretEnvKey?: string;
    onToggle?: () => void;
    onSaveConfig?: (path: string, secretEnvKey: string) => void;
  }
  let {
    running = false,
    loading = false,
    webhookPath = '',
    secretEnvKey = '',
    onToggle = () => {},
    onSaveConfig = () => {}
  }: Props = $props();
</script>

<StripeControl {running} {loading} {webhookPath} {secretEnvKey} {onToggle} {onSaveConfig} />
