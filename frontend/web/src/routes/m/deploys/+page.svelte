<script lang="ts">
	/**
	 * Mobile Deploys list.
	 * - Compact card layout, one row per deploy
	 * - Status pill (color-coded) using i18n keys projects.status.*
	 * - Swipe-left → delete; swipe-right → redeploy (action hint visible on overscroll)
	 * - Pull-to-refresh
	 *
	 * In-flight data is mocked; the layout is the deliverable.
	 * PILLAR-TAXONOMY-v2.md v2.2 §L43 (mobile gestures + responsive)
	 */
	import { t } from '$lib/i18n';

	type Status = 'running' | 'stopped' | 'building' | 'failed' | 'deploying';
	type Deploy = {
		id: string;
		project: string;
		branch: string;
		status: Status;
		startedAt: number;
		durationSec: number;
	};

	let deploys = $state<Deploy[]>([
		{
			id: 'd-1041',
			project: 'byteport-web',
			branch: 'main',
			status: 'running',
			startedAt: Date.now() - 60_000,
			durationSec: 60
		},
		{
			id: 'd-1040',
			project: 'api-gateway',
			branch: 'fix/auth',
			status: 'building',
			startedAt: Date.now() - 90_000,
			durationSec: 90
		},
		{
			id: 'd-1039',
			project: 'byteport-web',
			branch: 'feat/l43-mob',
			status: 'failed',
			startedAt: Date.now() - 1_800_000,
			durationSec: 12
		},
		{
			id: 'd-1038',
			project: 'docs-site',
			branch: 'main',
			status: 'stopped',
			startedAt: Date.now() - 7_200_000,
			durationSec: 240
		},
		{
			id: 'd-1037',
			project: 'api-gateway',
			branch: 'main',
			status: 'deploying',
			startedAt: Date.now() - 30_000,
			durationSec: 30
		}
	]);

	// ── Pull-to-refresh ───────────────────────────────────────────────────────
	let listEl: HTMLDivElement | undefined = $state();
	let pullDist = $state(0);
	let refreshing = $state(false);
	let touchStartY = 0;
	function onTouchStart(e: TouchEvent) {
		if (listEl && listEl.scrollTop <= 0) touchStartY = e.touches[0].clientY;
	}
	function onTouchMove(e: TouchEvent) {
		if (touchStartY === 0 || refreshing) return;
		const dy = e.touches[0].clientY - touchStartY;
		if (dy > 0 && listEl && listEl.scrollTop <= 0) {
			pullDist = Math.min(dy * 0.4, 90);
			if (pullDist > 0) e.preventDefault();
		}
	}
	async function onTouchEnd() {
		touchStartY = 0;
		if (pullDist > 60 && !refreshing) {
			refreshing = true;
			pullDist = 60;
			await new Promise((r) => setTimeout(r, 800));
			refreshing = false;
		}
		pullDist = 0;
	}

	// ── Swipe-to-action ───────────────────────────────────────────────────────
	function swipeOffset(id: string): number {
		return swipes[id] ?? 0;
	}
	let swipes = $state<Record<string, number>>({});
	let activeSwipe = '';
	function onRowTouchStart(e: TouchEvent, id: string) {
		activeSwipe = id;
		rowStartX = e.touches[0].clientX;
		rowStartVal = swipes[id] ?? 0;
	}
	let rowStartX = 0;
	let rowStartVal = 0;
	function onRowTouchMove(e: TouchEvent, id: string) {
		if (activeSwipe !== id) return;
		const dx = e.touches[0].clientX - rowStartX;
		const next = Math.max(-96, Math.min(96, rowStartVal + dx));
		swipes[id] = next;
	}
	function onRowTouchEnd(id: string) {
		const v = swipes[id] ?? 0;
		if (v < -48) deleteAction(id);
		else if (v > 48) redeployAction(id);
		swipes[id] = 0;
		activeSwipe = '';
	}
	function deleteAction(id: string) {
		deploys = deploys.filter((d) => d.id !== id);
	}
	function redeployAction(id: string) {
		deploys = deploys.map((d) =>
			d.id === id ? { ...d, status: 'deploying', startedAt: Date.now(), durationSec: 0 } : d
		);
	}

	function statusPillClass(s: Status): string {
		return `pill pill-${s}`;
	}
	function fmtAgo(ts: number): string {
		const s = Math.max(1, Math.round((Date.now() - ts) / 1000));
		if (s < 60) return `${s}s`;
		const m = Math.round(s / 60);
		if (m < 60) return `${m}m`;
		const h = Math.round(m / 60);
		if (h < 24) return `${h}h`;
		return `${Math.round(h / 24)}d`;
	}
</script>

<svelte:head><title>{$t('nav.deploys')} · BytePort</title></svelte:head>

