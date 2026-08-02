<script lang="ts">
	/**
	 * Mobile Projects list — hero header with brand, project cards with deploy
	 * count badge, FAB to create. Mirrors deploys/ touch patterns.
	 *
	 * PILLAR-TAXONOMY-v2.md v2.2 §L43
	 */
	import { t } from '$lib/i18n';

	type Project = {
		id: string;
		name: string;
		framework: 'sveltekit' | 'next' | 'go' | 'rust' | 'python';
		deployCount: number;
		lastDeployedAt: number;
	};

	let projects = $state<Project[]>([
		{
			id: 'p-1',
			name: 'byteport-web',
			framework: 'sveltekit',
			deployCount: 142,
			lastDeployedAt: Date.now() - 60_000
		},
		{
			id: 'p-2',
			name: 'api-gateway',
			framework: 'go',
			deployCount: 87,
			lastDeployedAt: Date.now() - 7_200_000
		},
		{
			id: 'p-3',
			name: 'docs-site',
			framework: 'sveltekit',
			deployCount: 34,
			lastDeployedAt: Date.now() - 86_400_000
		},
		{
			id: 'p-4',
			name: 'billing-svc',
			framework: 'rust',
			deployCount: 19,
			lastDeployedAt: Date.now() - 604_800_000
		}
	]);

	function frameworkColor(f: Project['framework']): string {
		return {
			sveltekit: 'badge-primary',
			next: 'badge-secondary',
			go: 'badge-info',
			rust: 'badge-warning',
			python: 'badge-success'
		}[f];
	}

	function relTime(ms: number): string {
		const diff = Date.now() - ms;
		if (diff < 60_000) return 'just now';
		if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m`;
		if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h`;
		return `${Math.floor(diff / 86_400_000)}d`;
	}
</script>

<header class="px-4 pt-4 pb-3 bg-base-200">
	<div class="flex items-center justify-between">
		<h1 class="text-xl font-semibold">{$t('nav.projects')}</h1>
		<span class="text-xs opacity-70">{projects.length}</span>
	</div>
</header>

<main class="px-4 py-3 space-y-3">
	{#each projects as p (p.id)}
		<article class="card card-compact bg-base-200 shadow-sm" aria-labelledby="p-{p.id}">
			<div class="card-body">
				<div class="flex items-start justify-between gap-2">
					<div class="flex-1 min-w-0">
						<h2 id="p-{p.id}" class="card-title text-base truncate">{p.name}</h2>
						<div class="flex items-center gap-2 mt-1">
							<span class="badge {frameworkColor(p.framework)} badge-sm"
								>{p.framework}</span
							>
							<span class="text-xs opacity-70"
								>{p.deployCount} {$t('projects.title').toLowerCase()}</span
							>
						</div>
					</div>
					<time
						class="text-xs opacity-70 shrink-0"
						datetime={new Date(p.lastDeployedAt).toISOString()}
					>
						{relTime(p.lastDeployedAt)}
					</time>
				</div>
			</div>
		</article>
	{/each}
</main>

<a
	href="/m/projects/new"
	aria-label={$t('projects.new')}
	class="fixed bottom-24 right-4 btn btn-primary btn-circle btn-lg shadow-lg"
>
	<svg viewBox="0 0 24 24" fill="none" class="h-6 w-6" aria-hidden="true">
		<path
			d="M12 5v14M5 12h14"
			stroke="currentColor"
			stroke-width="2.2"
			stroke-linecap="round"
		/>
	</svg>
</a>
