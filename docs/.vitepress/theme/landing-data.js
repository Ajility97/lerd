/* ===========================================================
   Lerd landing page, shared data
   Ported from the standalone landing mockup (data.js).
   Install commands use the canonical lerd.sh endpoints.
   =========================================================== */

import { MARKS } from './landing-marks.js'

/* ---- Brand glyphs: the store mark where one exists, a lettered tile otherwise ---- */
export const LOGOS = {
  laravel:  { ch: 'L', c: '#ff2d20' },
  symfony:  { ch: 'S', c: '#cbd5e1' },
  wordpress:{ ch: 'W', c: '#38bdf8' },
  drupal:   { ch: 'D', c: '#60a5fa' },
  typo3:    { ch: 'T', c: '#ff8700' },
  cake:     { ch: 'C', c: '#f87171' },
  magento:  { ch: 'M', c: '#f46f25' },
  statamic: { ch: 'S', c: '#a78bfa' },
  codeigniter: { ch: 'C', c: '#f97316' },
  tempest:  { ch: 'T', c: '#29ABE2' },
  winter:   { ch: 'W', c: '#2da7c7' },
  bedrock:  { ch: 'B', c: '#21759b' },
  lumen:    { ch: 'L', c: '#f4645f' },
  yii:      { ch: 'Y', c: '#40b3d8' },
  claude:   { ch: 'C', c: '#ff8a65' },
  cursor:   { ch: '⌘', c: '#e5e7eb' },
  codex:    { ch: '{', c: '#34d399' },
  gemini:   { ch: 'G', c: '#60a5fa' },
  copilot:  { ch: '⊙', c: '#cbd5e1' },
  junie:    { ch: 'J', c: '#a78bfa' },
  windsurf: { ch: '≈', c: '#38bdf8' },
  mysql:    { ch: 'M', c: '#38bdf8' },
  postgres: { ch: 'P', c: '#60a5fa' },
  redis:    { ch: 'R', c: '#ff5b50' },
  meili:    { ch: 'M', c: '#fb7185' },
  rustfs:   { ch: 'S', c: '#fbbf24' },
  mailpit:  { ch: '@', c: '#34d399' },
  mongo:    { ch: 'M', c: '#34d399' },
  stripe:   { ch: '$', c: '#a78bfa' },
  elastic:  { ch: 'E', c: '#00bfb3' },
  antigravity: { ch: 'A', c: '#60a5fa' },
}

// `bare` drops the tile and fills the box with the mark itself, for places that
// carry their own spacing (the framework strip) rather than needing a chip.
export function glyph(name, size, opts = {}) {
  const l = LOGOS[name] || { ch: '?', c: '#888' }
  const s = size || 100
  const m = MARKS[name]
  const inset = opts.bare ? 0.04 : 0.19
  const face = m
    ? `<svg x="${s * inset}" y="${s * inset}" width="${s * (1 - inset * 2)}" height="${s * (1 - inset * 2)}"
        viewBox="${m.vb}" fill="${l.c}">${m.d}</svg>`
    : `<text x="50%" y="54%" dominant-baseline="middle" text-anchor="middle"
        font-family="JetBrains Mono, monospace" font-weight="600"
        font-size="${s * (opts.bare ? 0.86 : 0.5)}" fill="${l.c}">${l.ch}</text>`
  // No stroke: the rect fills the viewBox, so a stroke on it is half-clipped on
  // every edge. The tint alone carries the plate.
  const tile = opts.bare
    ? ''
    : `<rect width="${s}" height="${s}" rx="${s * 0.26}" fill="${l.c}" fill-opacity="0.15"/>`
  return `<svg viewBox="0 0 ${s} ${s}" xmlns="http://www.w3.org/2000/svg" style="width:100%;height:100%">
      ${tile}
      ${face}
    </svg>`
}

export const INSTALL = 'curl -fsSL https://lerd.sh/install.sh | bash'

