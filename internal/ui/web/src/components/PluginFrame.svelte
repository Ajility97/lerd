<script lang="ts">
  import { apiUrl } from '$lib/api';
  import { answerPluginRequest } from '$lib/pluginBridge';

  interface Props {
    plugin: string;
    url: string;
    title: string;
    // Handed to the page as query parameters (site, domain, branch on a site tab).
    context?: Record<string, string>;
    // autoHeight sizes the frame to the page inside, which reports its height
    // through ui/lerd.js; an Overview card has no space of its own to fill.
    autoHeight?: boolean;
  }
  let { plugin, url, title, context = {}, autoHeight = false }: Props = $props();

  let frame = $state<HTMLIFrameElement | null>(null);
  let height = $state(48);

  const src = $derived.by(() => {
    const q = new URLSearchParams(context).toString();
    return apiUrl(url) + (q ? '?' + q : '');
  });

  $effect(() => {
    const onMessage = (e: MessageEvent) => {
      if (autoHeight && e.source === frame?.contentWindow && e.data?.type === 'lerd:resize') {
        // Capped, so a page that grows without end cannot push the Overview away.
        height = Math.min(Math.max(Number(e.data.height) || 0, 24), 800);
        return;
      }
      void answerPluginRequest(e, plugin, frame?.contentWindow);
    };
    window.addEventListener('message', onMessage);
    return () => window.removeEventListener('message', onMessage);
  });
</script>

<!-- No allow-same-origin: the page must never act with the dashboard's standing.
     It reaches its own endpoints through answerPluginRequest instead. -->
{#key src}
  <iframe
    bind:this={frame}
    {src}
    {title}
    sandbox="allow-scripts allow-forms allow-popups allow-modals"
    class="{autoHeight ? 'block' : 'flex-1'} w-full border-0 bg-transparent"
    style:height={autoHeight ? `${height}px` : undefined}
  ></iframe>
{/key}
