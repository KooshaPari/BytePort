<script lang="ts">
  /**
   * Branded splash screen for the BytePort desktop application.
   *
   * Pillar mapping (PILLAR-TAXONOMY-v2.md v2.2 §L51–L60):
   * - L51: branded-splash + animated + progress-aware (fade-in on mount, mascot breathing)
   * - L52: spring-physics via svelte/motion (uses `tweened` for opacity)
   * - L53/L54: mascot illustration with idle breathing loop
   *
   * Usage (Tauri):
   *   <Splash progress={0..1} onDismiss={() => ready = true} />
   *
   * Dismissable: clicking the surface calls `onDismiss` after a 600ms minimum
   * so the animation always plays at least once.
   */
  import { onMount, tweened } from 'svelte/motion';
  import { cubicOut } from 'svelte/easing';

  export let progress: number = 0;
  export let dismissable: boolean = true;
  export let onDismiss: (() => void) | null = null;
  export let minDurationMs: number = 600;

  let mounted = false;
  let elapsed = 0;
  let dismissed = false;

  const opacity = tweened(0, { duration: 480, easing: cubicOut });
  const scale = tweened(0.96, { duration: 720, easing: cubicOut });

  onMount(() => {
    mounted = true;
    opacity.set(1);
    scale.set(1);
    const start = performance.now();
    const tick = () => {
      elapsed = performance.now() - start;
      if (elapsed < minDurationMs || !dismissed) {
        requestAnimationFrame(tick);
      }
    };
    requestAnimationFrame(tick);
  });

  function handleDismiss() {
    if (!dismissable || elapsed < minDurationMs) return;
    dismissed = true;
    opacity.set(0, { duration: 280 });
    scale.set(1.04, { duration: 280 });
    setTimeout(() => onDismiss?.(), 320);
  }

  function handleKey(e: KeyboardEvent) {
    if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleDismiss();
    }
  }
</script>

<svelte:window on:keydown={handleKey} />

{#if mounted}
  <div
    class="splash"
    class:dismissable
    role="dialog"
    aria-modal="true"
    aria-labelledby="splash-title"
    aria-describedby="splash-desc"
    style="opacity: {$opacity}; transform: scale({$scale})"
    on:click={handleDismiss}
    on:keydown={handleKey}
    tabindex="-1"
  >
    <img class="art" src="/brand/splash.svg" alt="" aria-hidden="true" draggable="false" />
    <h1 id="splash-title" class="visually-hidden">BytePort</h1>
    <p id="splash-desc" class="visually-hidden">
      Self-hosted IaC + portfolio platform. Loading.
    </p>

    <div class="progress" role="progressbar" aria-valuenow={Math.round(progress * 100)} aria-valuemin="0" aria-valuemax="100" aria-label="Loading progress">
      <div class="progress-bar" style="width: {Math.max(0, Math.min(1, progress)) * 100}%" />
    </div>

    {#if dismissable && elapsed >= minDurationMs}
      <button class="dismiss" type="button" on:click|stopPropagation={handleDismiss} aria-label="Dismiss splash">
        Skip →
      </button>
    {/if}
  </div>
{/if}

<style>
  .splash {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1.5rem;
    background: radial-gradient(circle at 50% 35%, #1c2756 0%, #0a0f24 60%, #05080f 100%);
    color: #fff;
    user-select: none;
    transition: opacity 280ms cubic-bezier(0.2, 0.8, 0.2, 1);
    outline: none;
  }
  .splash.dismissable {
    cursor: pointer;
  }
  .splash:focus-visible {
    outline: 2px solid #7aa2ff;
    outline-offset: -4px;
  }
  .art {
    width: min(60vw, 480px);
    height: auto;
    pointer-events: none;
  }
  .progress {
    width: min(40vw, 280px);
    height: 4px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 999px;
    overflow: hidden;
  }
  .progress-bar {
    height: 100%;
    background: linear-gradient(90deg, #7aa2ff 0%, #a4b8ff 100%);
    border-radius: 999px;
    transition: width 320ms cubic-bezier(0.2, 0.8, 0.2, 1);
  }
  .dismiss {
    appearance: none;
    border: 1px solid rgba(255, 255, 255, 0.18);
    background: rgba(255, 255, 255, 0.04);
    color: rgba(255, 255, 255, 0.85);
    padding: 0.45rem 0.9rem;
    border-radius: 999px;
    font-size: 0.85rem;
    cursor: pointer;
    transition: background 200ms cubic-bezier(0.2, 0.8, 0.2, 1), transform 200ms;
  }
  .dismiss:hover {
    background: rgba(255, 255, 255, 0.1);
    transform: translateY(-1px);
  }
  .dismiss:focus-visible {
    outline: 2px solid #7aa2ff;
    outline-offset: 2px;
  }
  .visually-hidden {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .splash,
    .progress-bar,
    .dismiss {
      transition: none !important;
    }
  }
</style>