/* ---- The lineup: one full-width card per area, screenshot path under public/assets ---- */
export const LINEUP = [
  { id: 'sites', title: 'Sites', h: 'Run lerd init, get a .test domain.',
    p: 'Lerd detects the framework, runs composer and npm, migrates the database, starts the workers and issues a trusted certificate. Nginx and PHP-FPM serve it at https://project.test.',
    alt: 'Lerd site page with runtime, workers, services and request timing for acme.test',
    tags: ['Linux', 'macOS', { t: 'WSL2 beta', beta: true }], cli: 'lerd init',
    more: 'Explore sites', href: '/usage/sites', shot: 'landing/site-overview' },
  { id: 'dashboard', title: 'Dashboard', h: 'A web UI that runs the stack.',
    p: 'Every site, service, worker and PHP version in one place. Start and stop things, tail logs, open a shell or a Tinker REPL without leaving the browser. Installable as a PWA.',
    alt: 'Lerd web dashboard listing sites, services, workers, system health and resources',
    tags: ['Web UI', 'PWA', 'Tray'], cli: 'lerd dashboard',
    more: 'Explore the dashboard', href: '/features/web-ui', shot: 'landing/dashboard' },
  { id: 'services', title: 'Services', h: 'Every service, one command away.',
    p: "MySQL, PostgreSQL, Redis, Meilisearch, Mailpit, S3 and thirty more presets. Each one wires its host, port and credentials into the site's .env for you.",
    alt: 'Lerd services page with installed databases, caches, search and mail services',
    tags: ['Rootless', 'On demand', 'Idle suspend'], cli: 'lerd service start postgres',
    more: 'Explore services', href: '/getting-started/services', shot: 'landing/services' },
  { id: 'requests', title: 'Requests', status: 'NEW', h: 'See what every request cost.',
    p: 'Every request is timed and kept with its queries, logs and outgoing calls. Open the slow one, find the N+1, profile it as an SPX flame graph, and catch every dump() in a debug window instead of the response.',
    alt: 'Lerd request timing with response time histogram, throughput and slowest routes',
    tags: ['Flame graphs', 'Queries', 'dump()'], cli: 'lerd profile',
    more: 'Explore the inspector', href: '/features/request-timing', shot: 'landing/requests-timing' },
  { id: 'worktrees', title: 'Worktrees', h: 'A domain for every branch.',
    p: 'Spin up a git worktree and Lerd gives it its own domain, certificate, dependencies and database, cloned, shared or empty depending on its migrations.',
    alt: 'Lerd add worktree dialog with branch, base, database and frontend asset options',
    tags: ['git', 'Per-branch DB'], cli: 'lerd worktree add feature/checkout',
    more: 'Explore worktrees', href: '/features/git-worktrees', shot: 'landing/worktree-add' },
  { id: 'runtimes', title: 'Runtimes', h: 'PHP and Node, pinned per project.',
    p: 'Each site picks its own PHP version, from 7.4 to 8.6, and its own Node, with Xdebug one toggle away. Node, Python and Rails apps get dev servers on the same .test domains.',
    alt: 'Lerd PHP page with installed PHP-FPM versions and live PHP logs',
    tags: ['PHP 7.4–8.6', 'Node', 'Python', 'Rails'], cli: 'lerd isolate 8.4',
    more: 'Explore runtimes', href: '/usage/php', shot: 'landing/system-php' },
  { id: 'mcp', title: 'AI / MCP', h: 'Your assistant runs the environment.',
    p: 'A built-in MCP server lets Claude, Cursor, Codex and others run artisan and Tinker, manage services, read logs and diagnose slow routes for you.',
    alt: 'Lerd Tinker REPL in the browser, the same PHP shell AI assistants drive over MCP',
    tags: ['Claude', 'Cursor', 'Codex', 'Copilot'], cli: 'lerd mcp:inject',
    more: 'Explore MCP', href: '/features/mcp', shot: 'landing/tinker' },
  { id: 'terminal', title: 'Terminal', h: 'The whole stack, in your terminal.',
    p: 'A terminal dashboard with your sites, services, workers and logs, driven by keyboard or mouse. Press ctrl+p to jump anywhere or run any action.',
    alt: 'Lerd terminal dashboard (TUI) with sites, services, resources and system status',
    tags: ['TUI', 'Mouse', 'ctrl+p'], cli: 'lerd tui',
    more: 'Explore the TUI', href: '/features/tui', shot: 'landing/tui-dashboard' },
]

/* ---- Framework grid: logo key from LOGOS, docs page under /getting-started ---- */
export const FRAMEWORKS = [
  { logo: 'laravel', name: 'Laravel', page: 'laravel' },
  { logo: 'symfony', name: 'Symfony', page: 'symfony' },
  { logo: 'wordpress', name: 'WordPress', page: 'wordpress' },
  { logo: 'drupal', name: 'Drupal', page: 'drupal' },
  { logo: 'typo3', name: 'TYPO3', page: 'typo3' },
  { logo: 'magento', name: 'Magento', page: 'magento' },
  { logo: 'cake', name: 'CakePHP', page: 'cakephp' },
  { logo: 'statamic', name: 'Statamic', page: 'statamic' },
  { logo: 'codeigniter', name: 'CodeIgniter', page: 'codeigniter' },
  { logo: 'tempest', name: 'Tempest', page: 'tempest' },
  { logo: 'winter', name: 'Winter CMS', page: 'winter' },
  { logo: 'bedrock', name: 'Bedrock', page: 'bedrock' },
  { logo: 'lumen', name: 'Lumen', page: 'lumen' },
  { logo: 'yii', name: 'Yii', page: 'yii' },
]

