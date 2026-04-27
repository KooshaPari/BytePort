<script>
  	import * as Dialog from '$lib/components/ui/dialog';
	import addNewProject from '../routes/home/+page.svelte';
	import Icon from '@iconify/svelte';
	export let title = 'Default Title'; // Dialog title
	export let description = 'Default description'; // Dialog description
	export let icon = 'ic:baseline-add'; // Iconify icon (default add icon)
	export let onConfirm; // Callback for confirm action
	export let onCancel; // Callback for cancel action (optional)
</script>

<Dialog.Root>
	<Dialog.Trigger>
		<Icon
			on:click={onConfirm}
			class="active:text-dark-surfaceVariant hover:bg-dark-onPrimaryContainer active:bg-dark-onPrimary bg-dark-primary text-dark-onPrimary h-max w-max rounded-full p-5 text-4xl transition-all hover:-translate-y-2 hover:scale-105 active:translate-y-1 active:scale-100"
			{icon}
		/>
	</Dialog.Trigger>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{title}</Dialog.Title>
			<Dialog.Description>
				{description}
			</Dialog.Description>
		</Dialog.Header>
		<div class="actions mt-4 flex justify-end gap-2">
			<button
				on:click={onCancel}
				class="bg-dark-surfaceContainer text-dark-onSurface rounded px-4 py-2"
			>
				Cancel
			</button>
			<button on:click={onConfirm} class="bg-dark-primary text-dark-onPrimary rounded px-4 py-2">
				Confirm
			</button>
		</div>
	</Dialog.Content>
</Dialog.Root>

<style>
	/* Example styles */
	.actions button {
		transition: all 0.2s ease-in-out;
	}
	.actions button:hover {
		transform: scale(1.05);
	}
	.actions button:active {
		transform: scale(0.95);
	}
</style>
