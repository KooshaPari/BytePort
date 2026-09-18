<script lang="ts">
	/**
	 * Settings index.
	 *
	 * Was a shell with an empty body: the sidebar pointed "Integrations" at the
	 * profile route, and the main pane rendered nothing at all. It now indexes
	 * the two settings areas as grouped rows, each with a one-line description,
	 * and states the signed-in identity so the page is never a blank pane.
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, initializeUser } from '../../../stores/user';
	import { getApiBaseUrl } from '$lib/api';

	interface SettingsRow {
		label: string;
		description: string;
		href: string;
	}

	// Grouped deliberately: identity in one group, outbound connections in the
	// other. Order matches how often each is opened.
	const groups: { title: string; rows: SettingsRow[] }[] = [
		{
			title: 'Your account',
			rows: [
				{
					label: 'Profile',
					description: 'Display name, email address and password.',
					href: '/home/settings/profile'
				}
			]
		},
		{
			title: 'Connections',
			rows: [
				{
					label: 'Integrations',
					description: 'GitHub, AWS, AI provider keys and the portfolio endpoint.',
					href: '/home/settings/integrations'
				}
			]
		}
	];

	const sectionLabel =
		'px-1 text-[11px] font-medium tracking-wide text-dark-onSurfaceVariant uppercase';

	onMount(() => {
		const unsubscribe = user.subscribe((value) => {
			if (value.status === 'unauthenticated') {
				goto('/login');
			}
		});
		void initializeUser(getApiBaseUrl());
		return unsubscribe;
	});
</script>

<div class="mx-auto max-w-3xl space-y-6">
	{#each groups as group (group.title)}
		<section class="space-y-2">
			<h2 class={sectionLabel}>{group.title}</h2>
			<div class="border-border bg-dark-surfaceContainer overflow-hidden rounded-lg border">
				{#each group.rows as row (row.href)}
					<a
						href={row.href}
						class="border-border/60 hover:bg-dark-surfaceContainerHigh flex items-center justify-between gap-4 border-b px-4 py-3 transition-colors last:border-b-0"
					>
						<span class="min-w-0">
							<span class="text-dark-onSurface block text-[13px] font-medium"
								>{row.label}</span
							>
							<span
								class="text-dark-onSurfaceVariant mt-0.5 block text-[12px] leading-snug"
							>
								{row.description}
							</span>
						</span>
						<svg
							class="text-dark-onSurfaceVariant h-4 w-4 shrink-0"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.75"
							stroke-linecap="round"
							stroke-linejoin="round"
							aria-hidden="true"
						>
							<path d="m9 18 6-6-6-6" />
						</svg>
					</a>
				{/each}
			</div>
		</section>
	{/each}
</div>
