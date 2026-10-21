import { describe, it, expect, vi } from 'vitest';
import { ConfigFile, type ConfigFileSource } from './configFile.svelte';

function source(over: Partial<ConfigFileSource> = {}): ConfigFileSource {
  return {
    get: vi.fn().mockResolvedValue({ content: 'saved', path: '/cfg/a.conf', exists: true }),
    backups: vi.fn().mockResolvedValue({ ok: true, list: [{ name: 'a.conf.bkp.2' }, { name: 'a.conf.bkp.1' }] }),
    backup: vi.fn().mockResolvedValue('older'),
    ...over,
  };
}

describe('ConfigFile', () => {
  it('loads the file and its backups', async () => {
    const f = new ConfigFile();
    await f.load(source());
    expect(f.loading).toBe(false);
    expect(f.text).toBe('saved');
    expect(f.path).toBe('/cfg/a.conf');
    expect(f.latestBackup?.name).toBe('a.conf.bkp.2');
    expect(f.canReset).toBe(true);
    expect(f.canSave).toBe(false);
  });

  it('keeps the backups when the file itself fails to load', async () => {
    const f = new ConfigFile();
    await f.load(source({ get: vi.fn().mockRejectedValue(new Error('boom')) }));
    expect(f.error).toBe('boom');
    expect(f.backups).toHaveLength(2);
  });

  it('shows a backups listing failure without blanking the editor', async () => {
    const f = new ConfigFile();
    await f.load(source({ backups: vi.fn().mockResolvedValue({ ok: false, list: [], error: 'denied' }) }));
    expect(f.backupsError).toBe('denied');
    expect(f.error).toBe('');
    expect(f.text).toBe('saved');
  });

  it('drops a response that lands after the next load started', async () => {
    let resolveFirst!: (v: { content: string; path: string; exists: boolean }) => void;
    const slow = source({ get: vi.fn(() => new Promise<{ content: string; path: string; exists: boolean }>((r) => (resolveFirst = r))) });
    const f = new ConfigFile();
    const first = f.load(slow);
    await f.load(source({ get: vi.fn().mockResolvedValue({ content: 'second', path: '/b', exists: true }) }));
    resolveFirst({ content: 'first', path: '/a', exists: true });
    await first;
    expect(f.text).toBe('second');
  });

  it('marks edits dirty and reverts them to the saved bytes', async () => {
    const f = new ConfigFile();
    await f.load(source());
    f.text = 'edited';
    expect(f.dirty).toBe(true);
    expect(f.canSave).toBe(true);
    f.revert();
    expect(f.text).toBe('saved');
  });

  it('reports a failed refresh inline instead of replacing the editor', async () => {
    const src = source();
    const f = new ConfigFile();
    await f.load(src);
    vi.mocked(src.get).mockRejectedValueOnce(new Error('gone'));
    await f.refresh();
    expect(f.actionError).toBe('gone');
    expect(f.error).toBe('');
    expect(f.applying).toBe(false);
  });

  it('hands the latest backup to the restore confirmation and reloads once it lands', async () => {
    const src = source();
    const f = new ConfigFile();
    await f.load(src);
    const open = vi.fn();
    await f.restore(open);
    expect(src.backup).toHaveBeenCalledWith('a.conf.bkp.2');
    const [name, content, done] = open.mock.calls[0];
    expect([name, content]).toEqual(['a.conf.bkp.2', 'older']);
    vi.mocked(src.get).mockResolvedValueOnce({ content: 'older', path: '/cfg/a.conf', exists: true });
    await done();
    expect(f.text).toBe('older');
    expect(f.restoring).toBe(false);
  });

  it('shows why a backup could not be read', async () => {
    const f = new ConfigFile();
    await f.load(source({ backup: vi.fn().mockRejectedValue(new Error('unreadable')) }));
    const open = vi.fn();
    await f.restore(open);
    expect(open).not.toHaveBeenCalled();
    expect(f.actionError).toBe('unreadable');
  });
});
