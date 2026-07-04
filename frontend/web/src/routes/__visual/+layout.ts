import type { LayoutLoad } from './$types';

/**
 * Guard the /__visual/* routes so they're never bundled into the production
 * bundle shipped to users. Block outside of test/dev environments.
 */
export const load: LayoutLoad = ({ url }) => {
  const isDev = import.meta.env.DEV;
  const isVisualRun = process.env.PLAYWRIGHT === '1' || process.env.NODE_ENV === 'test';

  if (!isDev && !isVisualRun) {
    return {
      blocked: true,
      path: url.pathname,
    };
  }
  return { blocked: false, path: url.pathname };
};