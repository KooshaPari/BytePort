<script lang="ts">
	import Icon from '@iconify/svelte';
	import Project from '../+layout.svelte';
	import VMInstance from '../+layout.svelte';
	import { onMount, onDestroy } from 'svelte';
	import type { User } from '../../stores/user';
	import { setUser, user } from '../../stores/user';
	import { goto } from '$app/navigation';
	let client: User | null = null;
	// convert to map name, '/name'
	const menuItemsMap = new Map<string, string>([
		['Projects', '/home/projects'],
		['Instances', '/home/instances'],
		['Monitor', '/home/monitor'],
		['Settings', '/home/settings']
	]);
	let projects: Project[];
	let instances: VMInstance[];
	onMount(() => {
		const unsubscribe = user.subscribe((value) => {
			client = value.data;
			//console.log('V: ', value);
			if (value.status != 'authenticated') {
				goto('/login');
			}
		});
		onDestroy(() => {
			unsubscribe();
		});
		populateLists();
	});

	async function populateLists() {
		let response = await fetch('http://localhost:8080/projects', {
			method: 'GET',
			headers: {
				'Content-Type': 'application/json'
			},
			credentials: 'include'
		});

		if (response.ok) {
			const data = await response.json();
			projects = data;
			console.log('Projects:', projects);
		} else {
			const errorData = await response.json();
			console.error('Error:', errorData);
		}
		response = await fetch('http://localhost:8080/instances', {
			method: 'GET',
			headers: {
				'Content-Type': 'application/json'
			},

			credentials: 'include'
		});
		if (response.ok) {
			const data = await response.json();
			instances = data;
			console.log('Instances:', instances);
		} else {
			const errorData = await response.json();
			console.error('Error:', errorData);
		}
	}
</script>

<div class="bg-dark-surface flex h-screen w-screen overflow-x-hidden" id="mainDashPar">
	<div
		class="h-5/5 flex-ro bg-dark-surfaceContainer w-1/5 items-center justify-center"
		id="sideBar"
	>
		<img
			on:click={() => goto('/home')}
			class="py-10"
			alt="BytePort"
			src="/src/assets/img/byte.png"
		/>
		<div id="sideBarProfileCont"></div>
		<ul class="" id="menuList">
			{#each [...menuItemsMap] as [key, value]}
				<li class=" text-md w-5/5 py-2 text-center text-white">
					<button
						class="hover:bg-dark-surfaceContainerHigh active:bg-dark-surfaceContainer active:text-dark-surfaceBright w-4/5 py-2 text-center transition-all hover:-translate-y-1
						hover:rounded-full active:translate-y-0.5"
						on:click={() => {
							goto(value);
						}}
					>
						{key}
					</button>
				</li>
			{/each}
		</ul>
	</div>

	<div id="body" class="w-4/5">
		<div
			id="header"
			class=" w-5/5 bg-dark-surfaceContainerLow h-1/5 flex-col justify-between ps-2.5"
		>
			<div id="headerNav" class="h-3/5 pt-2.5">
				<div class="flex justify-end pe-2.5" id="navRight">
					<Icon
						class="hover:text-dark-primary active:text-dark-surfaceBright mx-1 h-6 w-6 cursor-pointer text-white"
						icon="ic:baseline-notifications"
					/>
					<Icon
						class="hover:text-dark-primary active:text-dark-surfaceBright mx-1 h-6 w-6 cursor-pointer text-white"
						icon="ic:baseline-account-circle"
					/>
				</div>
			</div>
			<div id="headerContent" class="h-2/5 text-4xl text-white">Hello.</div>
		</div>
		<div id="mainBody">
			<div id="instanceSec" class="w-5/5 h-2/5 overflow-x-scroll">
				<h1 class="text-dark-secondary p-2">Instances</h1>
				<div id="instances" class="flex w-max overflow-y-visible p-2">
					<div
						class="bg-dark-surfaceContainerHigh text-dark-onSurface m-0.5 mx-1.5 flex h-64 w-48 items-center justify-center rounded-lg transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
					>
						<Icon
							class="bg-dark-primary text-dark-onPrimary w-4/5 rounded-full text-4xl"
							icon="ic:baseline-add"
						/>
					</div>
					{#each instances as instance}
						<div
							class="bg-dark-surfaceContainerHigh text-dark-onSurface m-0.5 mx-1.5 h-64 w-48 rounded-lg transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
						>
							<img
								src="/src/assets/img/byteport copy.png"
								alt="BytePort"
								class="h-5/5 bg-dark-surfaceContainerHighest p-2"
							/>
							<div
								class=" bg-dark-surfaceContainerHigh grid-flow-col grid-cols-2 px-2 pb-3 pt-1 text-sm"
							>
								<p>{instance.Name}</p>
								<p>{instance.Status}</p>
								<p>{instance.OS}</p>
								<p>{instance.Project}</p>
								<p>{instance.LastUpdated}</p>
							</div>
						</div>
					{/each}
				</div>
			</div>
			<div id="projectsSec" class="w-5/5 h-2/5 overflow-x-scroll">
				<h1 class="text-dark-secondary p-2">Projects</h1>
				<div id="projects" class="flex w-max overflow-y-visible p-2">
					{#each projects as project}
						<div
							class="bg-dark-surfaceContainerHigh text-dark-onSurface m-0.5 mx-1.5 h-64 w-48 rounded-lg transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
						>
							<img
								src="/src/assets/img/byteport copy.png"
								alt="BytePort"
								class="h-5/5 bg-dark-surfaceContainerHighest p-2"
							/>
							<div class="h-5/5 bg-dark-surfaceContainerHigh flex-col px-2 pb-3 pt-3 text-sm">
								<h1>{project.Name}</h1>
								<h2>{project.Status}</h2>
								<h5>{project.LastUpdated}</h5>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
		<div id="footer"></div>
	</div>
</div>

<style>
</style>
