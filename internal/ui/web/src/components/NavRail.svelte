<script lang="ts">
  import { onMount } from 'svelte';
  import { tab, goToTab, TABS, type RailTab } from '$stores/route';
  import IconButton from './IconButton.svelte';
  import Icon, { type IconName } from './Icon.svelte';
  import RailLogo from './RailLogo.svelte';
  import ThemeSwitcher from './ThemeSwitcher.svelte';
  import NotificationCenter from './NotificationCenter.svelte';
  import ServiceIcon from './ServiceIcon.svelte';
  import VersionLabel from './VersionLabel.svelte';
  import {
    dashboardServices,
    dashboardOpen,
    openDashboard,
    openDocs,
    openProfiler,
    openPlugin
  } from '$stores/dashboard';
  import { railPlugins, loadPlugins, pluginUpdates } from '$stores/plugins';
  import PluginIcon from './PluginIcon.svelte';
  import { dashboardIconSvg } from '$lib/dashboardIcons';
  import { profilerEnabled, loadProfilerStatus } from '$stores/profiler';
  import { serviceLabel } from '$stores/services';
  import { accessMode } from '$stores/accessMode';
  import { desktopAppInstalled, insideDesktopApp, openInDesktopApp } from '$lib/notify';
  import { m } from '../paraglide/messages.js';

  // "Open in app" shows only in a browser when the desktop app is installed.
  const inDesktopApp = insideDesktopApp();
  const showOpenInApp = $derived($desktopAppInstalled && !inDesktopApp);

  onMount(() => {
    void loadProfilerStatus();
    void loadPlugins();
  });

  // Hide host-local launchers only when dashboard-control authority is
  // unavailable. Authenticated remote dashboards receive authority.
  const remote = $derived(!$accessMode.localControl);

  const labels = $derived<Record<RailTab, string>>({
    dashboard: m.nav_dashboard(),
    sites: m.nav_sites(),
    services: m.nav_services(),
    system: m.nav_system()
  });

  const icons: Record<RailTab, IconName> = {
    dashboard: 'dashboard',
    sites: 'sites',
    services: 'services',
    system: 'system'
  };
</script>

<aside
  class="hidden md:flex flex-col items-center w-14 shrink-0 min-h-0 bg-lerd-header-light dark:bg-lerd-header py-3 z-20"
>
  <RailLogo />

  <div class="flex flex-col gap-1 shrink-0">
    {#each TABS.filter((t) => t !== 'dashboard') as t (t)}
      <IconButton
        title={labels[t]}
        active={!$dashboardOpen && $tab === t}
        onclick={() => goToTab(t)}
      >
        <Icon name={icons[t]} />
      </IconButton>
    {/each}
  </div>

  {#if !remote}
    <div
      class="rail-rule-top flex flex-col items-center gap-1 mt-3 pt-3 w-10 min-h-0 flex-1 overflow-y-auto overscroll-contain no-scrollbar"
    >
      <IconButton
        title={m.nav_profiler()}
        active={$dashboardOpen?.name === 'profiler'}
        onclick={openProfiler}
      >
        <span class="relative flex items-center justify-center">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            {@html dashboardIconSvg('profiler')}
          </svg>
          {#if $profilerEnabled}
            <span
              title={m.profiler_toggle_on()}
              class="absolute -top-1 -right-1.5 w-2 h-2 rounded-full bg-emerald-500 ring-2 ring-lerd-header-light dark:ring-lerd-header"
            ></span>
          {/if}
        </span>
      </IconButton>
      {#each $dashboardServices as svc (svc.name)}
        <IconButton
          title={serviceLabel(svc.name) + ' ' + m.services_dashboard().toLowerCase()}
          active={$dashboardOpen?.name === svc.name}
          onclick={() => openDashboard(svc)}
        >
          <ServiceIcon
            name={svc.name}
            category={svc.category}
            icon={svc.icon}
            color={svc.color}
            preset={svc.preset}
            bare
            compact
            tint={false}
          />
        </IconButton>
      {/each}
      <!-- Plugin pages are the user's own additions, set apart from lerd's under a rule of
           their own and closed by the way to the Plugins page, so plugins can be found here. -->
      <span aria-hidden="true" data-rail-divider class="rail-rule shrink-0 w-10 my-2"></span>
      {#each $railPlugins as p (p.name)}
        <IconButton title={p.page?.title ?? p.name} active={$dashboardOpen?.plugin === p.name} onclick={() => openPlugin(p)}>
          <PluginIcon icon={p.page?.icon} />
        </IconButton>
      {/each}
      <IconButton title={m.system_plugins_title()} active={!$dashboardOpen && $tab === 'plugins'} onclick={() => goToTab('plugins')}>
        <span class="relative flex items-center justify-center">
          <Icon name="puzzle" />
          {#if $pluginUpdates > 0}
            <span
              aria-label={m.system_plugins_updatesBadge({ count: $pluginUpdates })}
              class="absolute -top-1 -right-1.5 min-w-3.5 h-3.5 px-0.5 rounded-full bg-lerd-red text-lerd-onred text-[9px] font-semibold flex items-center justify-center"
            >{$pluginUpdates > 9 ? '9+' : $pluginUpdates}</span>
          {/if}
        </span>
      </IconButton>
    </div>
  {/if}

  <div class="mt-auto shrink-0 flex flex-col items-center gap-2">
    <!-- Ordered by how often they need you: the badge first, the settings last. -->
    <NotificationCenter size="sm" />
    <IconButton
      title={m.nav_documentation()}
      active={$dashboardOpen?.name === 'docs'}
      onclick={openDocs}
      size="sm"
    >
      <Icon name="docs" />
    </IconButton>
    {#if showOpenInApp}
      <IconButton title={m.nav_open_in_app()} onclick={openInDesktopApp} size="sm">
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          stroke-width="1.75"
          stroke-linecap="round"
          stroke-linejoin="round"
          viewBox="0 0 24 24"
        >
          <path d="M14 3h7v7m0-7L10 14M19 14v5a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h5" />
        </svg>
      </IconButton>
    {/if}
    <ThemeSwitcher size="sm" />
    <VersionLabel />
  </div>
</aside>
