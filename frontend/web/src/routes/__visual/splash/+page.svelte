<script lang="ts">
	/**
	 * /__visual/splash — visual-regression fixture that renders only the splash
	 * on a black radial-gradient backdrop, so snapshots stay focused on the
	 * splash layer (not app chrome).
	 *
	 * Lives only in dev/test — guarded by an env check on the +layout.ts of /__visual.
	 */
	import Splash from '$lib/components/Splash.svelte';
</script>

<svelte:head>
	<title>Splash — visual regression fixture</title>
</svelte:head>

<Splash progress={0.5} dismissable={false} />
