<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
	import type { User } from '../stores/user.ts';
	import { user } from '../stores/user';
	import { onMount, onDestroy } from 'svelte';
	import type { Repository } from '../lib/git';

	let userRepos: Repository[];
	let client: User | null = null;

	// read in user store and set Client
	const unsubscribe = user.subscribe((value) => {
		client = value.data;
	});
	onMount(async () => {
		await getRepos();
	});
	async function fetchUserRepositories(): Promise<Repository[]> {
		try {
			const response = await fetch('http://localhost:8081/api/github/repositories', {
				method: 'GET',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`Error fetching repositories: ${response.statusText}`);
			}

			const rawData = await response.json();
			let data = JSON.parse(json.stringify(rawData));
			// Parse and map the raw JSON into Repository objects
			const repositories: Repository[] = data.map((repo: any): Repository => {
				return {
					id: repo.id,
					node_id: repo.node_id,
					name: repo.name,
					full_name: repo.full_name,
					private: repo.private,
					owner: {
						login: repo.owner?.login,
						id: repo.owner?.id,
						node_id: repo.owner?.node_id,
						avatar_url: repo.owner?.avatar_url,
						html_url: repo.owner?.html_url,
						type: repo.owner?.type,
						site_admin: repo.owner?.site_admin
					},
					html_url: repo.html_url,
					description: repo.description,
					fork: repo.fork,
					created_at: repo.created_at,
					updated_at: repo.updated_at,
					pushed_at: repo.pushed_at,
					ssh_url: repo.ssh_url,
					clone_url: repo.clone_url,
					svn_url: repo.svn_url,
					homepage: repo.homepage,
					size: repo.size,
					stargazers_count: repo.stargazers_count,
					watchers_count: repo.watchers_count,
					language: repo.language,
					has_issues: repo.has_issues,
					has_projects: repo.has_projects,
					has_downloads: repo.has_downloads,
					has_wiki: repo.has_wiki,
					has_pages: repo.has_pages,
					has_discussions: repo.has_discussions,
					forks_count: repo.forks_count,
					mirror_url: repo.mirror_url,
					archived: repo.archived,
					disabled: repo.disabled,
					open_issues_count: repo.open_issues_count,
					license: repo.license, // Or transform if detailed license object is required
					allow_forking: repo.allow_forking,
					is_template: repo.is_template,
					web_commit_signoff_required: repo.web_commit_signoff_required,
					topics: repo.topics || [],
					visibility: repo.visibility,
					forks: repo.forks,
					open_issues: repo.open_issues,
					watchers: repo.watchers,
					default_branch: repo.default_branch,
					permissions: {
						admin: repo.permissions?.admin,
						maintain: repo.permissions?.maintain,
						push: repo.permissions?.push,
						triage: repo.permissions?.triage,
						pull: repo.permissions?.pull
					}
				};
			});

			console.log('userRepos:', repositories);
			return repositories;
		} catch (error) {
			console.error('Failed to fetch repositories:', error);
			return [];
		}
	}

	async function getRepos() {
		userRepos = await fetchUserRepositories();
	}
</script>

<Command.Root>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Repositories">
			{#each userRepos as repo}
				<Command.Item>
					<Icon class="me-4 ms-2" icon="fa:github" />
					{repo.name}
					{#if repo.private}
						<Icon class="me-4 ms-2" icon="fa:lock" />
					{:else}
						<Icon class="me-4 ms-2" icon="fa:globe" />
					{/if}
				</Command.Item>
			{/each}
		</Command.Group>
	</Command.List>
</Command.Root>
