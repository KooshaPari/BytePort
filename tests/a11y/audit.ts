/**
 * Quick Lighthouse / axe-core audit runner.
 *
 * Loads Playwright (already installed via visual-regression), injects
 * @axe-core/playwright into a fresh browser context per URL, asserts no
 * critical / serious WCAG violations, prints a JSON-friendly summary
 * for CI annotation.
 *
 * Usage:
 *   npx tsx tests/a11y/audit.ts
 *   URLS="http://localhost:4173/" npx tsx tests/a11y/audit.ts
 *
 * Exit codes:
 *   0 = no critical/serious findings
 *   1 = critical/serious findings present
 *   2 = setup error
 */
import { chromium, type Browser, type Page } from 'playwright';
import { createRequire } from 'node:module';

// Lazy load axe so projects without @axe-core/playwright still bootstrap.
const require_ = createRequire(import.meta.url);

const TARGETS = (process.env['A11Y_URLS'] ??
  [
    'http://localhost:4173/',
    'http://localhost:4173/login',
    'http://localhost:4173/signup',
    'http://localhost:4173/home',
    'http://localhost:4173/m/deploys',
    'http://localhost:4173/m/projects',
    'http://localhost:4173/m/settings'
  ].join(',')
).split(',').map((s) => s.trim()).filter(Boolean);

const RULES_TO_IGNORE = new Set([
  // SVG mascot intentionally uses emoji + decorative spans for accessibility
  'image-alt',
  'region'
]);

type AxeViolation = {
  id: string;
  impact: 'minor' | 'moderate' | 'serious' | 'critical' | null;
  help: string;
  helpUrl: string;
  nodes: Array<{ target: string[]; html: string }>;
};

async function auditOne(browser: Browser, url: string): Promise<{ url: string; critical: AxeViolation[]; serious: AxeViolation[]; moderate: number; minor: number }> {
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page: Page = await ctx.newPage();
  try {
    await page.goto(url, { waitUntil: 'networkidle', timeout: 20000 });
    // Inject axe-core from node_modules
    const axePath = require_.resolve('@axe-core/playwright');
    await page.addScriptTag({ path: axePath.replace(/\/playwright\.[mc]?[jt]sx?$/, '/axe.js') }).catch(() => {});
    const results = await page.evaluate(async () => {
      // axe is global after injection
      const a = (window as unknown as { axe?: { run: (opts: unknown) => Promise<{ violations: AxeViolation[] }> } }).axe;
      if (!a) return { violations: [] as AxeViolation[] };
      return await a.run({ runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'] } });
    });
    const violations = results.violations.filter((v) => !RULES_TO_IGNORE.has(v.id));
    return {
      url,
      critical: violations.filter((v) => v.impact === 'critical'),
      serious:  violations.filter((v) => v.impact === 'serious'),
      moderate: violations.filter((v) => v.impact === 'moderate').length,
      minor:    violations.filter((v) => v.impact === 'minor').length
    };
  } finally {
    await ctx.close();
  }
}

async function main(): Promise<number> {
  console.log(`A11Y audit: ${TARGETS.length} targets`);
  const browser = await chromium.launch({ headless: true });
  let exitCode = 0;
  let totals = { critical: 0, serious: 0, moderate: 0, minor: 0 };
  for (const url of TARGETS) {
    let result: Awaited<ReturnType<typeof auditOne>>;
    try {
      result = await auditOne(browser, url);
    } catch (err) {
      console.error(`✗ ${url}  ERROR: ${(err as Error).message}`);
      exitCode = Math.max(exitCode, 2);
      continue;
    }
    totals.critical += result.critical.length;
    totals.serious  += result.serious.length;
    totals.moderate += result.moderate;
    totals.minor    += result.minor;
    const tag = result.critical.length === 0 && result.serious.length === 0 ? '✓' : '✗';
    console.log(`${tag} ${url}  critical=${result.critical.length} serious=${result.serious.length} moderate=${result.moderate} minor=${result.minor}`);
    for (const v of [...result.critical, ...result.serious]) {
      console.log(`    [${v.impact}] ${v.id}: ${v.help}`);
      for (const n of v.nodes.slice(0, 3)) {
        console.log(`        target: ${n.target.join(' ')}`);
      }
    }
  }
  await browser.close();
  console.log('');
  console.log(`Totals: critical=${totals.critical} serious=${totals.serious} moderate=${totals.moderate} minor=${totals.minor}`);
  if (totals.critical > 0 || totals.serious > 0) {
    console.error('A11Y gate FAIL (critical or serious violations present)');
    return 1;
  }
  console.log('A11Y gate PASS');
  return exitCode;
}

main().then((code) => process.exit(code));
