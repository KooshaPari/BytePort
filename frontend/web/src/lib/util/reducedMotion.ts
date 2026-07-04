/**
 * useReducedMotion — bridge to `window.matchMedia('(prefers-reduced-motion: reduce)')`.
 *
 * Used by Splash.svelte, EmptyState.svelte, and any future motion-bearing components
 * to short-circuit heavy animations when the user has opted out.
 *
 * Reactive — Svelte-style readable store.
 */
import { readable } from 'svelte/store';
import { browser } from '$app/environment';

export const prefersReducedMotion = readable<boolean>(false, (set) => {
  if (!browser) return;
  const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
  set(mq.matches);
  const handler = (e: MediaQueryListEvent) => set(e.matches);
  mq.addEventListener('change', handler);
  return () => mq.removeEventListener('change', handler);
});