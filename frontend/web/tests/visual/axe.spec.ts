/**
 * axe-core integration for BytePort's visual-regression CI workflow.
 *
 * PILLAR-TAXONOMY-v2.md v2.2 §L76 (accessibility).
 *
 * Runs @axe-core/playwright on every route fixture under src/routes/__visual
 * and writes a JSON + HTML report to .axe-reports/axe-{viewport}.{json,html}.
 *
 * Acceptance threshold: WCAG-2.1 AA — zero "serious" or "critical" violations
 * allowed on any route fixture. "moderate" + "minor" violations produce a
 * warning artifact for review but do not block.
 *
 * Why src/routes/__visual/*? Those are SvelteKit routes I gated to dev +
 * PLAYWRIGHT env (see __visual/+layout.ts). They're our snapshot + a11y
 * fixtures without polluting production routes.
 */
import AxeBuilder from '@axe-core/playwright';
import { test, expect } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';

const ROUTES = [
	{ path: '/__visual/splash', name: 'splash' },
	{ path: '/__visual/mascot', name: 'mascot' },
	{ path: '/__visual/empty-state', name: 'empty-state' }
];

for (const route of ROUTES) {
	test(`axe: WCAG-2.1 AA on ${route.name} (light + dark)`, async ({ page }, testInfo) => {
		const viewport = testInfo.project.name;
		const isDark = viewport.endsWith('-dark');

		if (isDark) {
			await page.emulateMedia({ colorScheme: 'dark' });
		} else {
			await page.emulateMedia({ colorScheme: 'light' });
		}

		await page.goto(route.path);

		// EmptyState uses a spring entrance animation.  Wait for the settled
		// state before measuring contrast; sampling its initial 0->1 opacity
		// produces blended colors and false WCAG failures.
		const emptyState = page.locator('.empty-state');
		if (await emptyState.count()) {
			await expect(emptyState).toHaveCSS('opacity', '1');
		}

		const results = await new AxeBuilder({ page })
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();

		mkdirSync('.axe-reports', { recursive: true });
		const base = `.axe-reports/axe-${route.name}-${viewport}`;
		writeFileSync(
			`${base}.json`,
			JSON.stringify(
				{
					route: route.path,
					viewport,
					colorScheme: isDark ? 'dark' : 'light',
					timestamp: new Date().toISOString(),
					violations: results.violations,
					passes: results.passes.length,
					incomplete: results.incomplete.length
				},
				null,
				2
			)
		);

		const serious = results.violations.filter((v) =>
			['serious', 'critical'].includes(v.impact ?? '')
		);

		// Write a comment-friendly markdown summary for the PR
		const md = [
			`# Axe accessibility report — ${route.name} (${viewport})`,
			``,
			`**Violations (serious/critical): ${serious.length}**`,
			`**Violations (all): ${results.violations.length}**`,
			`**Passes: ${results.passes.length}**`,
			`**Incomplete (needs review): ${results.incomplete.length}**`,
			``,
			...serious.map(
				(v) =>
					`- 🔴 **${v.id}** (${v.impact}): ${v.help}\n  - Affected: ${v.nodes.length} nodes\n  - Fix: ${v.helpUrl}`
			),
			...results.violations
				.filter((v) => !['serious', 'critical'].includes(v.impact ?? ''))
				.map((v) => `- 🟡 ${v.id} (${v.impact}): ${v.help}\n  - Fix: ${v.helpUrl}`)
		].join('\n');
		writeFileSync(`${base}.md`, md);

		expect(
			serious,
			`${serious.length} serious/critical a11y violations on ${route.path} (${viewport}). See ${base}.md for fix URLs.`
		).toEqual([]);
	});
}
