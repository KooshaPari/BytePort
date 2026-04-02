/**
 * Sentry configuration for BytePort
 * 
 * Traces to: FR-BYTEPORT-SENTRY-001
 * 
 * Error tracking for Next.js frontend with SSR/SSG support
 */

import { init } from '@sentry/nextjs';

const SENTRY_DSN = process.env.NEXT_PUBLIC_SENTRY_DSN;

if (SENTRY_DSN && process.env.NODE_ENV !== 'development') {
  init({
    dsn: SENTRY_DSN,
    environment: process.env.NODE_ENV || 'production',
    release: process.env.VERCEL_GIT_COMMIT_SHA || process.env.BYTEPORT_VERSION,
    
    // Performance monitoring
    tracesSampleRate: 0.1,
    
    // Error sampling
    sampleRate: 1.0,
    
    // Attach stack traces
    attachStacktrace: true,
    
    // Before send to filter out noisy errors
    beforeSend(event, hint) {
      const error = hint.originalException;
      
      // Filter out common non-actionable errors
      if (error instanceof Error) {
        const ignoredPatterns = [
          /ResizeObserver loop limit exceeded/,
          /Non-Error promise rejection/,
          /Network Error/,
          /Failed to fetch/,
        ];
        
        if (ignoredPatterns.some(pattern => pattern.test(error.message))) {
          return null;
        }
      }
      
      return event;
    },
    
    // Ignore certain errors
    ignoreErrors: [
      'ResizeObserver loop limit exceeded',
      'Non-Error promise rejection captured',
      'Network request failed',
      'Failed to fetch',
    ],
  });
}

export { SENTRY_DSN };
