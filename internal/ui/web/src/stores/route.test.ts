import { describe, it, expect } from 'vitest';
import { parseHash, goToTab, TABS } from './route';

describe('parseHash', () => {
  it('defaults to dashboard for empty hash', () => {
    expect(parseHash('')).toEqual({ tab: 'dashboard', rest: '' });
    expect(parseHash('#')).toEqual({ tab: 'dashboard', rest: '' });
  });

  it('parses bare tab', () => {
    for (const t of TABS) {
      expect(parseHash('#' + t)).toEqual({ tab: t, rest: '' });
      expect(parseHash(t)).toEqual({ tab: t, rest: '' });
    }
  });

  it('parses nested route', () => {
    expect(parseHash('#services/mysql')).toEqual({ tab: 'services', rest: 'mysql' });
    expect(parseHash('#sites/foo.test/logs')).toEqual({ tab: 'sites', rest: 'foo.test/logs' });
  });

  // Plugins is a page of its own outside the rail's tabs; links from before it
  // left System, such as plugins.lerd.sh's install button, still land on it.
  it('parses the Plugins page and its old System address', () => {
    expect(parseHash('#plugins')).toEqual({ tab: 'plugins', rest: '' });
    expect(parseHash('#plugins/browse/stats')).toEqual({ tab: 'plugins', rest: 'browse/stats' });
    expect(parseHash('#system/plugins')).toEqual({ tab: 'plugins', rest: '' });
    expect(parseHash('#system/plugins/browse/stats')).toEqual({ tab: 'plugins', rest: 'browse/stats' });
    expect(TABS).not.toContain('plugins');
  });

  it('falls back to dashboard for unknown tab', () => {
    expect(parseHash('#nope')).toEqual({ tab: 'dashboard', rest: '' });
  });
});

describe('goToTab', () => {
  it('writes bare tab to hash', () => {
    goToTab('system');
    expect(location.hash).toBe('#system');
  });

  it('writes tab with rest', () => {
    goToTab('sites', 'foo.test');
    expect(location.hash).toBe('#sites/foo.test');
  });
});
