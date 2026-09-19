<script lang="ts">
	/**
	 * Step 3 of the new-project wizard: what is about to be deployed.
	 *
	 * Replaces the 2023 review card, which was a `Card` full of duplicated
	 * `Label` + `span` pairs under two headings both called "Repository", and
	 * crashed outright when no repository had been chosen (`repo.owner` was read
	 * without a guard).
	 *
	 * Values are read-only, so they are a definition list rather than a form:
	 * nothing here looks editable, which is the point of a review step.
	 */
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import type { Repository } from '$lib/git';

	let {
		name,
		description,
		type,
		platform,
		repository
	}: {
		name: string;
		description: string;
		type: string;
		platform: string;
		repository: Repository | null;
	} = $props();

	function humanise(value: string): string {
		return value.replace(/[-_]/g, ' ').replace(/^./, (character) => character.toUpperCase());
	}
</script>

<Card padding="none">
	<dl class="divide-border divide-y">
		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="text-dark-onSurfaceVariant w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide uppercase"
			>
				Repository
			</dt>
			<dd class="flex min-w-0 items-center gap-2">
				{#if repository}
					<span class="text-dark-onSurface truncate text-[13px]"
						>{repository.full_name}</span
					>
					{#if repository.private}
						<Badge tone="neutral">Private</Badge>
					{/if}
				{:else}
					<span class="text-dark-onSurfaceVariant text-[13px]">Not selected</span>
				{/if}
			</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="text-dark-onSurfaceVariant w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide uppercase"
			>
				Name
			</dt>
			<dd class="text-dark-onSurface min-w-0 truncate text-[13px]">{name || 'Untitled'}</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="text-dark-onSurfaceVariant w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide uppercase"
			>
				Description
			</dt>
			<dd class="text-dark-onSurface min-w-0 text-[13px] leading-snug">
				{description || 'No description'}
			</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="text-dark-onSurfaceVariant w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide uppercase"
			>
				Platform
			</dt>
			<dd class="text-dark-onSurface text-[13px]">{humanise(platform)}</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="text-dark-onSurfaceVariant w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide uppercase"
			>
				Type
			</dt>
			<dd class="text-dark-onSurface text-[13px]">{humanise(type)}</dd>
		</div>
	</dl>
</Card>
