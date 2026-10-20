<script lang="ts">
  import type { Snippet } from 'svelte';
  import Modal from '$components/Modal.svelte';
  import Icon from '$components/Icon.svelte';
  import PluginIcon from '$components/PluginIcon.svelte';
  import PluginRequirements from '$components/PluginRequirements.svelte';
  import type { PluginFailure } from '$stores/plugins';
  import { m } from '../../paraglide/messages.js';

  // One plugin in full, opened from an installed card or a catalog card: its
  // screenshots, where it comes from and what it adds. The caller brings the
  // actions, since installing, updating and removing live with each list.
  interface Props {
    name: string;
    description?: string;
    version?: string;
    author?: string;
    icon?: string;
    images?: string[];
    repo?: string;
    commit?: string;
    adds?: string[];
    requires?: string[];
    missing?: string[];
    platforms?: string[];
    lerd?: string;
    tags?: string[];
    error?: string;
    failure?: PluginFailure;
    onclose: () => void;
    actions?: Snippet;
  }
  let { name, description, version, author, icon, images = [], repo, commit, adds = [], requires = [], missing = [], platforms = [], lerd, tags = [], error, failure, onclose, actions }: Props = $props();

  const PLATFORM_NAMES: Record<string, string> = { linux: 'Linux', macos: 'macOS' };
  let shown = $state(0);
  let hovered = $state(false);
  const step = (by: number) => (shown = (shown + by + images.length) % images.length);
  // Arrow keys step the slider, unless a field has the focus.
  function onKey(e: KeyboardEvent) {
    if (images.length < 2 || (e.target instanceof Element && e.target.closest('input, textarea'))) return;
    if (e.key === 'ArrowLeft') step(-1);
    if (e.key === 'ArrowRight') step(1);
  }
  // The slides move on by themselves, the way the lineup on lerd.sh does: the
  // active thumbnail's bar fills over the slide time and the slide changes when
  // it ends, so the two never drift. Hovering pauses the fill, and with reduced
  // motion there is no fill and so no change.
  const SLIDE_MS = 5000;
  const repoLabel = $derived(repo?.replace(/^https:\/\/(www\.)?github\.com\//, '') ?? '');
</script>

{#snippet label(text: string)}
  <dt class="text-[10px] font-semibold uppercase tracking-wider text-gray-400 pt-0.5">{text}</dt>
{/snippet}

<svelte:window onkeydown={onKey} />

<Modal open title={name} {onclose} size={images.length ? 'xl' : 'lg'} footer={actions}>
  <div class="p-5 space-y-5">
    <div class="flex items-start gap-3">
      <span class="shrink-0 flex items-center justify-center w-10 h-10 rounded-lg bg-gray-100 dark:bg-white/5 text-gray-600 dark:text-gray-300">
        <PluginIcon {icon} />
      </span>
      <div class="min-w-0">
        <p class="text-sm text-gray-700 dark:text-gray-200">{description ?? ''}</p>
        <p class="mt-0.5 text-xs text-gray-400">
          {#if author}{m.system_plugins_by({ author })}{/if}
          {#if author && version} · {/if}
          {#if version}<span class="font-mono tabular-nums">{version}</span>{/if}
        </p>
      </div>
    </div>

    {#if images.length}
      <figure class="space-y-2" onmouseenter={() => (hovered = true)} onmouseleave={() => (hovered = false)}>
        <!-- A fixed frame, so stepping between screenshots of other shapes never resizes the dialog. -->
        <div class="group/slider relative aspect-[16/10] w-full max-w-[88vh] mx-auto rounded-lg border border-gray-200 dark:border-lerd-border bg-gray-50 dark:bg-lerd-bg overflow-hidden">
          <img src={images[shown]} alt="" class="w-full h-full object-contain" />
          {#if images.length > 1}
            <button type="button" onclick={() => step(-1)} aria-label={m.common_pagePrev()}
              class="absolute left-2 top-1/2 -translate-y-1/2 w-9 h-9 flex items-center justify-center rounded-full bg-black/50 text-white backdrop-blur-sm opacity-70 group-hover/slider:opacity-100 hover:bg-black/70 transition">
              <Icon name="chevron" class="w-4 h-4 rotate-90" />
            </button>
            <button type="button" onclick={() => step(1)} aria-label={m.common_pageNext()}
              class="absolute right-2 top-1/2 -translate-y-1/2 w-9 h-9 flex items-center justify-center rounded-full bg-black/50 text-white backdrop-blur-sm opacity-70 group-hover/slider:opacity-100 hover:bg-black/70 transition">
              <Icon name="chevron" class="w-4 h-4 -rotate-90" />
            </button>
            <span class="absolute bottom-2 right-2 px-2 py-0.5 rounded-full bg-black/50 text-white text-[11px] tabular-nums backdrop-blur-sm">{shown + 1} / {images.length}</span>
          {/if}
        </div>
        {#if images.length > 1}
          <div class="flex gap-2 overflow-x-auto">
            {#each images as src, i (src)}
              <button type="button" onclick={() => (shown = i)} aria-label="{i + 1} / {images.length}" aria-pressed={shown === i}
                class="relative shrink-0 w-28 aspect-video rounded-md overflow-hidden border-2 {shown === i ? 'border-lerd-red' : 'border-transparent opacity-70 hover:opacity-100'}">
                <img {src} alt="" loading="lazy" class="w-full h-full object-cover object-left-top" />
                {#if shown === i}
                  {#key shown}
                    <span data-slide-bar class="slide-bar absolute left-0 bottom-0 h-0.5 w-full bg-lerd-red" class:paused={hovered} style:--slide-time="{SLIDE_MS}ms" onanimationend={() => step(1)}></span>
                  {/key}
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </figure>
    {/if}

    {#if error}
      <p class="text-xs text-red-600 dark:text-red-400 break-words">{m.system_plugins_broken()}: {error}</p>
    {/if}
    {#if failure}
      <div class="flex items-start gap-3 rounded-lg border border-red-200 dark:border-red-500/20 bg-red-50 dark:bg-red-500/5 px-3 py-2">
        <p class="flex-1 min-w-0 text-xs text-red-700 dark:text-red-300 break-words">
          {m.system_plugins_lastFailure({ what: failure.what })}: {failure.error.split('\n')[0]}
        </p>
        {#if failure.report_url}
          <a href={failure.report_url} target="_blank" rel="noopener noreferrer" class="shrink-0 text-xs font-medium text-lerd-red hover:underline">{m.system_plugins_report()}</a>
        {/if}
      </div>
    {/if}

    <dl class="grid grid-cols-[7rem_1fr] gap-x-4 gap-y-2.5 text-xs">
      {@render label(m.system_plugins_source())}
      <dd class="min-w-0 text-gray-600 dark:text-gray-300">
        {#if repo}
          <a href={repo} target="_blank" rel="noopener noreferrer" class="hover:text-lerd-red break-all">{repoLabel}</a>
          {#if commit}<span class="ml-1.5 font-mono text-gray-400">{commit.slice(0, 7)}</span>{/if}
        {:else}
          {m.system_plugins_notListed()}
        {/if}
      </dd>
      {#if adds.length}
        {@render label(m.system_plugins_adds())}
        <dd class="flex flex-wrap gap-1">
          {#each adds as a (a)}<span class="text-[10px] font-mono px-1.5 py-0.5 rounded-sm bg-gray-100 dark:bg-white/5 text-gray-500 dark:text-gray-400">{a}</span>{/each}
        </dd>
      {/if}
      {#if requires.length}
        {@render label(m.system_plugins_requires())}
        <dd><PluginRequirements {requires} {missing} bare /></dd>
      {/if}
      {#if platforms.length}
        {@render label(m.system_plugins_runsOn())}
        <dd class="text-gray-600 dark:text-gray-300">{platforms.map((p) => PLATFORM_NAMES[p] ?? p).join(', ')}</dd>
      {/if}
      {#if lerd}
        {@render label('Lerd')}
        <dd class="font-mono text-gray-600 dark:text-gray-300">{lerd}</dd>
      {/if}
      {#if tags.length}
        {@render label(m.system_plugins_tags())}
        <dd class="flex flex-wrap gap-1">
          {#each tags as t (t)}<span class="text-[10px] px-1.5 py-0.5 rounded-sm border border-gray-200 dark:border-lerd-border text-gray-500 dark:text-gray-400">{t}</span>{/each}
        </dd>
      {/if}
    </dl>
  </div>
</Modal>

<style>
  .slide-bar {
    transform-origin: left;
    animation: plugin-slide var(--slide-time) linear forwards;
  }
  .slide-bar.paused {
    animation-play-state: paused;
  }
  @keyframes plugin-slide {
    from { transform: scaleX(0); }
    to { transform: scaleX(1); }
  }
  @media (prefers-reduced-motion: reduce) {
    .slide-bar { animation: none; transform: none; }
  }
</style>
