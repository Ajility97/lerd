export interface ConfigFileSource {
  get(): Promise<{ content: string; path: string; exists: boolean }>;
  backups(): Promise<{ ok: boolean; list: { name: string }[]; error?: string }>;
  backup(name: string): Promise<string>;
}

function errorText(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

// The editing state every config file editor shares: the saved bytes and the
// buffer, its backups, and the errors kept apart so a failed action never
// blanks the editor the user is typing in.
export class ConfigFile {
  original = $state('');
  text = $state('');
  path = $state('');
  exists = $state(false);
  loading = $state(true);
  // Fatal load error, shown in place of the editor.
  error = $state('');
  actionError = $state('');
  backupsError = $state('');
  backups = $state<{ name: string }[]>([]);
  restoring = $state(false);
  // From an action landing until the reload finishes, so a second write
  // cannot go out against state the editor has not caught up with.
  applying = $state(false);
  copied = $state(false);

  dirty = $derived(this.text !== this.original);
  latestBackup = $derived(this.backups[0]);
  hasBackup = $derived(this.backups.length > 0 && !this.loading && !this.error);
  canRevert = $derived(this.dirty && !this.loading && !this.error && !this.applying);
  canReset = $derived(this.exists && !this.loading && !this.error && !this.applying);
  canSave = $derived(this.dirty && !this.loading && !this.error && !this.applying);

  #source: ConfigFileSource | null = null;
  // Bumped on every load, so a response for a file the editor has since
  // switched away from is dropped rather than overwriting the new one.
  #generation = 0;
  #copyTimer: ReturnType<typeof setTimeout> | null = null;

  async load(source: ConfigFileSource) {
    this.#source = source;
    const gen = ++this.#generation;
    this.loading = true;
    this.error = '';
    this.actionError = '';
    this.backupsError = '';
    this.original = '';
    this.text = '';
    this.path = '';
    this.backups = [];
    // allSettled so a failure on one half keeps the other: backups that
    // loaded still offer Restore when the file itself would not read.
    const [cfg, list] = await Promise.allSettled([source.get(), source.backups()]);
    if (gen !== this.#generation) return;
    if (cfg.status === 'fulfilled') this.#take(cfg.value);
    else this.error = errorText(cfg.reason);
    this.#takeBackups(list);
    this.loading = false;
  }

  async refresh() {
    const source = this.#source;
    if (!source) return;
    const gen = this.#generation;
    this.applying = true;
    try {
      const [cfg, list] = await Promise.allSettled([source.get(), source.backups()]);
      if (gen !== this.#generation) return;
      if (cfg.status === 'fulfilled') this.#take(cfg.value);
      else this.actionError = errorText(cfg.reason);
      this.#takeBackups(list);
    } finally {
      if (gen === this.#generation) this.applying = false;
    }
  }

  // Reads the latest backup and hands it to the confirmation, which calls
  // `done` once the restore landed.
  async restore(open: (name: string, content: string, done: () => Promise<void>) => void) {
    this.actionError = '';
    const source = this.#source;
    const latest = this.latestBackup;
    if (!source || !latest) return;
    this.restoring = true;
    try {
      const content = await source.backup(latest.name);
      open(latest.name, content, () => this.refresh());
    } catch (e: unknown) {
      this.actionError = errorText(e);
    } finally {
      this.restoring = false;
    }
  }

  revert() {
    this.text = this.original;
  }

  async copy() {
    try {
      await navigator.clipboard.writeText(this.text);
      this.copied = true;
      this.#clearCopyTimer();
      this.#copyTimer = setTimeout(() => (this.copied = false), 1500);
    } catch {
      /* clipboard refused; nothing to show */
    }
  }

  dispose() {
    this.#clearCopyTimer();
  }

  #clearCopyTimer() {
    if (this.#copyTimer) clearTimeout(this.#copyTimer);
  }

  #take(cfg: { content: string; path: string; exists: boolean }) {
    this.original = cfg.content;
    this.text = cfg.content;
    this.path = cfg.path;
    this.exists = cfg.exists;
  }

  #takeBackups(list: PromiseSettledResult<Awaited<ReturnType<ConfigFileSource['backups']>>>) {
    if (list.status === 'rejected') {
      this.backupsError = errorText(list.reason);
    } else if (list.value.ok) {
      this.backups = list.value.list;
      this.backupsError = '';
    } else {
      this.backupsError = list.value.error || 'Could not load backups';
    }
  }
}
