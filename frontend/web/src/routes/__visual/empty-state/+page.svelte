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
    --bg: #ffffff;
    --text: #111827;
    --text-muted: #4b5563;
    --accent: #1d4ed8;
    --accent-hover: #1e40af;
    --accent-foreground: #ffffff;
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg, #fff);
    color: var(--text, #111827);
  }

  @media (prefers-color-scheme: dark) {
    .fixture {
      --bg: #101418;
      --text: #e1e2e8;
      --text-muted: #bec9c7;
      --accent: #9bcbfb;
      --accent-hover: #cce5ff;
      --accent-foreground: #003353;
    }
  }
</style>
