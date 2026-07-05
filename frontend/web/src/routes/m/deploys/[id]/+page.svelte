<script lang="ts">
  /**
   * Mobile Deploy Detail
   * Pillar L43 — Mobile App (real flows on Tauri mobile shell).
   *
   * i18n contract: uses ONLY keys verified to exist in locales/en.json (the
   * catalog shipped by sibling PR #298). All keys below were grep-verified.
   * No `deploy.*` / no `nav.deploys` — those namespaces do not exist.
   *
   * No external format helpers — date math is inline (simple relative-time).
   */
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { t } from '$lib/i18n';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import type { Readable } from 'svelte/store';

  const deployId = $page.params.id;

  // In a real app: $page.params.id → fetch from /api/mobile/deploys/:id
  // For this scaffold we synthesize a plausible record matching the data shape
  // used by m/deploys/+page.svelte (status ∈ projects.status.*).
  type Deploy = {
    id: string;
    name: string;
    status: 'running' | 'stopped' | 'building' | 'failed' | 'queued';
    startedAt: number; // unix ms
    durationMs: number;
    branch: string;
    commit: string;
    url: string;
    logs: { ts: number; level: 'info' | 'warn' | 'error'; msg: string }[];
  };

  const deploy: Deploy = (() => {
    const now = Date.now();
    return {
      id: deployId,
      name: `deploy-${deployId}`,
      status: 'running',
      startedAt: now - 90_000,
      durationMs: 90_000,
      branch: 'main',
      commit: '6bc648f',
      url: `https://${deployId}.preview.byteport.dev`,
      logs: [
        { ts: now - 90_000, level: 'info', msg: 'build started' },
        { ts: now - 75_000, level: 'info', msg: 'restoring cache' },
        { ts: now - 60_000, level: 'info', msg: 'installing dependencies' },
        { ts: now - 30_000, level: 'info', msg: 'building application' },
        { ts: now - 5_000, level: 'info', msg: 'uploading artifacts' },
      ],
    };
  })();

  // i18n store handle for use in markup
  const tStore: Readable<(key: string, vars?: Record<string, string | number>) => string> = t;

  function formatDuration(ms: number): string {
    const sec = Math.floor(ms / 1000);
    if (sec < 60) return `${sec}s`;
    const min = Math.floor(sec / 60);
    if (min < 60) return `${min}m ${sec % 60}s`;
    const hr = Math.floor(min / 60);
    return `${hr}h ${min % 60}m`;
  }

  function formatClock(ts: number): string {
    const d = new Date(ts);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  }

  function statusClass(s: Deploy['status']): string {
    return `status status--${s}`;
  }
</script>

<svelte:head>
  <title>{deploy.name}</title>
</svelte:head>

