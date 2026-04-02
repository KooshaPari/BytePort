/**
 * Analytics integration for BytePort
 * 
 * Traces to: FR-BYTEPORT-ANALYTICS-001
 * 
 * Product analytics for IAC Deployment + UX Generation platform
 */

import { initAnalytics, track, identify, EventType } from '@phenotype/analytics';

const ANALYTICS_ENDPOINT = process.env.NEXT_PUBLIC_ANALYTICS_ENDPOINT || 'https://analytics.phenotype.dev/v1/events';
const ANALYTICS_KEY = process.env.NEXT_PUBLIC_ANALYTICS_KEY || '';

export function initBytePortAnalytics() {
  if (!ANALYTICS_KEY) {
    console.warn('[Analytics] No API key configured');
    return;
  }

  initAnalytics({
    endpoint: ANALYTICS_ENDPOINT,
    apiKey: ANALYTICS_KEY,
    environment: process.env.NODE_ENV || 'development',
    version: process.env.BYTEPORT_VERSION || 'dev',
    batchSize: 50,
    flushIntervalMs: 15000,
    debug: process.env.NODE_ENV === 'development',
  });
}

// Track deployment events
export function trackDeploymentStarted(projectId: string, config: Record<string, unknown>) {
  track(EventType.WORKFLOW_STARTED, {
    workflow: 'deployment',
    project_id: projectId,
    provider: config.provider,
    region: config.region,
  });
}

export function trackDeploymentCompleted(projectId: string, duration: number, success: boolean) {
  track(success ? EventType.WORKFLOW_COMPLETED : EventType.WORKFLOW_FAILED, {
    workflow: 'deployment',
    project_id: projectId,
    duration_ms: duration,
  });
}

// Track UX generation events
export function trackUXGenerationStarted(projectId: string, template: string) {
  track(EventType.FEATURE_USED, {
    feature: 'ux_generation',
    action: 'started',
    project_id: projectId,
    template,
  });
}

export function trackUXGenerationCompleted(projectId: string, duration: number) {
  track(EventType.OPERATION_COMPLETED, {
    operation: 'ux_generation',
    project_id: projectId,
    duration_ms: duration,
  });
}

// User identification
export function trackUserLogin(userId: string, email: string) {
  identify(userId, {
    email,
    login_time: new Date().toISOString(),
  });
}

export function trackPageView(page: string, properties?: Record<string, unknown>) {
  track(EventType.PAGE_VIEW, {
    page,
    ...properties,
  });
}
