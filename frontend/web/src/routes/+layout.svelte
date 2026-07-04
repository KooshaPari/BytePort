<script lang="ts">
	import '../app.css';
	import { user } from '../stores/user';
	import SkipLink from '$lib/components/SkipLink.svelte';
	import LiveAnnouncer from '$lib/components/LiveAnnouncer.svelte';
	let { children } = $props();
</script>

<SkipLink href="#main-content" label="Skip to main content" />
<LiveAnnouncer />
<main id="main-content" data-testid="main" tabindex="-1">
	{@render children()}
</main>

<style>
	main {
		outline: none; /* tabindex=-1 lets the skip link move focus, but we don't want a focus ring on the wrapper */
	}
</style>
