import type { LayoutLoad } from './$types';

/**
 * Guard the /__visual/* routes so they're never bundled into the production
 * bundle shipped to users. Block outside of test/dev environments.
 */
export const load: LayoutLoad = ({ url }) => {
  const isDev = import.meta.env.DEV;
  // `process.env` is not exposed to the browser bundle in a production
  // SvelteKit build.  Use a Vite-prefixed flag so the preview server used by
  // Playwright can render the fixtures while the normal production build
  // remains fail-closed.
  const isVisualRun =
    import.meta.env.VITE_PLAYWRIGHT === '1' ||
    process.env.PLAYWRIGHT === '1' ||
    process.env.NODE_ENV === 'test';

  if (!isDev && !isVisualRun) {
    return {
      blocked: true,
      path: url.pathname,
    };
  }
  return { blocked: false, path: url.pathname };
};