<section class="page">
	<header class="ph">
		<h1>{$t('nav.deploys')}</h1>
		<span class="count">{deploys.length}</span>
	</header>

	<div
		class="list"
		class:pulling={pullDist > 0}
		bind:this={listEl}
		ontouchstart={onTouchStart}
		ontouchmove={onTouchMove}
		ontouchend={onTouchEnd}
	>
		{#if pullDist > 0 || refreshing}
			<div class="ptr" style="height:{pullDist}px">
				{#if refreshing}
					<span class="spin" aria-hidden="true"></span>
					<span>{$t('common.loading')}</span>
				{:else}
					<span aria-hidden="true">↓</span>
					<span>{$t('common.retry')}</span>
				{/if}
			</div>
		{/if}

		{#each deploys as d (d.id)}
			{@const off = swipeOffset(d.id)}
			<div class="row">
				<div class="swipe-bg delete" aria-hidden="true">{$t('common.delete')}</div>
				<div class="swipe-bg redeploy" aria-hidden="true">{$t('projects.deploy')}</div>
				<article
					class="card"
					style="transform:translateX({off}px)"
					ontouchstart={(e) => onRowTouchStart(e, d.id)}
					ontouchmove={(e) => onRowTouchMove(e, d.id)}
					ontouchend={() => onRowTouchEnd(d.id)}
				>
					<div class="meta">
						<span class="proj">{d.project}</span>
						<span class={statusPillClass(d.status)}>
							<span class="dot" aria-hidden="true"></span>
							{$t(`projects.status.${d.status}`)}
						</span>
					</div>
					<div class="sub">
						<code>{d.branch}</code>
						<span class="ago">{fmtAgo(d.startedAt)} ago</span>
					</div>
				</article>
			</div>
		{:else}
			<p class="empty">{$t('emptyStates.noData.title')}</p>
		{/each}
	</div>
</section>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.ph {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
	}
	.ph h1 {
		font-size: 1.25rem;
		margin: 0;
	}
	.count {
		background: var(--muted, #101418);
		color: var(--muted-foreground, #bec9c7);
		padding: 2px 8px;
		border-radius: 999px;
		font-size: 0.75rem;
	}

	.list {
		flex: 1;
		overflow-y: auto;
		overscroll-behavior: contain;
	}
	.ptr {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		color: var(--muted-foreground, #bec9c7);
		font-size: 0.85rem;
		transition: height 120ms ease-out;
	}
	.spin {
		width: 14px;
		height: 14px;
		border-radius: 50%;
		border: 2px solid var(--border, #3f4948);
		border-top-color: var(--primary, #80d5cf);
		animation: spin 0.8s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.row {
		position: relative;
		margin-bottom: 8px;
	}
	.swipe-bg {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		padding: 0 1rem;
		font-size: 0.85rem;
		font-weight: 600;
		color: #fff;
		border-radius: 12px;
	}
	.swipe-bg.delete {
		background: var(--destructive, #ffb4ab);
		color: #690005;
		justify-content: flex-end;
	}
	.swipe-bg.redeploy {
		background: var(--primary, #80d5cf);
		color: #003734;
	}
	.card {
		position: relative;
		background: var(--card, #0e1514);
		border: 1px solid var(--border, #3f4948);
		border-radius: 12px;
		padding: 0.75rem 1rem;
		min-height: 56px;
		z-index: 1;
		transition: transform 80ms;
	}
	.meta {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 0.5rem;
	}
	.proj {
		font-weight: 600;
		color: var(--foreground, #dde4e2);
	}
	.sub {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 4px;
		font-size: 0.85rem;
		color: var(--muted-foreground, #bec9c7);
	}
	.sub code {
		background: var(--muted, #101418);
		padding: 1px 6px;
		border-radius: 6px;
	}

	.pill {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 8px;
		border-radius: 999px;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.pill .dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
	}
	.pill-running {
		background: rgba(128, 213, 207, 0.15);
		color: #80d5cf;
	}
	.pill-running .dot {
		background: #80d5cf;
	}
	.pill-building {
		background: rgba(131, 210, 227, 0.15);
		color: #83d2e3;
	}
	.pill-building .dot {
		background: #83d2e3;
		animation: blink 1s infinite;
	}
	.pill-deploying {
		background: rgba(155, 203, 251, 0.15);
		color: #9bcbfb;
	}
	.pill-deploying .dot {
		background: #9bcbfb;
		animation: blink 1s infinite;
	}
	.pill-stopped {
		background: rgba(190, 201, 199, 0.12);
		color: #bec9c7;
	}
	.pill-stopped .dot {
		background: #bec9c7;
	}
	.pill-failed {
		background: rgba(255, 180, 171, 0.18);
		color: #ffb4ab;
	}
	.pill-failed .dot {
		background: #ffb4ab;
	}
	@keyframes blink {
		50% {
			opacity: 0.4;
		}
	}

	.empty {
		color: var(--muted-foreground, #bec9c7);
		text-align: center;
		padding: 2rem 1rem;
		margin: 0;
	}

	@media (min-width: 640px) {
		.page {
			max-width: 760px;
			margin: 0 auto;
		}
		.card {
			padding: 1rem 1.25rem;
		}
	}
</style>
