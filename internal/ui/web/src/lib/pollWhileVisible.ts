// pollWhileVisible runs fn now and every intervalMs, skipping ticks while the
// tab is hidden and catching up the moment it shows again. Returns the stop.
export function pollWhileVisible(fn: () => void, intervalMs: number): () => void {
  let stopped = false;
  function tick() {
    if (stopped) return;
    if (typeof document !== 'undefined' && document.hidden) return;
    fn();
  }
  tick();
  const timer = setInterval(tick, intervalMs);

  const onVisibility = () => {
    if (typeof document === 'undefined') return;
    if (!document.hidden) tick();
  };
  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', onVisibility);
  }

  return () => {
    stopped = true;
    clearInterval(timer);
    if (typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', onVisibility);
    }
  };
}
