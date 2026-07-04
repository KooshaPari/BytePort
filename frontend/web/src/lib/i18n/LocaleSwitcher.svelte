/**
 * LocaleSwitcher — accessible locale picker.
 *
 * Renders a `<select>` by default (`compact` mode) or a row of flag buttons
 * (`full` mode). The select version is keyboard-friendly and screen-reader
 * compatible; the full version adds visual flag buttons for sighted users
 * who like to scan.
 *
 * Persistence is handled by the i18n module (localStorage), so this
 * component only fires `setLocale` and lets the store propagate.
 */

<script lang="ts">
  import { locale, setLocale, LOCALES, LOCALE_LABELS, LOCALE_FLAGS, type Locale } from '$lib/i18n';

  export let mode: 'compact' | 'full' = 'compact';
  export let label: string = 'Language'; // i18n key: 'settings.language'
  export let id: string = 'locale-switcher';

  function onSelect(e: Event) {
    const t = e.target as HTMLSelectElement;
    setLocale(t.value as Locale);
  }

  function pick(l: Locale) {
    setLocale(l);
  }
</script>

{#if mode === 'compact'}
  <label class="locale-select" for={id}>
    <span class="locale-select__label">{label}</span>
    <select id={id} bind:value={$locale} on:change={onSelect} aria-label={label}>
      {#each LOCALES as l}
        <option value={l}>
          {LOCALE_FLAGS[l]} {LOCALE_LABELS[l]}
        </option>
      {/each}
    </select>
  </label>
{:else}
  <fieldset class="locale-flags" aria-label={label}>
    <legend class="locale-flags__legend">{label}</legend>
    {#each LOCALES as l}
      <button
        type="button"
        class="locale-flags__btn"
        class:is-active={$locale === l}
        aria-pressed={$locale === l}
        on:click={() => pick(l)}
        title={LOCALE_LABELS[l]}
      >
        <span aria-hidden="true" class="locale-flags__flag">{LOCALE_FLAGS[l]}</span>
        <span class="locale-flags__name">{LOCALE_LABELS[l]}</span>
      </button>
    {/each}
  </fieldset>
{/if}

<style>
  .locale-select {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
  }

  .locale-select__label {
    font-weight: 500;
    color: var(--bp-text-muted, #6b7280);
  }

  .locale-select select {
    padding: 0.375rem 0.5rem;
    border-radius: 0.375rem;
    border: 1px solid var(--bp-border, #e5e7eb);
    background: var(--bp-bg, white);
    color: var(--bp-text, #111827);
    font: inherit;
    cursor: pointer;
    min-width: 11ch;
  }

  .locale-select select:focus-visible {
    outline: 2px solid var(--bp-accent, #3b82f6);
    outline-offset: 2px;
  }

  .locale-flags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    border: 0;
    padding: 0;
    margin: 0;
  }

  .locale-flags__legend {
    font-weight: 500;
    color: var(--bp-text-muted, #6b7280);
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
    width: 100%;
  }

  .locale-flags__btn {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.5rem 0.75rem;
    border-radius: 9999px;
    border: 1px solid var(--bp-border, #e5e7eb);
    background: var(--bp-bg, white);
    color: var(--bp-text, #111827);
    font: inherit;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 120ms ease;
  }

  .locale-flags__btn:hover {
    border-color: var(--bp-accent, #3b82f6);
    transform: translateY(-1px);
  }

  .locale-flags__btn:focus-visible {
    outline: 2px solid var(--bp-accent, #3b82f6);
    outline-offset: 2px;
  }

  .locale-flags__btn.is-active {
    background: var(--bp-accent, #3b82f6);
    color: white;
    border-color: var(--bp-accent, #3b82f6);
  }

  .locale-flags__flag {
    font-size: 1rem;
    line-height: 1;
  }

  @media (prefers-reduced-motion: reduce) {
    .locale-flags__btn {
      transition: none;
    }
    .locale-flags__btn:hover {
      transform: none;
    }
  }
</style>
