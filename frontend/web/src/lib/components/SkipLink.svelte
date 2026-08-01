<!--
 * WCAG-AA accessible `SkipLink` component.
 * Renders a focus-visible link at the top of the page that skips
 * directly to `&lt;main id="main"&gt;` — keyboard users (Tab once) jump past
 * nav/header without 20+ Tabs.
 *
 * PILLAR-TAXONOMY-v2.md v2.2 §L76 (accessibility).
 *
 * Mount this component inside `+layout.svelte` immediately after the header.
 *
 * Design: link is offscreen until focused, then animates in with
 * prefers-reduced-motion respected (no transform when reduced).
 -->
<script lang="ts">
  import { t } from '$lib/i18n';

  export let href = '#main';
  export let label: string | undefined;
  const tStore = t;
</script>

<a class="skip-link" {href}>{label ?? $tStore('common.skipToMain')}</a>

<style>
  .skip-link {
    position: absolute;
    top: 0.5rem;
    left: 0.5rem;
    z-index: 1000;
    padding: 0.5rem 1rem;
    background: var(--accent, #4f8cff);
    color: white;
    text-decoration: none;
    border-radius: 0.375rem;
    font-weight: 600;
    font-size: 0.875rem;
    transform: translateY(-200%);
    transition: transform 0.15s ease-out;
  }
  .skip-link:focus {
    transform: translateY(0);
    outline: 2px solid white;
    outline-offset: 2px;
  }
  @media (prefers-reduced-motion: reduce) {
    .skip-link {
      transition: none;
    }
  }
</style>
