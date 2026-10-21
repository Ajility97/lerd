// Runs in the own origin of each site this plugin is switched on for, so a plain
// fetch of /_lerd/plugins/<name>/api/ reaches its endpoints. It shares the page
// with the app, so keep it small and draw inside a shadow root.
(() => {
  const name = document.currentScript.src.split('/').slice(-2, -1)[0];
  const host = document.createElement('div');
  const root = host.attachShadow({ mode: 'closed' });
  root.innerHTML = `
    <style>
      button { position: fixed; right: 12px; bottom: 12px; z-index: 2147483647;
        font: 12px system-ui, sans-serif; padding: 6px 10px; border-radius: 999px;
        border: 0; background: #1f2937; color: #fff; cursor: pointer; opacity: .85; }
    </style>
    <button type="button">${name}</button>`;
  const button = root.querySelector('button');
  button.addEventListener('click', async () => {
    const res = await fetch(`/_lerd/plugins/${name}/api/greet`);
    button.textContent = res.ok ? (await res.text()).trim() : `${name}: ${res.status}`;
  });
  document.addEventListener('DOMContentLoaded', () => document.body.appendChild(host));
})();
