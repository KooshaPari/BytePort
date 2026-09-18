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
	<dl class="divide-y divide-border">
		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
			>
				Repository
			</dt>
			<dd class="flex min-w-0 items-center gap-2">
				{#if repository}
					<span class="truncate text-[13px] text-dark-onSurface">{repository.full_name}</span>
					{#if repository.private}
						<Badge tone="neutral">Private</Badge>
					{/if}
				{:else}
					<span class="text-[13px] text-dark-onSurfaceVariant">Not selected</span>
				{/if}
			</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
			>
				Name
			</dt>
			<dd class="min-w-0 truncate text-[13px] text-dark-onSurface">{name || 'Untitled'}</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
			>
				Description
			</dt>
			<dd class="min-w-0 text-[13px] leading-snug text-dark-onSurface">
				{description || 'No description'}
			</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
			>
				Platform
			</dt>
			<dd class="text-[13px] text-dark-onSurface">{humanise(platform)}</dd>
		</div>

		<div class="flex items-start gap-3 px-3 py-2.5">
			<dt
				class="w-24 shrink-0 pt-0.5 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase"
			>
				Type
			</dt>
			<dd class="text-[13px] text-dark-onSurface">{humanise(type)}</dd>
		</div>
	</dl>
</Card>
