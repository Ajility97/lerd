import { render, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import ConfirmNginxSaveModal from './ConfirmNginxSaveModal.svelte';
import ConfirmEnvRestoreModal from './ConfirmEnvRestoreModal.svelte';
import ConfirmTuningRestoreModal from './ConfirmTuningRestoreModal.svelte';
import ConfirmTuningResetModal from './ConfirmTuningResetModal.svelte';
import ConfirmPhpIniResetModal from './ConfirmPhpIniResetModal.svelte';
import {
  modal,
  closeModal,
  openNginxSaveModal,
  openEnvRestoreModal,
  openTuningRestoreModal,
  openTuningResetModal,
  openPhpIniResetModal,
} from '$stores/modals';

const { saveSiteNginx, restoreSiteEnv } = vi.hoisted(() => ({ saveSiteNginx: vi.fn(), restoreSiteEnv: vi.fn() }));
vi.mock('$stores/sites', async (importOriginal) => {
  const actual = await importOriginal<typeof import('$stores/sites')>();
  return { ...actual, saveSiteNginx, restoreSiteEnv };
});
const { restoreServiceTuning, resetServiceTuning } = vi.hoisted(() => ({
  restoreServiceTuning: vi.fn(),
  resetServiceTuning: vi.fn(),
}));
vi.mock('$stores/services', async (importOriginal) => {
  const actual = await importOriginal<typeof import('$stores/services')>();
  return { ...actual, restoreServiceTuning, resetServiceTuning };
});
const { resetPhpIni } = vi.hoisted(() => ({ resetPhpIni: vi.fn() }));
vi.mock('$stores/phpVersions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('$stores/phpVersions')>();
  return { ...actual, resetPhpIni };
});
const { notifyLocalFailure } = vi.hoisted(() => ({ notifyLocalFailure: vi.fn() }));
vi.mock('$lib/notify', async (importOriginal) => {
  const actual = await importOriginal<typeof import('$lib/notify')>();
  return { ...actual, notifyLocalFailure };
});

beforeEach(() => {
  vi.clearAllMocks();
  closeModal();
});

describe('ConfirmNginxSaveModal', () => {
  const target = { domain: 'acme.test', scope: 'server' as const, content: 'new', original: 'old', exists: true };

  it('saves with the backup choice, closes and runs the refresh', async () => {
    saveSiteNginx.mockResolvedValue({ ok: true });
    const onSuccess = vi.fn();
    openNginxSaveModal(target, onSuccess);
    const { getByText, getByRole } = render(ConfirmNginxSaveModal);
    await fireEvent.click(getByRole('checkbox'));
    await fireEvent.click(getByText('Save'));
    expect(saveSiteNginx).toHaveBeenCalledWith('acme.test', 'server', 'new', true);
    await vi.waitFor(() => expect(get(modal).kind).toBeNull());
    expect(onSuccess).toHaveBeenCalled();
  });

  it('stays open with nginx diagnostic when validation fails', async () => {
    saveSiteNginx.mockResolvedValue({ ok: false, error: 'invalid config', validationOutput: 'unknown directive "foo"' });
    const onSuccess = vi.fn();
    openNginxSaveModal(target, onSuccess);
    const { getByText } = render(ConfirmNginxSaveModal);
    await fireEvent.click(getByText('Save'));
    await vi.waitFor(() => expect(getByText('invalid config')).toBeTruthy());
    expect(getByText('unknown directive "foo"')).toBeTruthy();
    expect(get(modal).kind).toBe('nginxSave');
    expect(onSuccess).not.toHaveBeenCalled();
  });

  it('hides the backup choice when nothing exists on disk yet', () => {
    openNginxSaveModal({ ...target, exists: false });
    const { queryByRole } = render(ConfirmNginxSaveModal);
    expect(queryByRole('checkbox')).toBeNull();
  });
});

describe('ConfirmEnvRestoreModal', () => {
  it('previews the diff and restores the exact backup shown', async () => {
    restoreSiteEnv.mockResolvedValue({ ok: true });
    openEnvRestoreModal({
      domain: 'acme.test',
      branch: '',
      file: '.env',
      current: 'A=1\nB=2',
      backupName: '.env.bkp.20260101-000000',
      backup: 'A=1\nB=3',
    });
    const { getByText } = render(ConfirmEnvRestoreModal);
    expect(getByText('- B=2')).toBeTruthy();
    expect(getByText('+ B=3')).toBeTruthy();
    await fireEvent.click(getByText('Restore'));
    expect(restoreSiteEnv).toHaveBeenCalledWith('acme.test', '', '.env', '.env.bkp.20260101-000000');
    await vi.waitFor(() => expect(get(modal).kind).toBeNull());
  });
});

describe('ConfirmTuningRestoreModal', () => {
  it('reports the rollback when the service fails to restart', async () => {
    restoreServiceTuning.mockResolvedValue({ ok: false, error: 'restart failed', rolledBack: true });
    openTuningRestoreModal({ name: 'mysql', current: 'a', backupName: 'mysql.conf.bkp.1', backup: 'b' });
    const { getByText } = render(ConfirmTuningRestoreModal);
    await fireEvent.click(getByText('Restore'));
    await vi.waitFor(() => expect(getByText('restart failed')).toBeTruthy());
    expect(getByText(/previous config was restored and is running again/)).toBeTruthy();
    expect(get(modal).kind).toBe('tuningRestore');
  });
});

describe('ConfirmTuningResetModal', () => {
  it('names the auto backup kept by a rolled back reset', async () => {
    resetServiceTuning.mockResolvedValue({ ok: false, error: 'restart failed', rolledBack: true, autoBackupName: 'mysql.conf.bkp.2' });
    openTuningResetModal({ name: 'mysql', path: '/cfg/mysql.conf' });
    const { getByText } = render(ConfirmTuningResetModal);
    await fireEvent.click(getByText('Reset'));
    await vi.waitFor(() => expect(getByText(/kept as mysql.conf.bkp.2/)).toBeTruthy());
  });
});

describe('ConfirmPhpIniResetModal', () => {
  it('raises a notification that outlives the modal when FPM does not restart', async () => {
    resetPhpIni.mockResolvedValue({ ok: false, error: 'fpm down' });
    openPhpIniResetModal({ version: '8.4', label: 'PHP 8.4', path: '/cfg/98-user.ini' });
    const { getByText } = render(ConfirmPhpIniResetModal);
    await fireEvent.click(getByText('Reset'));
    await vi.waitFor(() => expect(getByText('fpm down')).toBeTruthy());
    expect(notifyLocalFailure).toHaveBeenCalledWith('php_ini', 'PHP-FPM restart failed for PHP 8.4', 'fpm down');
  });
});
