import { apiFetch } from '$lib/api';
import { m } from '../paraglide/messages.js';

// The two reads every config editor makes of its backups: the list, which
// reports a failure rather than throwing so the editor stays usable, and one
// backup's bytes, which throws so the restore shows why it could not start.
export async function fetchBackupList<T>(path: string): Promise<{ ok: boolean; list: T[]; error?: string }> {
  try {
    const res = await apiFetch(path);
    if (!res.ok) {
      return { ok: false, list: [], error: m.common_backupsLoadFailed({ status: res.status }) };
    }
    const list = (await res.json()) as T[];
    return { ok: true, list: Array.isArray(list) ? list : [] };
  } catch (e) {
    return { ok: false, list: [], error: e instanceof Error ? e.message : m.common_requestFailed() };
  }
}

export async function fetchBackupText(path: string): Promise<string> {
  const res = await apiFetch(path);
  if (!res.ok) throw new Error(m.common_backupLoadFailed({ status: res.status }));
  return await res.text();
}
