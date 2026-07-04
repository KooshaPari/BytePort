/**
 * BytePort i18n — zero-dependency custom implementation.
 *
 * Why not svelte-i18n / paraglide? Both add 8-30 kB and a bundler transform
 * for functionality that fits in ~80 lines of TS. We pay $0 in deps and
 * keep full control over pluralization, interpolation, and locale persistence.
 *
 * Locale files live in src/lib/i18n/locales/*.json — flat keys separated by
 * dots, e.g. { "nav.projects": "Projects" }.
 *
 * Usage:
 *   import { t, locale, LOCALES, LOCALE_LABELS } from '$lib/i18n';
 *   $t('nav.projects')
 *   $t('greeting', { name: 'Alice' })  // "Hello, Alice"
 */

import { writable, derived, type Readable } from 'svelte/store';
import en from './locales/en.json';
import es from './locales/es.json';
import fr from './locales/fr.json';
import de from './locales/de.json';
import ja from './locales/ja.json';
import zh from './locales/zh.json';

export type Locale = 'en' | 'es' | 'fr' | 'de' | 'ja' | 'zh';

export const LOCALES: readonly Locale[] = ['en', 'es', 'fr', 'de', 'ja', 'zh'] as const;

export const LOCALE_LABELS: Record<Locale, string> = {
  en: 'English',
  es: 'Español',
  fr: 'Français',
  de: 'Deutsch',
  ja: '日本語',
  zh: '中文'
};

export const LOCALE_FLAGS: Record<Locale, string> = {
  en: '🇬🇧',
  es: '🇪🇸',
  fr: '🇫🇷',
  de: '🇩🇪',
  ja: '🇯🇵',
  zh: '🇨🇳'
};

export const LOCALE_NATIVE_NUMERALS: Record<Locale, string> = {
  en: '0-9',
  es: '0-9',
  fr: '0-9',
  de: '0-9',
  ja: '0-9 / 一二三',
  zh: '0-9 / 一二三'
};

const messages: Record<Locale, Record<string, unknown>> = {
  en,
  es,
  fr,
  de,
  ja,
  zh
};

const STORAGE_KEY = 'byteport-locale';
const isBrowser = typeof window !== 'undefined';

/** Persisted locale store. SSR defaults to 'en'. Browser hydrates from localStorage or navigator.language. */
export const locale = writable<Locale>('en');

if (isBrowser) {
  const saved = localStorage.getItem(STORAGE_KEY) as Locale | null;
  const browser = (navigator.language || 'en').split('-')[0] as Locale;
  const initial: Locale = saved && LOCALES.includes(saved)
    ? saved
    : LOCALES.includes(browser) ? browser : 'en';
  locale.set(initial);
}

locale.subscribe((l) => {
  if (isBrowser) {
    try {
      localStorage.setItem(STORAGE_KEY, l);
      document.documentElement.lang = l;
      document.documentElement.dir = 'ltr';
    } catch {
      // localStorage may be unavailable in private browsing; ignore.
    }
  }
});

function resolveKey(msgs: Record<string, unknown>, key: string): string | undefined {
  const parts = key.split('.');
  let cur: unknown = msgs;
  for (const p of parts) {
    if (cur && typeof cur === 'object' && p in (cur as Record<string, unknown>)) {
      cur = (cur as Record<string, unknown>)[p];
    } else {
      return undefined;
    }
  }
  return typeof cur === 'string' ? cur : undefined;
}

/** Replaces {name}, {count}, etc. with params[k]. Missing keys stay literal so bugs surface in UI. */
function interpolate(template: string, params?: Record<string, string | number>): string {
  if (!params) return template;
  return template.replace(/\{(\w+)\}/g, (_, k: string) => {
    if (k in params) return String(params[k]);
    return `{${k}}`;
  });
}

/**
 * Translation function store.
 *
 * Usage in components: `{$t('nav.projects')}` — Svelte auto-subscribes to
 * stores prefixed with `$`. Falls back to English, then to the key itself
 * (so missing translations are visible in dev, not silent in prod).
 */
export const t: Readable<(key: string, params?: Record<string, string | number>) => string> =
  derived(locale, ($locale) => {
    return (key: string, params?: Record<string, string | number>) => {
      const fromLocale = resolveKey(messages[$locale], key);
      const value = fromLocale ?? resolveKey(messages.en, key);
      if (value === undefined) {
        if (isBrowser && import.meta.env.DEV) {
          console.warn(`[i18n] Missing key: "${key}" in locale "${$locale}"`);
        }
        return key;
      }
      return interpolate(value, params);
    };
  });

/** Format a date in the active locale via Intl.DateTimeFormat. */
export function formatDate(d: Date | string | number, opts: Intl.DateTimeFormatOptions = {}): string {
  const date = d instanceof Date ? d : new Date(d);
  return new Intl.DateTimeFormat($locale_safe(), opts).format(date);
}

/** Format a number in the active locale via Intl.NumberFormat. */
export function formatNumber(n: number, opts: Intl.NumberFormatOptions = {}): string {
  return new Intl.NumberFormat($locale_safe(), opts).format(n);
}

/** SSR-safe read of the current locale (defaults to 'en' on server). */
export function $locale_safe(): Locale {
  let cur: Locale = 'en';
  const unsub = locale.subscribe((l) => { cur = l; });
  unsub();
  return cur;
}

/** Manually set the active locale (also persists to localStorage in browser). */
export function setLocale(l: Locale) {
  if (!LOCALES.includes(l)) return;
  locale.set(l);
}