import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import SiteServiceAddModal from './SiteServiceAddModal.svelte';
import { presets, presetsLoaded } from '$stores/presets';
import { modal } from '$stores/modals';

vi.mock('$stores/presets', async (importOriginal) => ({
  ...(await importOriginal<typeof import('$stores/presets')>()),
  loadPresets: vi.fn()
}));

const { addSuggestedService, loadSites } = vi.hoisted(() => ({
  addSuggestedService: vi.fn(),
  loadSites: vi.fn()
}));
vi.mock('$stores/sites', async (importOriginal) => ({
  ...(await importOriginal<typeof import('$stores/sites')>()),
  addSuggestedService,
  loadSites
}));

const { confirmDownload } = vi.hoisted(() => ({ confirmDownload: vi.fn() }));
vi.mock('$stores/downloadConfirm', async (importOriginal) => ({
  ...(await importOriginal<typeof import('$stores/downloadConfirm')>()),
  confirmDownload
}));

const postgres = {
  name: 'postgres',
  family: 'postgres',
  versions: [{ tag: '18' }, { tag: '17' }, { tag: '16' }],
  default_version: '16'
};
const pgvector = {
  name: 'postgres-pgvector',
  family: 'postgres',
  description: 'PostgreSQL with pgvector',
  versions: [{ tag: '18' }, { tag: '17' }],
  default_version: '18'
};

function open(list: unknown[]) {
  presets.set(list as never);
  modal.set({
    kind: 'siteServiceAdd',
    siteServiceAdd: { domain: 'shop.test', name: 'postgres' }
  });
  render(SiteServiceAddModal);
}

describe('SiteServiceAddModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    presetsLoaded.set(true);
    confirmDownload.mockResolvedValue(true);
    addSuggestedService.mockResolvedValue({ ok: true });
  });

  afterEach(() => {
    presets.set([]);
    presetsLoaded.set(false);
    modal.set({ kind: null });
  });

  it('starts on the suggested service at its default version', async () => {
    open([postgres, pgvector]);
    await fireEvent.click(screen.getByText('Add'));
    await vi.waitFor(() =>
      expect(addSuggestedService).toHaveBeenCalledWith('shop.test', 'postgres', 'postgres', '16')
    );
    expect(confirmDownload).toHaveBeenCalledWith(expect.any(String), { preset: 'postgres', version: '16' });
    await vi.waitFor(() => expect(loadSites).toHaveBeenCalled());
  });

  // What already runs here adds without a download, so it is the first pick.
  it('starts on a variant this machine already runs', async () => {
    open([postgres, { ...pgvector, installed: true, installed_tags: ['17'] }]);
    await fireEvent.click(screen.getByText('Add'));
    await vi.waitFor(() =>
      expect(addSuggestedService).toHaveBeenCalledWith('shop.test', 'postgres', 'postgres-pgvector', '17')
    );
  });

  it('adds nothing when the download is declined', async () => {
    confirmDownload.mockResolvedValue(false);
    open([postgres, pgvector]);
    await fireEvent.click(screen.getByText('Add'));
    await vi.waitFor(() => expect(confirmDownload).toHaveBeenCalled());
    expect(addSuggestedService).not.toHaveBeenCalled();
  });

  it('shows why the add failed and stays open', async () => {
    addSuggestedService.mockResolvedValue({ ok: false, error: 'lerd link: boom' });
    open([postgres, pgvector]);
    await fireEvent.click(screen.getByText('Add'));
    expect(await screen.findByText('lerd link: boom')).toBeInTheDocument();
  });
});
