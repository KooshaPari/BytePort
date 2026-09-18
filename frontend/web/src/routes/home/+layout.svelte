<script lang="ts">
	/**
	 * Authenticated shell for every /home route.
	 *
	 * The sidebar, page header and backend indicator previously lived inside each
	 * page (copied seven times, drifting each time), which is why navigation looked
	 * different on every screen. They now exist once, here.
	 */
	import AppShell from '$lib/components/shell/AppShell.svelte';

	let { children } = $props();
</script>

<AppShell>
	{@render children()}
</AppShell>
