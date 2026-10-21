import { describe, it, expect, vi, afterEach } from 'vitest';
import { pollWhileVisible } from './pollWhileVisible';

function setHidden(hidden: boolean) {
  Object.defineProperty(document, 'hidden', { configurable: true, get: () => hidden });
}

describe('pollWhileVisible', () => {
  afterEach(() => {
    vi.useRealTimers();
    setHidden(false);
  });

  it('runs at once and on every interval', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const stop = pollWhileVisible(fn, 1000);
    vi.advanceTimersByTime(2000);
    expect(fn).toHaveBeenCalledTimes(3);
    stop();
  });

  it('skips ticks while hidden and catches up when the tab shows', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const stop = pollWhileVisible(fn, 1000);
    setHidden(true);
    vi.advanceTimersByTime(3000);
    expect(fn).toHaveBeenCalledTimes(1);
    setHidden(false);
    document.dispatchEvent(new Event('visibilitychange'));
    expect(fn).toHaveBeenCalledTimes(2);
    stop();
  });

  it('stops ticking and listening once stopped', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    pollWhileVisible(fn, 1000)();
    vi.advanceTimersByTime(3000);
    document.dispatchEvent(new Event('visibilitychange'));
    expect(fn).toHaveBeenCalledTimes(1);
  });
});
