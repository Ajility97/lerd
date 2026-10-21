<script lang="ts">
  import ReloadWatcherModal from '$components/ReloadWatcherModal.svelte';
  import { type Site, installOctaneReloadWatcher } from '$stores/sites';
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
  title={m.octaneWatcher_title()}
  body={m.octaneWatcher_body()}
  command={m.octaneWatcher_command()}
  installLabel={m.octaneWatcher_installAction()}
  installingLabel={m.octaneWatcher_installing()}
  failedText={m.octaneWatcher_installFailed()}
  install={() => installOctaneReloadWatcher(site)}
  {onclose}
  {oninstalled}
/>
