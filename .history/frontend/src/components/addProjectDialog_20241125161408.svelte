<script>
	import * as Dialog from '$lib/components/ui/dialog';
	import addNewProject from '../routes/home/+page.svelte';
	import Icon from '@iconify/svelte';
  import {Project, VMInstance} from'$lib/
  let newProject: Project;
</script>

<Dialog.Root>
	<Dialog.Trigger
		><Icon
			on:click={addNewProject}
			class="active:text-dark-surfaceVariant hover:bg-dark-onPrimaryContainer active:bg-dark-onPrimary bg-dark-primary text-dark-onPrimary h-max w-max rounded-full p-5 text-4xl transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
			icon="ic:baseline-add"
		/></Dialog.Trigger
	>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{newProject.}</Dialog.Title>
			<Dialog.Description>
				This action cannot be undone. This will permanently delete your account and remove your data
				from our servers.
			</Dialog.Description>
		</Dialog.Header>
	</Dialog.Content>
</Dialog.Root>
