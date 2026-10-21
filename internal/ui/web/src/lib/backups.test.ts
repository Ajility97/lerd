import { describe, it, expect, vi, afterEach } from 'vitest';
import { fetchBackupList, fetchBackupText } from './backups';

const respond = (body: string, status = 200) => vi.stubGlobal('fetch', vi.fn(async () => new Response(body, { status })));

describe('backups', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('lists backups, and reports a failed listing instead of throwing', async () => {
    respond('[{"name":"a.conf.bkp.1"}]');
    expect(await fetchBackupList<{ name: string }>('/api/x/backups')).toEqual({ ok: true, list: [{ name: 'a.conf.bkp.1' }] });
    respond('', 500);
    const res = await fetchBackupList('/api/x/backups');
    expect(res.ok).toBe(false);
    expect(res.list).toEqual([]);
  });

  it('reads one backup, and throws when it cannot', async () => {
    respond('worker_processes 2;');
    expect(await fetchBackupText('/api/x/backups/a')).toBe('worker_processes 2;');
    respond('', 404);
    await expect(fetchBackupText('/api/x/backups/a')).rejects.toThrow();
  });
});
