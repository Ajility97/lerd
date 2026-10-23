// Follows the dashboard's theme inside a plugin page: the dashboard posts
// whether it is dark and its palette, on load and whenever either changes.
(() => {
  const root = document.documentElement;
  // The dashboard's language, so screen readers and the browser read the page
  // in it; the page's own lang attribute wins when it sets one.
  const lang = new URLSearchParams(location.search).get('lang');
  if (lang && !root.getAttribute('lang')) root.lang = lang;
  root.classList.toggle('dark', matchMedia('(prefers-color-scheme: dark)').matches);
  window.addEventListener('message', (e) => {
    const msg = e.data;
    if (e.source !== window.parent || !msg || msg.type !== 'lerd:theme') return;
    root.classList.toggle('dark', !!msg.dark);
    for (const [name, value] of Object.entries(msg.vars || {})) {
      if (/^--lerd-[a-z-]+$/.test(name) && typeof value === 'string') root.style.setProperty(name, value);
    }
  });
})();
