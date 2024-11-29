<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Button from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';
	import ProjectForm from './projectForm.svelte';
	import type { Project, Instance } from '$lib/utils';
	import GitSearch from './gitSearch.svelte';
	let projheadTxt: string = 'Welcome.';
	let stage = 1;
	let projHeadDescr: string = "Let's Begin First Time Setup";
	let newProject: Project = {
		Owner: '',
		User: null,
		Repository: null,
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
		stage = 1;
		setStage();
		return;
	}
	async function setStage() {
		const projForm = document.querySelector(`#form-stage-${stage}`) as HTMLFormElement;
		if (!projForm.checkValidity()) {
			projForm.reportValidity(); // Show validation errors
			return; // Stop progression
		}
		const formData = new FormData(projForm);
		const data = Object.fromEntries(formData.entries());
		switch (stage) {
 
				case 2:
					newProject.Name = data.name as string;
					newProject.Description = data.description as string;
					
					break;
 
				default:
					console.log('Odd Stage');
					break;
			}

			console.log(`Stage ${stage} data collected:`, userData);

		stage++;
		switch (stage) {
			case 1:
				projheadTxt = "Let's Start With Some Basic Information...";
				projHeadDescr = 'Enter Your AWS Credentials Below';
				break;
			case 2:
				projheadTxt = "Let's Continue With OpenAI Credentials...";
				projHeadDescr = 'Enter Your OpenAI Credentials Below';
				break;
			case 3:
				projheadTxt = "Let's Connect Your Portfolio...";
				projHeadDescr = 'Provide Your Portfolios';
				break;
			case 4:
				projheadTxt = "Let's Connect Your Git Provider";
				projHeadDescr = 'Please Continue On GitHub';

				break;
			case 5:
				projheadTxt = 'Setup Complete!';
				projHeadDescr = 'You have completed the first-time setup.';
				break;
			default:
				projheadTxt = 'Setup ERR!';
				break;
		}
	}
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

			<Dialog.Description><h1>{projheadTxt}</h1></Dialog.Description>
		</Dialog.Header>
		{#if stage == 1}
			<GitSearch
				select={(e) => {
					const selectedRepo = e;
					console.log('TEST');
					newProject.Repository = selectedRepo;
					setStage();
				}}
			/>
		{:else if stage == 2}<ProjectForm />{:else if stage == 3}{:else if stage == 4}{/if}
		<Dialog.Footer>
			{#if stage == 3}
				<Button.Root
					variant="ghost"
					on:click={() => {
						stage = stage - 2;
						setStage();
					}}>Back</Button.Root
				>
				<Button.Root
					variant="outline"
					on:click={() => {
						setStage();
					}}>Next</Button.Root
				>
			{:else if stage == 2}
				<Button.Root
					variant="ghost"
					on:click={() => {
						stage = stage - 2;
						setStage();
					}}>Back</Button.Root
				>
				<Button.Root
					variant="outline"
					on:click={() => {
						setStage();
					}}>Review</Button.Root
				>
			{/if}
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
