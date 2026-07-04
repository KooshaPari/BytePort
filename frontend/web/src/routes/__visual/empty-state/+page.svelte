<script lang="ts">
  /**
   * /__visual/empty-state?variant=… — fixture for snapshotting the EmptyState
   * component in isolation. Variant maps to the `illustration` prop.
   */
  import { page } from '$app/stores';
  import EmptyState from '$lib/components/EmptyState.svelte';

  type Variant = 'no-data' | 'no-results' | 'error' | 'mascot';
  $: variant = ($page.url.searchParams.get('variant') as Variant | null) ?? 'no-data';

  const variants: Record<Variant, { title: string; description: string }> = {
    'no-data': {
      title: 'No projects yet',
      description: 'Create your first self-hosted service to get started with BytePort.',
    },
    'no-results': {
      title: 'No matches',
      description: 'Try adjusting your filters or search terms.',
    },
    error: {
      title: 'Something went wrong',
      description: 'We could not load this view. Try again in a moment.',
    },
    mascot: {
      title: 'Welcome back',
      description: 'Byte is happy to see you again.',
    },
  };
</script>

<svelte:head>
  <title>Empty state — visual regression fixture</title>
</svelte:head>

<main class="fixture">
  <EmptyState
    illustration={variant}
    title={variants[variant].title}
    description={variants[variant].description}
    primaryAction={{ label: 'Get started', href: '#' }}
  />
</main>

<style>
  .fixture {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg, #fff);
    color: var(--text, #111827);
  }
</style>