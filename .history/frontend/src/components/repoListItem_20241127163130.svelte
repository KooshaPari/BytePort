<script>
	import * as Command from '$lib/components/ui/command';
	import * as Avatar from '$lib/components/ui/avatar';
	import Icon from '@iconify/svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Root } from '$lib/components/ui/button';
	// repo is passed as a prop
	export let repo;
</script>

{repo.name}
