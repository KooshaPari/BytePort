<script lang="ts">
	import { goto } from '$app/navigation';
	import tempUser from '../+layout.svelte';
	import { initializeUser, setUser, user } from '../../stores/user';
	import type { User } from '../../stores/user';
	let variant = 'primary'; // Options: primary, secondary, destructive
	let disabled = false;
	let newUser: User;
	let Error: string = '';
	const SERVER_URL = 'http://localhost:8080';
	async function login() {
		let newUser = {
			Email: document.forms['regUser']['email'].value,
			Password: document.forms['regUser']['password'].value
		};

		const { Email, Password } = newUser;
		try {
			const response = await fetch(`${SERVER_URL}/login`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ Email, Password }),
				credentials: 'include'
			});

			console.log('Response Status:', response.status);
			console.log('Response OK:', response.ok);

			const data = await response.json();

			if (response.ok) {
				console.log('Login successful:', data);
				await initializeUser();

				//setUser(true, data as User);
				goto('/home');
			} else {
				Error = data.message || data.error || 'An unknown error occurred';
				console.log('Login failed:', Error);
			}
		} catch (err) {
			console.error('Error during Login:', err);
			Error = 'An error occurred during login.';
		}
	}
</script>

<div class="bg-dark-surface h-screen w-screen overflow-x-hidden">
	<div id="header" class=" w-5/5 bg-dark-surfaceContainerLow h-1/5 flex-col justify-between ps-2.5">
		<div id="headerNav" class="h-3/5 pt-2.5"></div>
		<div id="headerContent" class="h-2/5 text-4xl text-white">Hello.</div>
	</div>
	<div id="body" class="px-2.5 pt-5">
		<h1 class="text-2xl text-white">Please Register Below...</h1>
		<div class="space-y-2">
			<label for="name" class="text-foreground text-sm font-medium">Name</label>
			<input
				type="text"
				id="name"
				placeholder="Enter your name"
				class="bg-input text-foreground border-border focus:ring-primary w-full rounded-lg border px-3 py-2 focus:outline-none focus:ring"
			/>
		</div>
		<div class="bg-card text-card-foreground rounded-lg p-6 shadow-md">
			<h2 class="text-lg font-semibold">Card Title</h2>
			<p class="mt-2 text-sm">This is a sample card component to test your theme configuration.</p>
			<div class="mt-4">
				<button class="bg-primary text-primary-foreground hover:bg-primary/90 rounded-md px-4 py-2">
					Action
				</button>
			</div>
		</div>
		<button
			class="rounded-lg px-4 py-2 font-medium transition-colors duration-150
		{variant === 'primary' && 'bg-primary text-primary-foreground hover:bg-primary/90'}
		{variant === 'secondary' && 'bg-secondary text-secondary-foreground hover:bg-secondary/90'}
		{variant === 'destructive' && 'bg-destructive text-destructive-foreground hover:bg-destructive/90'}
		{disabled && 'cursor-not-allowed opacity-50'}"
			{disabled}
		>
			Click Me
		</button>
		<div id="logCont">
			<form class="flex-row" name="regUser" on:submit|preventDefault={login}>
				<div>
					<label for="email">Email</label>
					<input name="email" placeholder="Email" required type="email" />
				</div>
				<div>
					<label for="password">Password</label>
					<input
						name="password"
						pattern="(?=.*\d)(?=.*[a-z])(?=.*[A-Z])+"
						type="password"
						required
						placeholder="Password"
					/>
				</div>
				<div>
					<input
						type="submit"
						value="Log In"
						class="bg-dark-surfaceContainerHigh text-dark-onSurface hover:bg-dark-surfaceContainerHighest active:bg-dark-surfaceContainer rounded-full p-2"
					/>
					<button
						on:click={() => goto('/signup')}
						class="bg-dark-surfaceContainerHigh text-dark-onSurface hover:bg-dark-surfaceContainerHighest active:bg-dark-surfaceContainer my-3 rounded-full p-2"
					>
						Sign up
					</button>
				</div>
			</form>
		</div>
	</div>
</div>

<style>
	#logCont form > div > input {
		@apply bg-dark-surfaceContainerHigh text-dark-onSurface placeholder-dark-onSurfaceVariant selection:bg-dark-surfaceContainer hover:bg-dark-surfaceContainerHighest my-2 rounded-full;
		border: none;
	}
	#logCont form > div > label {
		@apply text-dark-onSurface;
	}
	#logCont form > div {
		@apply h-1/5 w-screen flex-row items-center justify-center;
	}
</style>
