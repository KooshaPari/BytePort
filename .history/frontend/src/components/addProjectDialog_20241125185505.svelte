<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import addNewProject from '../routes/home/+page.svelte';
	import Icon from '@iconify/svelte';
	import { Project, VMInstance } from '../routes/+layout.svelte';
	let newProject: Project = {
		UUID: '',
		Name: 'New Project',
		Description: '',
		LastUpdated: 'Today',
		Status: '',
		Type: '',
		Instances: []
	};
</script>
	function addNewInstance() {
		// Open Add New Instance Modal
		// Select Project, OS, and Name, Double Check NVMS Configuration
		// Add Instance to List

		return;
	}
	function addNewProject() {
		// open dialog

		return;
	}
<Dialog.Root>
	<Dialog.Trigger
		><Icon
			on:click={()=> addNewProject()}
			class="active:text-dark-surfaceVariant hover:bg-dark-onPrimaryContainer active:bg-dark-onPrimary bg-dark-primary text-dark-onPrimary h-max w-max rounded-full p-5 text-4xl transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
			icon="ic:baseline-add"
		/></Dialog.Trigger
	>
  <Dialog.Header>
			<Dialog.Title>{newProject.Name}</Dialog.Title>
      
			<Dialog.Description>
				
			</Dialog.Description>
		</Dialog.Header>
	<Dialog.Content>
		
	</Dialog.Content>
  <Dialog.Footer></Dialog.Footer>
</Dialog.Root>
