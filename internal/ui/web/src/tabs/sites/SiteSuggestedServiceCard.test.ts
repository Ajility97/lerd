import { render, fireEvent } from '@testing-library/svelte';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import SiteSuggestedServiceCard from './SiteSuggestedServiceCard.svelte';
import { presets, presetsLoaded } from '$stores/presets';

const apiFetch = vi.fn();
vi.mock('$lib/api', async (orig) => ({
  ...(await orig<typeof import('$lib/api')>()),
  apiFetch: (...args: unknown[]) => apiFetch(...args)
}));

const { loadSites } = vi.hoisted(() => ({ loadSites: vi.fn() }));
vi.mock('$stores/sites', async (importOriginal) => ({
  ...(await importOriginal<typeof import('$stores/sites')>()),
  loadSites
}));

vi.mock('$stores/presets', async (importOriginal) => ({
  ...(await importOriginal<typeof import('$stores/presets')>()),
  loadPresets: vi.fn()
}));

const { confirmDownload } = vi.hoisted(() => ({ confirmDownload: vi.fn() }));
vi.mock('$stores/downloadConfirm', async (importOriginal) => ({
  ...(await importOriginal<typeof import('$stores/downloadConfirm')>()),
  confirmDownload
}));

function ok() {
  return new Response(JSON.stringify({ ok: true }), { headers: { 'Content-Type': 'application/json' } });
}

describe('SiteSuggestedServiceCard', () => {
  beforeEach(() => {
    apiFetch.mockReset();
    apiFetch.mockResolvedValue(ok());
    confirmDownload.mockReset();
    confirmDownload.mockResolvedValue(true);
    presets.set([{ name: 'solr' }] as never);
    presetsLoaded.set(true);
  });

  // Adding may install the service, so it asks before downloading its image,
  // and a decline leaves the site as it was.
  it('asks before downloading and adds nothing on a decline', async () => {
    confirmDownload.mockResolvedValue(false);
    const { getByLabelText } = render(SiteSuggestedServiceCard, { suggestion: { name: 'solr', reason: 'Search backend' }, domain: 'shop.test' });
    await fireEvent.click(getByLabelText(/add/i));
    await vi.waitFor(() =>
      expect(confirmDownload).toHaveBeenCalledWith(expect.any(String), { service: 'solr', action: 'add' })
    );
    expect(apiFetch).not.toHaveBeenCalled();
  });

  it('adds the service to the site it was suggested for', async () => {
    const { getByLabelText } = render(SiteSuggestedServiceCard, { suggestion: { name: 'solr', reason: 'Search backend', package: 'drupal/search_api_solr' }, domain: 'shop.test' });
    await fireEvent.click(getByLabelText(/add/i));
    expect(apiFetch).toHaveBeenCalledWith('/api/sites/shop.test/service:add?name=solr', { method: 'POST' });
    await vi.waitFor(() => expect(loadSites).toHaveBeenCalled());
  });

  // With presets or versions to choose from, Add opens the picker instead.
  it('opens the picker when the suggestion has variants', async () => {
    const { get } = await import('svelte/store');
    const { modal } = await import('$stores/modals');
    presets.set([{ name: 'mysql' }, { name: 'mariadb', family: 'mariadb', env_role: 'mysql' }] as never);
    const { getByLabelText } = render(SiteSuggestedServiceCard, { suggestion: { name: 'mysql' }, domain: 'shop.test' });
    await fireEvent.click(getByLabelText(/add/i));
    await vi.waitFor(() =>
      expect(get(modal)).toEqual({ kind: 'siteServiceAdd', siteServiceAdd: { domain: 'shop.test', name: 'mysql' } })
    );
    expect(confirmDownload).not.toHaveBeenCalled();
    expect(apiFetch).not.toHaveBeenCalled();
  });

  // A preset list that failed to load says nothing about choices, so the
  // default must not be added in place of the picker.
  it('adds nothing when the presets cannot be loaded', async () => {
    presetsLoaded.set(false);
    const { getByLabelText } = render(SiteSuggestedServiceCard, { suggestion: { name: 'postgres' }, domain: 'shop.test' });
    await fireEvent.click(getByLabelText(/add/i));
    await new Promise((r) => setTimeout(r, 0));
    expect(confirmDownload).not.toHaveBeenCalled();
    expect(apiFetch).not.toHaveBeenCalled();
  });

  it('says why the service is suggested', () => {
    const { container } = render(SiteSuggestedServiceCard, {
      suggestion: { name: 'solr', reason: 'Search backend', package: 'drupal/search_api_solr' },
      domain: 'shop.test'
    });
    expect(container.textContent).toContain('Search backend');
  });

  it('dismisses the suggestion for that site', async () => {
    const { getByLabelText } = render(SiteSuggestedServiceCard, { suggestion: { name: 'solr', reason: 'Search backend', package: 'drupal/search_api_solr' }, domain: 'shop.test' });
    await fireEvent.click(getByLabelText(/suggest/i));
    expect(apiFetch).toHaveBeenCalledWith('/api/sites/shop.test/service:dismiss?name=solr', { method: 'POST' });
  });

  // A suggestion must not read as a service the site already has: the card is
  // dashed and its label faded until hovered, while its actions stay clickable.
  it('fades the suggestion but not its actions', () => {
    const { container, getByTestId, getByLabelText } = render(SiteSuggestedServiceCard, {
      suggestion: { name: 'solr', reason: 'Search backend' },
      domain: 'shop.test'
    });
    expect(container.firstElementChild?.className).toContain('border-dashed');
    const identity = getByTestId('suggested-identity').className;
    expect(identity).toContain('opacity-60');
    expect(identity).toContain('group-hover:opacity-100');
    expect(identity).toContain('grayscale');
    expect(identity).toContain('group-hover:grayscale-0');
    expect(getByLabelText('Add Solr to this site').className).not.toContain('opacity-60');
  });
});

