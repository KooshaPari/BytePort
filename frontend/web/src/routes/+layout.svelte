<script lang="ts">
	/**
	 * Root layout.
	 *
	 * Deliberately chrome-free: /login, /signup and the full-text search page are
	 * standalone surfaces, so the application shell lives in `home/+layout.svelte`
	 * where it wraps the authenticated routes only.
	 */
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>BytePort</title>
</svelte:head>

{@render children()}
