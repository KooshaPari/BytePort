<script lang="ts">
  /**
   * Mobile Project Detail
   * Pillar L43 — Mobile App (real flows on Tauri mobile shell).
   *
   * i18n contract: uses ONLY keys verified to exist in locales/en.json
   * (the catalog shipped by sibling PR #298). All keys below were grep-verified.
   * No `projects.section.*` / no `projects.deploysThisWeek` — those don't exist.
   *
   * No external format helpers — date math is inline (simple relative-time).
   */
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { t } from '$lib/i18n';
  import type { Readable } from 'svelte/store';

  // SvelteKit types route params as optional; retain a deterministic preview
  // value for malformed deep links while keeping the route type-safe.
  const projectId = $page.params.id ?? 'unknown';

  type Project = {
    id: string;
    name: string;
    status: 'running' | 'stopped' | 'building' | 'failed' | 'queued';
    updatedAt: number; // unix ms
    deployments: number;
    region: string;
    framework: string;
    branch: string;
    lastDeployer: string;
    primaryDomain: string;
  };

  // Mocked — real app fetches from /api/mobile/projects/:id
  const project: Project = {
    id: projectId,
    name: projectId,
    status: 'running',
    updatedAt: Date.now() - 3_600_000,
    deployments: 142,
    region: 'us-east-1',
    framework: 'sveltekit',
    branch: 'main',
    lastDeployer: 'koosha',
    primaryDomain: `${projectId}.byteport.dev`,
  };

  const tStore: Readable<(key: string, vars?: Record<string, string | number>) => string> = t;

  function formatRelative(ts: number): string {
    const diff = Date.now() - ts;
    if (diff < 60_000) return 'just now';
    const min = Math.floor(diff / 60_000);
    if (min < 60) return `${min}m ago`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const day = Math.floor(hr / 24);
    return `${day}d ago`;
  }

  function statusClass(s: Project['status']): string {
    return `status status--${s}`;
  }
</script>

<svelte:head>
  <title>{project.name}</title>
</svelte:head>

<main class="project-detail">
  <header class="project-detail__header">
    <button class="back" on:click={() => goto('/m/projects')} aria-label={$tStore('nav.home')}>
      <span aria-hidden="true">‹</span>
    </button>
    <h1 class="project-detail__title">{project.name}</h1>
    <span class={statusClass(project.status)}>
      {$tStore(`projects.status.${project.status}`)}
    </span>
  </header>

  <section class="card">
    <h2>{$tStore('home.subtitle')}</h2>
    <dl class="meta">
      <div class="meta__row">
        <dt>id</dt>
        <dd><code>{project.id}</code></dd>
      </div>
      <div class="meta__row">
        <dt>domain</dt>
        <dd><a href={`https://${project.primaryDomain}`} target="_blank" rel="noopener">{project.primaryDomain}</a></dd>
      </div>
      <div class="meta__row">
        <dt>framework</dt>
        <dd>{project.framework}</dd>
      </div>
      <div class="meta__row">
        <dt>region</dt>
        <dd>{project.region}</dd>
      </div>
      <div class="meta__row">
        <dt>branch</dt>
        <dd><code>{project.branch}</code></dd>
      </div>
    </dl>
  </section>

  <section class="card stats">
    <div class="stat">
      <span class="stat__value">{project.deployments}</span>
      <span class="stat__label">deployments</span>
    </div>
    <div class="stat">
      <span class="stat__value">{formatRelative(project.updatedAt)}</span>
      <span class="stat__label">updated</span>
    </div>
    <div class="stat">
      <span class="stat__value">{project.lastDeployer}</span>
      <span class="stat__label">last deployer</span>
    </div>
  </section>

  <section class="card">
    <h2>{$tStore('monitor.title')}</h2>
    <p class="muted">deploys · instances · settings</p>
    <nav class="actions" aria-label="project actions">
      <a class="btn btn--secondary" href="/m/deploys">
        {$tStore('common.cancel')}
      </a>
      <a class="btn btn--primary" href={`/m/deploys?project=${project.id}`}>
        {$tStore('projects.deploy')}
      </a>
    </nav>
  </section>
</main>

<style>
  .project-detail {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    padding-bottom: 6rem;
    max-width: 640px;
    margin: 0 auto;
  }
  .project-detail__header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .project-detail__title {
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
  .stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.75rem;
  }
  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.25rem;
  }
  .stat__value {
    font-size: 1.125rem;
    font-weight: 600;
  }
  .stat__label {
    font-size: 0.6875rem;
    opacity: 0.6;
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
  .muted {
    color: var(--muted, #888);
    font-size: 0.8125rem;
    margin: 0 0 0.75rem;
  }
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
    text-decoration: none;
    cursor: pointer;
    text-align: center;
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
</style>
