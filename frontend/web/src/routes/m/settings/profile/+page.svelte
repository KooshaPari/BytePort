<script lang="ts">
  /**
   * Mobile Profile sub-route.
   * PILLAR-TAXONOMY-v2.md v2.2 §L43 (mobile flows) + §L80 (personalization).
   *
   * i18n contract: uses ONLY keys verified to exist in locales/en.json.
   * - `settings.profile.title` / `.email` / `.joinedAt` — verified missing in catalog
   * - `common.back` — verified missing
   * - `nav.settings` — verified present
   *
   * No format imports — inline simple relative-time.
   */
  import { goto } from '$app/navigation';
  import { t } from '$lib/i18n';
  import type { Readable } from 'svelte/store';

  const tStore: Readable<(key: string, vars?: Record<string, string | number>) => string> = t;

  const profile = {
    login: 'koosha',
    displayName: 'Koosha Pari',
    email: '[email protected]',
    joinedAt: Date.now() - 1000 * 60 * 60 * 24 * 420, // ~14 months ago
  };

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
</script>

<svelte:head>
  <title>{profile.displayName}</title>
</svelte:head>

<main class="profile">
  <header class="profile__header">
    <button class="back" on:click={() => goto('/m/settings')} aria-label="back">
      <span aria-hidden="true">‹</span>
    </button>
    <h1 class="profile__title">{$tStore('nav.settings')}</h1>
  </header>

  <article class="hero">
    <div class="avatar" aria-hidden="true">{profile.displayName.charAt(0)}</div>
    <div>
      <h2>{profile.displayName}</h2>
      <p class="login">@{profile.login}</p>
    </div>
  </article>

  <ul class="rows">
    <li>
      <span class="lbl">email</span>
      <span class="val">{profile.email}</span>
    </li>
    <li>
      <span class="lbl">joined</span>
      <span class="val">{formatRelative(profile.joinedAt)}</span>
    </li>
  </ul>
</main>

<style>
  .profile {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    padding-bottom: 6rem;
    max-width: 640px;
    margin: 0 auto;
  }
  .profile__header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .profile__title {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
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
  .hero {
    padding: 1rem;
    border-radius: 0.75rem;
    background: var(--card-bg, #161616);
    border: 1px solid var(--border, #2a2a2a);
    display: flex;
    align-items: center;
    gap: 1rem;
  }
  .avatar {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: var(--accent, #4f8cff);
    color: white;
    display: grid;
    place-items: center;
    font-weight: 700;
    font-size: 1.4rem;
  }
  .hero h2 {
    margin: 0;
    font-size: 1.1rem;
  }
  .login {
    margin: 0;
    opacity: 0.7;
    font-size: 0.875rem;
  }
  .rows {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .rows li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: var(--card-bg, #161616);
    border: 1px solid var(--border, #2a2a2a);
    border-radius: 0.625rem;
    font-size: 0.875rem;
  }
  .lbl {
    opacity: 0.6;
    text-transform: lowercase;
  }
  .val {
    font-weight: 600;
  }
</style>