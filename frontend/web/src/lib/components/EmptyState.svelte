<script lang="ts">
	/**
	 * Empty-state component with optional illustration slot.
	 *
	 * Pillar mapping (PILLAR-TAXONOMY-v2.md v2.2 §L73 Empty States):
	 * - Bronze: helpful empty states with title + body + action slot
	 * - Silver: illustration slot accepts mascot/SVG/asset
	 * - Gold:   actionable + sample-prompt via the `primaryAction` prop
	 *
	 * Designed to be wrapped in any list/grid that may be empty.
	 */
	import { onMount } from 'svelte';
	import { spring } from 'svelte/motion';
	import { cubicOut } from 'svelte/easing';

	export let title: string;
	export let description: string;
	export let illustration: 'no-data' | 'no-results' | 'error' | 'mascot' | null = null;
	export let primaryAction: { label: string; href?: string; onClick?: () => void } | null = null;
	export let secondaryAction: { label: string; href?: string; onClick?: () => void } | null =
		null;
	export let reducedMotion: boolean = false;

	const enterScale = spring(0.94, { stiffness: 0.18, damping: 0.7 });
	const enterOpacity = spring(0, { stiffness: 0.18, damping: 0.85 });

	onMount(() => {
		enterScale.set(1);
		enterOpacity.set(1);
	});
</script>

<div
	class="empty-state"
	role="status"
	aria-live="polite"
	style="
    transform: scale({$enterScale});
    opacity: {$enterOpacity};
  "
>
	{#if illustration}
		<div class="illustration" aria-hidden="true">
			{#if illustration === 'mascot'}
				<img src="/brand/mascot.svg" alt="" />
			{:else if illustration === 'no-data'}
				<svg viewBox="0 0 240 160" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
					<rect
						x="20"
						y="40"
						width="200"
						height="100"
						rx="12"
						fill="none"
						stroke="currentColor"
						stroke-opacity="0.25"
						stroke-width="2"
						stroke-dasharray="6 6"
					/>
					<line
						x1="40"
						y1="70"
						x2="120"
						y2="70"
						stroke="currentColor"
						stroke-opacity="0.4"
						stroke-width="2"
						stroke-linecap="round"
					/>
					<line
						x1="40"
						y1="90"
						x2="160"
						y2="90"
						stroke="currentColor"
						stroke-opacity="0.3"
						stroke-width="2"
						stroke-linecap="round"
					/>
					<line
						x1="40"
						y1="110"
						x2="100"
						y2="110"
						stroke="currentColor"
						stroke-opacity="0.2"
						stroke-width="2"
						stroke-linecap="round"
					/>
					<circle
						cx="190"
						cy="115"
						r="14"
						fill="none"
						stroke="currentColor"
						stroke-opacity="0.4"
						stroke-width="2"
					/>
					<line
						x1="200"
						y1="125"
						x2="210"
						y2="135"
						stroke="currentColor"
						stroke-opacity="0.4"
						stroke-width="2"
						stroke-linecap="round"
					/>
				</svg>
			{:else if illustration === 'no-results'}
				<svg viewBox="0 0 240 160" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
					<circle
						cx="100"
						cy="70"
						r="40"
						fill="none"
						stroke="currentColor"
						stroke-opacity="0.35"
						stroke-width="3"
					/>
					<line
						x1="130"
						y1="100"
						x2="170"
						y2="140"
						stroke="currentColor"
						stroke-opacity="0.5"
						stroke-width="4"
						stroke-linecap="round"
					/>
					<line
						x1="80"
						y1="60"
						x2="120"
						y2="80"
						stroke="currentColor"
						stroke-opacity="0.2"
						stroke-width="2"
						stroke-linecap="round"
					/>
				</svg>
			{:else if illustration === 'error'}
				<svg viewBox="0 0 240 160" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
					<circle
						cx="120"
						cy="80"
						r="48"
						fill="none"
						stroke="currentColor"
						stroke-opacity="0.4"
						stroke-width="3"
					/>
					<line
						x1="100"
						y1="60"
						x2="140"
						y2="100"
						stroke="currentColor"
						stroke-opacity="0.55"
						stroke-width="4"
						stroke-linecap="round"
					/>
					<line
						x1="140"
						y1="60"
						x2="100"
						y2="100"
						stroke="currentColor"
						stroke-opacity="0.55"
						stroke-width="4"
						stroke-linecap="round"
					/>
				</svg>
			{/if}
		</div>
	{/if}

	<h3 class="title">{title}</h3>
	<p class="description">{description}</p>

	{#if primaryAction || secondaryAction}
		<div class="actions">
			{#if primaryAction}
				{#if primaryAction.href}
					<a class="btn primary" href={primaryAction.href}>{primaryAction.label}</a>
				{:else}
					<button type="button" class="btn primary" on:click={primaryAction.onClick}
						>{primaryAction.label}</button
					>
				{/if}
			{/if}
			{#if secondaryAction}
				{#if secondaryAction.href}
					<a class="btn secondary" href={secondaryAction.href}>{secondaryAction.label}</a>
				{:else}
					<button type="button" class="btn secondary" on:click={secondaryAction.onClick}
						>{secondaryAction.label}</button
					>
				{/if}
			{/if}
		</div>
	{/if}
</div>

<style>
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		padding: 3rem 1.5rem;
		color: var(--text-muted, #6b7280);
		gap: 0.75rem;
	}
	.illustration {
		width: 240px;
		height: 160px;
		color: var(--accent, #5a7dff);
		margin-bottom: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.illustration img {
		width: 80%;
		height: auto;
	}
	.title {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text, #111827);
		margin: 0;
	}
	.description {
		font-size: 0.95rem;
		color: var(--text-muted, #6b7280);
		max-width: 36ch;
		margin: 0;
	}
	.actions {
		display: flex;
		gap: 0.75rem;
		margin-top: 1rem;
		flex-wrap: wrap;
		justify-content: center;
	}
	.btn {
		appearance: none;
		border: 1px solid transparent;
		border-radius: 0.5rem;
		padding: 0.55rem 1rem;
		font-size: 0.9rem;
		font-weight: 500;
		cursor: pointer;
		transition:
			background 180ms cubic-bezier(0.2, 0.8, 0.2, 1),
			transform 180ms,
			box-shadow 180ms;
		text-decoration: none;
	}
	.btn.primary {
		background: var(--accent, #5a7dff);
		color: var(--accent-foreground, #fff);
	}
	.btn.primary:hover {
		background: var(--accent-hover, #4a6dff);
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(90, 125, 255, 0.25);
	}
	.btn.secondary {
		background: transparent;
		color: var(--text, #111827);
		border-color: var(--border, rgba(0, 0, 0, 0.12));
	}
	.btn.secondary:hover {
		background: var(--surface-hover, rgba(0, 0, 0, 0.04));
	}
	.btn:focus-visible {
		outline: 2px solid var(--accent, #5a7dff);
		outline-offset: 2px;
	}

	@media (prefers-reduced-motion: reduce) {
		.empty-state,
		.btn {
			transition: none !important;
		}
	}
</style>
