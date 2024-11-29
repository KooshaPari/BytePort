<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import Icon from '@iconify/svelte';
	import type { User } from '../stores/user.ts';
	import { user } from '../stores/user';
	import { onMount, onDestroy } from 'svelte';
	import type { Repository } from '../lib/git';
	import RepoListItem from './repoListItem.svelte';
	import * as Avatar from '$lib/components/ui/avatar';

	let userRepos: Repository[];

	let client: User | null = null;
	const CACHE_KEY = 'user_repositories';
	const CACHE_DURATION = 1000 * 60 * 60;

	// read in user store and set Client
	async function fetchAndCacheRepos() {
		const repos = await fetchUserRepositories();
		userRepos = repos;
		localStorage.setItem(
			CACHE_KEY,
			JSON.stringify({
				timestamp: Date.now(),
				data: repos
			})
		);
	}

	function getCachedRepos() {
		const cached = localStorage.getItem(CACHE_KEY);
		if (cached) {
			const { timestamp, data } = JSON.parse(cached);
			if (Date.now() - timestamp < CACHE_DURATION) {
				return data;
			}
		}
		return null;
	}

	onMount(async () => {
		const cachedRepos = getCachedRepos();
		if (cachedRepos) {
			userRepos = cachedRepos;
		} else {
			await fetchAndCacheRepos();
		}
	});

	// Function to manually refresh repos
	async function refreshRepos() {
		await fetchAndCacheRepos();
	}
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
			console.log('rawData:', rawData);
			let data: Repository[] = rawData as Repository[];

			console.log('data:', data);

			return data;
		} catch (error) {
			console.error('Failed to fetch repositories:', error);
			return [];
		}
	}
</script>

<script lang="ts">
 import Check from "lucide-svelte/icons/check";
 import ChevronsUpDown from "lucide-svelte/icons/chevrons-up-down";
 import * as Command from "$lib/registry/default/ui/command/index.js";
 import * as Popover from "$lib/registry/default/ui/popover/index.js";
 import { Button } from "$lib/registry/default/ui/button/index.js";
 import { cn } from "$lib/utils.js";
 import { tick } from "svelte";
 
 const frameworks = [
  {
   value: "sveltekit",
   label: "SvelteKit",
  },
  {
   value: "next.js",
   label: "Next.js",
  },
  {
   value: "nuxt.js",
   label: "Nuxt.js",
  },
  {
   value: "remix",
   label: "Remix",
  },
  {
   value: "astro",
   label: "Astro",
  },
 ];
 
 let open = false;
 let value = "";
 
 $: selectedValue =
  frameworks.find((f) => f.value === value)?.label ??
  "Select a framework...";
 
 // We want to refocus the trigger button when the user selects
 // an item from the list so users can continue navigating the
 // rest of the form with the keyboard.
 function closeAndFocusTrigger(triggerId: string) {
  open = false;
  tick().then(() => {
   document.getElementById(triggerId)?.focus();
  });
 }
</script>
 
<Popover.Root bind:open let:ids>
 <Popover.Trigger asChild let:builder>
  <Button
   builders={[builder]}
   variant="outline"
   role="combobox"
   aria-expanded={open}
   class="w-[200px] justify-between"
  >
   {selectedValue}
   <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
  </Button>
 </Popover.Trigger>
 <Popover.Content class="w-[200px] p-0">
  <Command.Root>
   <Command.Input placeholder="Search framework..." />
   <Command.Empty>No framework found.</Command.Empty>
   <Command.Group>

    {#each userRepos as repo}
     <Command.Item
      value={repo.name}
      onSelect={(currentValue) => {
       value = currentValue;
       closeAndFocusTrigger(ids.trigger);
      }}
     >
      <Check
       class={cn(
        "mr-2 h-4 w-4",
        value !== framework.value && "text-transparent"
       )}
      />
      {framework.label}
     </Command.Item>
    {/each}
   </Command.Group>
  </Command.Root>
 </Popover.Content>
</Popover.Root>