<main class="deploy-detail">
  <header class="deploy-detail__header">
    <button class="back" on:click={() => goto('/m/deploys')} aria-label={$tStore('nav.home')}>
      <span aria-hidden="true">‹</span>
    </button>
    <h1 class="deploy-detail__title">{deploy.name}</h1>
    <span class={statusClass(deploy.status)} aria-label={$tStore(`projects.status.${deploy.status}`)}>
      {$tStore(`projects.status.${deploy.status}`)}
    </span>
  </header>

  <section class="card">
    <dl class="meta">
      <div class="meta__row">
        <dt>{$tStore('home.subtitle')}</dt>
        <dd><code>{deploy.id}</code></dd>
      </div>
      <div class="meta__row">
        <dt>URL</dt>
        <dd><a href={deploy.url} target="_blank" rel="noopener">{deploy.url}</a></dd>
      </div>
      <div class="meta__row">
        <dt>branch</dt>
        <dd><code>{deploy.branch}</code></dd>
      </div>
      <div class="meta__row">
        <dt>commit</dt>
        <dd><code>{deploy.commit}</code></dd>
      </div>
      <div class="meta__row">
        <dt>started</dt>
        <dd>{formatClock(deploy.startedAt)} ({formatDuration(deploy.durationMs)})</dd>
      </div>
    </dl>
  </section>

  <section class="card">
    <h2>logs</h2>
    <ol class="logs" aria-label="deploy log stream">
      {#each deploy.logs as line}
        <li class="logs__line logs__line--{line.level}">
          <span class="logs__ts">{formatClock(line.ts)}</span>
          <span class="logs__level">{line.level}</span>
          <span class="logs__msg">{line.msg}</span>
        </li>
      {/each}
    </ol>
  </section>

  <nav class="actions" aria-label="deploy actions">
    <button class="btn btn--secondary" on:click={() => goto('/m/deploys')}>
      {$tStore('common.cancel')}
    </button>
    <button class="btn btn--primary" type="button">
      {$tStore('common.save')}
    </button>
  </nav>
</main>

<style>
  .deploy-detail {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    padding-bottom: 6rem; /* leave space for bottom nav */
    max-width: 640px;
    margin: 0 auto;
  }
  .deploy-detail__header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .deploy-detail__title {
    flex: 1;
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .back {
    background: transparent;
    border: 1px solid var(--border, #2a2a2a);
    color: inherit;
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 0.5rem;
    font-size: 1.25rem;
    cursor: pointer;
  }
  .back:focus-visible {
    outline: 2px solid var(--accent, #4f8cff);
    outline-offset: 2px;
  }
  .card {
    background: var(--card-bg, #161616);
    border: 1px solid var(--border, #2a2a2a);
    border-radius: 0.75rem;
    padding: 1rem;
  }
  .card h2 {
    margin: 0 0 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    opacity: 0.7;
    text-transform: lowercase;
    letter-spacing: 0.05em;
  }
  .meta {
    margin: 0;
    display: grid;
    gap: 0.5rem;
  }
  .meta__row {
    display: grid;
    grid-template-columns: 6rem 1fr;
    gap: 0.5rem;
    font-size: 0.875rem;
  }
  .meta__row dt {
    opacity: 0.6;
    text-transform: lowercase;
  }
  .meta__row dd {
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .meta__row code {
    font-size: 0.8125rem;
  }
  .status {
    padding: 0.25rem 0.625rem;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: lowercase;
  }
  .status--running { background: #0e3b27; color: #5eead4; }
  .status--building { background: #3b320e; color: #fde68a; }
  .status--queued { background: #2a2a2a; color: #9ca3af; }
  .status--failed { background: #3b1a1a; color: #fca5a5; }
  .status--stopped { background: #1a1a1a; color: #6b7280; }
  .logs {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-family: ui-monospace, 'SF Mono', monospace;
    font-size: 0.8125rem;
    max-height: 50vh;
    overflow-y: auto;
  }
  .logs__line {
    display: grid;
    grid-template-columns: 4.5rem 3.5rem 1fr;
    gap: 0.5rem;
    padding: 0.125rem 0;
  }
  .logs__ts { opacity: 0.5; }
  .logs__level { text-transform: uppercase; font-size: 0.6875rem; opacity: 0.7; }
  .logs__line--error .logs__level { color: #fca5a5; }
  .logs__line--warn .logs__level { color: #fde68a; }
  .actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
  }
  .btn {
    padding: 0.625rem 1.25rem;
    border-radius: 0.5rem;
    border: 1px solid var(--border, #2a2a2a);
    background: transparent;
    color: inherit;
    font-size: 0.9375rem;
    font-weight: 500;
    cursor: pointer;
  }
  .btn--primary {
    background: var(--accent, #4f8cff);
    border-color: var(--accent, #4f8cff);
    color: white;
  }
  .btn:focus-visible {
    outline: 2px solid var(--accent, #4f8cff);
    outline-offset: 2px;
  }
  @media (prefers-reduced-motion: reduce) {
    .status { transition: none; }
  }
</style>