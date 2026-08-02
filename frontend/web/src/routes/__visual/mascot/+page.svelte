<script lang="ts">
	/** /__visual/mascot — render only the mascot asset on a transparent background. */
</script>

<svelte:head>
	<title>Mascot — visual regression fixture</title>
</svelte:head>

<main class="fixture">
	<img class="mascot" src="/brand/mascot.svg" alt="Byte, the BytePort mascot" />
</main>

<style>
	.fixture {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #0a0f24 0%, #1c2756 100%);
	}
	.mascot {
		width: min(60vw, 360px);
		height: auto;
	}
</style>
