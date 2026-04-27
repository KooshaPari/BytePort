<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import Icon from '@iconify/svelte';
	import type { Project, VMInstance } from '../routes/+layout.svelte';
	import GitSearch from './gitSearch.svelte';
	let stage;
	let newProject: Project = {
		UUID: '',
		Name: 'New Project',
		Description: '',
		LastUpdated: 'Today',
		Status: '',
		Type: '',
		Instances: []
	};

	function addNewProject() {
		// open dialog

		return;
	}
	function setStage() {}
</script>

<Dialog.Root>
	<Dialog.Trigger
		><Icon
			on:click={() => addNewProject()}
			class="active:text-dark-surfaceVariant hover:bg-dark-onPrimaryContainer active:bg-dark-onPrimary bg-dark-primary text-dark-onPrimary h-max w-max rounded-full p-5 text-4xl transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
			icon="ic:baseline-add"
		/></Dialog.Trigger
	>

	<Dialog.Content
		><Dialog.Header>
			<Dialog.Title>{newProject.Name}</Dialog.Title>

			<Dialog.Description><h1>Select Repository</h1></Dialog.Description>
		</Dialog.Header>
		{#if stage == 1}
		<GitSearch />
		{:else if stage == 2}
		
		{:else if stage == 3}
		{:else if stage == 4}
		{/if}
		<Dialog.Footer>

		</Dialog.Footer>
		</Dialog.Content>
</Dialog.Root>