/* ---- Services showcase ---- */
export const SVC_SHOW = [
  { logo: 'mysql',    name: 'MySQL',       port: ':3306' },
  { logo: 'postgres', name: 'PostgreSQL',  port: ':5432' },
  { logo: 'redis',    name: 'Redis',       port: ':6379' },
  { logo: 'meili',    name: 'Meilisearch', port: ':7700' },
  { logo: 'rustfs',   name: 'RustFS / S3', port: ':9000' },
  { logo: 'mailpit',  name: 'Mailpit',     port: ':1025' },
  { logo: 'mongo',    name: 'MongoDB',     port: ':27017' },
  { logo: 'stripe',   name: 'Stripe Mock', port: ':12111' },
  { logo: 'elastic',  name: 'Elasticsearch', port: ':9200' },
]

/* ---- The rest of the store, named rather than carded, so the section stays compact.
   Split the way the store itself splits them: a preset with admin_for fronts another
   service rather than being one you would reach for on its own. ---- */
export const SVC_MORE = [
  'MariaDB', 'Valkey', 'Memcached', 'Typesense', 'ClickHouse',
  'OpenSearch', 'Solr', 'RabbitMQ', 'Kafka', 'Beanstalkd', 'Soketi', 'Mercure',
  'Selenium', 'Gotenberg', 'LocalStack', 'SpamAssassin', 'pgvector', 'TimescaleDB',
]

export const SVC_ADMIN = [
  'phpMyAdmin', 'Adminer', 'pgAdmin', 'Mongo Express', 'RedisInsight',
  'Elasticvue', 'Typesense Dashboard', 'OpenSearch Dashboards', 'Kafbat UI',
]

/* ---- Comparison ---- */
export const CMP = {
  cols: ['Lerd', 'Laravel Herd', 'Laragon', 'DDEV', 'Lando', 'Sail'],
  rows: [
    { f: 'Runs on Linux',            v: ['yes', 'no', 'no', 'yes', 'yes', 'yes'] },
    { f: 'Runs on macOS',            v: ['yes', 'yes', 'no', 'yes', 'yes', 'yes'] },
    { f: 'Runs on Windows',          v: ['partial', 'yes', 'yes', 'yes', 'yes', 'yes'] },
    { f: 'No Docker daemon',         v: ['yes', 'yes', 'yes', 'no', 'no', 'no'] },
    { f: 'Rootless · no sudo',       v: ['yes', 'yes', 'partial', 'partial', 'partial', 'no'] },
    { f: 'Open source',              v: ['yes', 'no', 'no', 'yes', 'yes', 'yes'] },
    { f: 'Free, no paid tier',       v: ['yes', 'partial', 'partial', 'yes', 'yes', 'yes'] },
    { f: 'Automatic .test + TLS',    v: ['yes', 'yes', 'yes', 'yes', 'partial', 'no'] },
    { f: 'Zero per-project config',  v: ['yes', 'yes', 'yes', 'no', 'no', 'no'] },
    { f: 'Git worktrees as sites',   v: ['yes', 'no', 'no', 'no', 'no', 'no'] },
    { f: 'Built-in Web UI',          v: ['yes', 'yes', 'no', 'partial', 'no', 'no'] },
    { f: 'Terminal dashboard',       v: ['yes', 'no', 'no', 'no', 'no', 'no'] },
    { f: 'MCP server for AI agents', v: ['yes', 'yes', 'no', 'no', 'no', 'no'] },
    { f: 'Profiler & debug window',  v: ['yes', 'yes', 'no', 'no', 'no', 'no'] },
  ],
}

/* ---- Principles ---- */
export const PRINCIPLES = [
  { h: 'No Docker daemon', p: 'Rootless Podman containers under systemd user units. Nothing runs as root, nothing idles in the background.' },
  { h: 'No sudo', p: 'Installs to ~/.local/bin. DNS, TLS and ports are set up without touching system files.' },
  { h: 'No per-project config', p: 'Frameworks are detected from the repo. No YAML to commit, no compose file to maintain.' },
  { h: 'No hardcoded stack', p: 'Frameworks and services ship as versioned definitions that update within a day, no release needed.' },
  { h: 'No licence fee', p: 'MIT licensed, open source end to end. The whole Herd workflow, free on every OS.' },
  { h: 'No system pollution', p: 'Remove Lerd and your machine is exactly as it was. Every file lives under your home.' },
]
