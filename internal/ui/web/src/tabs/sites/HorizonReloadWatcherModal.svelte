<script lang="ts">
  import ReloadWatcherModal from '$components/ReloadWatcherModal.svelte';
  import { type Site, installHorizonReloadWatcher } from '$stores/sites';
  import { m } from '../../paraglide/messages.js';

  interface Props {
    open: boolean;
    site: Site;
    onclose: () => void;
    oninstalled: () => void;
  }
  let { open, site, onclose, oninstalled }: Props = $props();
</script>

<ReloadWatcherModal
  {open}
  title={m.horizonWatcher_title()}
  body={m.horizonWatcher_body()}
  command={m.horizonWatcher_command()}
  installLabel={m.horizonWatcher_installAction()}
  installingLabel={m.horizonWatcher_installing()}
  failedText={m.horizonWatcher_installFailed()}
  install={() => installHorizonReloadWatcher(site)}
  {onclose}
  {oninstalled}
/>
