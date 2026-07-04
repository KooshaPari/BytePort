/**
 * Locale-aware formatters.
 *
 * Centralizes Intl helpers so route components can call `formatDate()` /
 * `formatNumber()` without worrying about the active locale. Works in SSR
 * too: defaults to 'en' on the server.
 */

import { locale } from './index';

/** Current locale value, updated reactively when $locale changes. */
let active: 'en' | 'es' | 'fr' | 'de' | 'ja' | 'zh' = 'en';
locale.subscribe((l) => (active = l));

const DATE_DEFAULT: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: 'short',
  day: 'numeric',
};

const TIME_DEFAULT: Intl.DateTimeFormatOptions = {
  hour: '2-digit',
  minute: '2-digit',
};

const NUMBER_DEFAULT: Intl.NumberFormatOptions = {
  maximumFractionDigits: 2,
};

const RELATIVE_UNITS: Record<string, Intl.RelativeTimeFormatUnit> = {
  s: 'second',
  m: 'minute',
  h: 'hour',
  d: 'day',
  w: 'week',
  mo: 'month',
  y: 'year',
};

export function formatDate(
  input: Date | string | number,
  opts: Intl.DateTimeFormatOptions = DATE_DEFAULT,
): string {
  const d = input instanceof Date ? input : new Date(input);
  return new Intl.DateTimeFormat(active, opts).format(d);
}

export function formatTime(
  input: Date | string | number,
  opts: Intl.DateTimeFormatOptions = TIME_DEFAULT,
): string {
  return formatDate(input, opts);
}

export function formatDateTime(input: Date | string | number): string {
  return formatDate(input, { ...DATE_DEFAULT, ...TIME_DEFAULT });
}

export function formatNumber(n: number, opts: Intl.NumberFormatOptions = NUMBER_DEFAULT): string {
  return new Intl.NumberFormat(active, opts).format(n);
}

export function formatCurrency(
  cents: number,
  currency = 'USD',
  locale_override?: string,
): string {
  return new Intl.NumberFormat(locale_override ?? active, {
    style: 'currency',
    currency,
  }).format(cents / 100);
}

export function formatPercent(p: number, fractionDigits = 0): string {
  return new Intl.NumberFormat(active, {
    style: 'percent',
    maximumFractionDigits: fractionDigits,
  }).format(p);
}

/**
 * Relative time formatter ("3 minutes ago", "in 2 days").
 *
 * Pass a past Date and `now` (defaults to now); pass a future Date and
 * `future=true`. Granularity is chosen automatically.
 */
export function formatRelative(
  input: Date | string | number,
  options: { now?: Date | number; future?: boolean } = {},
): string {
  const target = input instanceof Date ? input : new Date(input);
  const now = options.now ? new Date(options.now) : new Date();
  const diff = (target.getTime() - now.getTime()) / 1000;
  const sign = diff >= 0 ? 1 : -1;
  const abs = Math.abs(diff);

  let unit: keyof typeof RELATIVE_UNITS;
  let value: number;
  if (abs < 60) [unit, value] = ['s', Math.round(abs)];
  else if (abs < 3600) [unit, value] = ['m', Math.round(abs / 60)];
  else if (abs < 86400) [unit, value] = ['h', Math.round(abs / 3600)];
  else if (abs < 604800) [unit, value] = ['d', Math.round(abs / 86400)];
  else if (abs < 2629800) [unit, value] = ['w', Math.round(abs / 604800)];
  else if (abs < 31557600) [unit, value] = ['mo', Math.round(abs / 2629800)];
  else [unit, value] = ['y', Math.round(abs / 31557600)];

  const future = options.future ?? diff >= 0;
  const adjustedValue = future ? value * sign : value * sign;
  return new Intl.RelativeTimeFormat(active, { numeric: 'auto' }).format(
    adjustedValue,
    RELATIVE_UNITS[unit],
  );
}
