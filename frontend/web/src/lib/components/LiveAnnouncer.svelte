/**
 * Live-region announcer for SPA navigation + dynamic state changes.
 *
 * PILLAR-TAXONOMY-v2.md v2.2 §L76 (accessibility).
 *
 * Mount this component once in +layout.svelte. Screen-reader users
 * get a polite announcement whenever you call announce() from anywhere
 * via the exported function. Use it for:
 *   - SPA route changes (so the new page name is read out)
 *   - Form submission results ("Saved", "Error: email invalid")
 *   - Async operation completions ("Project deployed")
 *
 * politeness: 'polite' (default) queues after current speech;
 *             'assertive' interrupts (use sparingly, e.g. errors)
 *
 * The visible label is sr-only — only AT users perceive it.
 */
<script lang="ts" module>
  let _instance: { announce: (msg: string, politeness?: 'polite' | 'assertive') => void } | null =
    null;

  export function announce(msg: string, politeness: 'polite' | 'assertive' = 'polite'): void {
    if (_instance) {
      _instance.announce(msg, politeness);
    }
  }
</script>

<script lang="ts">
  let politeMsg = $state('');
  let assertiveMsg = $state('');

  export function announce(msg: string, politeness: 'polite' | 'assertive' = 'polite'): void {
    if (politeness === 'polite') {
      politeMsg = '';
      // tick: setTimeout ensures screen reader picks up the change
      setTimeout(() => (politeMsg = msg), 50);
    } else {
      assertiveMsg = '';
      setTimeout(() => (assertiveMsg = msg), 50);
    }
  }

  // expose to module-level export
  $effect(() => {
    _instance = { announce };
  });
</script>

<div class="sr-only" aria-live="polite" aria-atomic="true">{politeMsg}</div>
<div class="sr-only" aria-live="assertive" aria-atomic="true">{assertiveMsg}</div>

<style>
  .sr-only {
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
</style>
