import { loadSites } from '$stores/sites';
import { m } from '../paraglide/messages.js';

// Runs one site-changing action at a time for a modal: holds the busy flag
// and the error, reloads the site list on success and flashes what happened.
export class SiteAction {
  loading = $state(false);
  error = $state('');
  flash = $state('');
  #flashTimer: ReturnType<typeof setTimeout> | null = null;

  async run(fn: () => Promise<{ ok: boolean; error?: string }>, successMsg: string) {
    this.loading = true;
    this.error = '';
    try {
      const r = await fn();
      if (!r.ok) {
        this.error = r.error || m.common_failed();
        return;
      }
      await loadSites();
      this.#showFlash(successMsg);
    } finally {
      this.loading = false;
    }
  }

  #showFlash(msg: string) {
    this.flash = msg;
    if (this.#flashTimer) clearTimeout(this.#flashTimer);
    this.#flashTimer = setTimeout(() => (this.flash = ''), 3000);
  }
}
