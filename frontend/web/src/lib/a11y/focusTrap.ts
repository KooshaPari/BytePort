/**
 * FocusTrap — WCAG-AA 2.4.3 (focus order) helper.
 *
 * PILLAR-TAXONOMY-v2.md v2.2 §L76 (accessibility).
 *
 * Traps keyboard focus inside an element while open. Returns focus to
 * the previously-focused element on close. Use for modals, drawers,
 * command palettes, sheets.
 *
 * Usage:
 *   <script>
 *     import { focusTrap } from '$lib/a11y/focusTrap';
 *     let open = $state(false);
 *   </script>
 *   <dialog use:focusTrap={{ active: open }}>
 *     ...
 *   </dialog>
 */
import type { Action } from 'svelte/action';

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
  'audio[controls]',
  'video[controls]',
  '[contenteditable]:not([contenteditable="false"])',
].join(',');

export interface FocusTrapOptions {
  active: boolean;
  initialFocus?: HTMLElement | 'first' | 'container';
  returnFocus?: HTMLElement;
}

export const focusTrap: Action<HTMLElement, FocusTrapOptions> = (node, options) => {
  let opts = options;

  $effect(() => {
    if (!opts?.active) return;

    const container = node;
    const previouslyFocused = document.activeElement as HTMLElement | null;
    opts.returnFocus = previouslyFocused ?? undefined;

    // Focus initial element
    requestAnimationFrame(() => {
      if (opts.initialFocus && opts.initialFocus instanceof HTMLElement) {
        opts.initialFocus.focus();
      } else if (opts.initialFocus === 'container' || !opts.initialFocus) {
        const first = container.querySelector<HTMLElement>(FOCUSABLE_SELECTOR);
        first?.focus();
      }
    });

    function getFocusable(): HTMLElement[] {
      return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
        (el) => !el.hasAttribute('aria-hidden') && el.offsetParent !== null,
      );
    }

    function onKeyDown(e: KeyboardEvent) {
      if (e.key !== 'Tab') return;

      const focusable = getFocusable();
      if (focusable.length === 0) {
        e.preventDefault();
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const current = document.activeElement as HTMLElement;

      if (e.shiftKey) {
        if (current === first || !container.contains(current)) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (current === last || !container.contains(current)) {
          e.preventDefault();
          first.focus();
        }
      }
    }

    document.addEventListener('keydown', onKeyDown);

    return () => {
      document.removeEventListener('keydown', onKeyDown);
      opts.returnFocus?.focus?.();
    };
  });

  return {
    update(next: FocusTrapOptions) {
      opts = next;
    },
  };
};