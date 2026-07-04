<script lang="ts">
  /**
   * /__visual — parent route. Renders a 404-style block when the env check
   * in `+layout.ts` rejects the page (i.e. production builds).
   */
  import type { LayoutData } from './$types';
  export let data: LayoutData;
</script>

{#if data.blocked}
  <main class="blocked">
    <h1>404 — Not found</h1>
    <p>The page <code>{data.path}</code> does not exist.</p>
  </main>
{:else}
  <slot />
{/if}

<style>
  .blocked {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    color: #6b7280;
    text-align: center;
    gap: 0.5rem;
  }
</style>