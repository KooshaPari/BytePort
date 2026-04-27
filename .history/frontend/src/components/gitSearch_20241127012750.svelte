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

			const rawData = await response.text();
			console.log('rawData:', rawData);
			let data: Repository[] = JSON.parse(rawData) as Repository[];

			console.log('data:', data);

			return data;
		} catch (error) {
			console.error('Failed to fetch repositories:', error);
			return [];
		}
	}

	async function getRepos() {
		const testJson = '[{"id":855465094,"node_id":"R_kgDOMv1chg","name":"BookstoreApp","full_name":"CSE-360-Group-9-Bookstore/BookstoreApp","private":false,"owner":{"login":"CSE-360-Group-9-Bookstore","id":186760737,"node_id":"O_kgDOCyG-IQ","avatar_url":"https://avatars.githubusercontent.com/u/186760737?v=4","gravatar_id":"","url":"https://api.github.com/users/CSE-360-Group-9-Bookstore","html_url":"https://github.com/CSE-360-Group-9-Bookstore","followers_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/followers","following_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/following{/other_user}","gists_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/gists{/gist_id}","starred_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/subscriptions","organizations_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/orgs","repos_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/repos","events_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/events{/privacy}","received_events_url":"https://api.github.com/users/CSE-360-Group-9-Bookstore/...(line too long; chars omitted)';
		
		try {
			userRepos = JSON.parse(testJson) as Repository[];
			console.log('Parsed repositories:', userRepos);
		} catch (error) {
			console.error('Failed to parse repository data:', error);
			userRepos = [];
		}
	}
	function testPart(test: String) {
		let data = JSON.stringify({ test: test });
		let newdata = JSON.stringify({ repo: .data });
		let arr = JSON.parse(data);
		console.log('data:', data);
		console.log('newdata:', newdata);
		console.log('arr:', arr);
		let reps = arr as Repository[];
		console.log('reps:', reps);
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
