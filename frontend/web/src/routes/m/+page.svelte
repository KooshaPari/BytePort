<script lang="ts">
  /**
   * Mobile home — landing page behind the bottom-nav "home" icon.
   * Shows a compact welcome, recent activity, and quick-action chips.
   * PILLAR-TAXONOMY-v2.md v2.2 §L43 (responsive PWA — first meaningful paint)
   */
  import { t } from '$lib/i18n';

  let now = $state(new Date());
  $effect(() => {
    const id = setInterval(() => (now = new Date()), 30_000);
    return () => clearInterval(id);
  });

  const greeting = $derived(() => {
    const h = now.getHours();
    if (h < 5) return 'common.loading';
    if (h < 12) return 'home.welcome';
    if (h < 18) return 'home.welcome';
    return 'home.welcome';
  });
</script>

<svelte:head><title>BytePort</title></svelte:head>

<section class="home">
  <header class="hero" aria-labelledby="home-h">
    <img class="mascot" src="/brand/mascot.svg" alt="" />
    <div>
      <h1 id="home-h">BytePort</h1>
      <p class="sub">{$t('emptyStates.mascot.title')}</p>
    </div>
  </header>

  <div class="quick" role="group" aria-label="Quick actions">
    <a class="chip" href="/m/projects">
      <span class="dot dot-p" aria-hidden="true"></span>
      {$t('nav.projects')}
    </a>
    <a class="chip" href="/m/deploys">
      <span class="dot dot-d" aria-hidden="true"></span>
      {$t('nav.deploys')}
    </a>
    <a class="chip" href="/m/settings">
      <span class="dot dot-s" aria-hidden="true"></span>
      {$t('nav.settings')}
    </a>
  </div>

  <p class="greeting">{$t(greeting())}</p>
</section>

<style>
  .home { display: flex; flex-direction: column; gap: 1.25rem; }
  .hero {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: var(--card, #0e1514);
    border: 1px solid var(--border, #3f4948);
    border-radius: 14px;
    padding: 1rem;
  }
  .mascot { width: 56px; height: 56px; flex: 0 0 56px; }
  .hero h1 { font-size: 1.4rem; margin: 0; color: var(--foreground, #dde4e2); }
  .hero .sub { margin: 4px 0 0; color: var(--muted-foreground, #bec9c7); font-size: 0.85rem; }

  .quick {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.5rem;
  }
  .chip {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    min-height: 44px;
    padding: 0.5rem 0.75rem;
    background: var(--surface, #101418);
    border: 1px solid var(--border, #3f4948);
    border-radius: 10px;
    color: var(--foreground, #dde4e2);
    text-decoration: none;
    font-size: 0.9rem;
    transition: background 140ms;
  }
  .chip:hover { background: var(--muted, #101418); }
  .dot { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }
  .dot-p { background: var(--primary, #80d5cf); }
  .dot-d { background: var(--secondary, #83d2e3); }
  .dot-s { background: var(--accent, #9bcbfb); }

  .greeting { color: var(--muted-foreground, #bec9c7); font-size: 0.9rem; margin: 0; }

  @media (min-width: 640px) {
    .home { max-width: 720px; margin: 0 auto; }
    .hero { padding: 1.5rem 2rem; }
    .mascot { width: 80px; height: 80px; flex-basis: 80px; }
  }
</style>
