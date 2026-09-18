<script lang="ts">
	import { goto } from '$app/navigation';
	import { initializeUser, setUser, user } from '../../stores/user';
	import type { User } from '../../stores/user';

	let newUser: User;
	let Error: string = '';

	// Get base URL with correct port
	const getBaseUrl = () => {
		const url = new URL(window.location.href);
		url.hostname = url.hostname.split('.').slice(-2).join('.');
		url.port = '8081';
		return url.origin;
	};

	async function signUpUser() {
		const regUserForm = document.forms.namedItem('regUser');
		if (!regUserForm) {
			Error = 'Signup form was not found.';
			return;
		}
		const formData = new FormData(regUserForm);
		let newUser = {
			Name: String(formData.get('name') ?? ''),
			Email: String(formData.get('email') ?? ''),
			Password: String(formData.get('password') ?? '')
		};

		const { Name, Email, Password } = newUser;
		try {
			const baseUrl = getBaseUrl();
			const response = await fetch(`${baseUrl}/signup`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ Name, Email, Password }),
				credentials: 'include'
			});

			console.log('Response Status:', response.status);
			console.log('Response OK:', response.ok);

			const data = await response.json();

			if (response.ok) {
				console.log('Signup successful:', data);
				await initializeUser(baseUrl);
				goto('/fts');
			} else {
				Error = data.message || data.error || 'An unknown error occurred';
				console.log('Signup failed:', Error);
			}
		} catch (err) {
			console.error('Error during signup:', err);
			Error = 'An error occurred during signup.';
		}
	}
</script>

<div class="bg-dark-surface h-screen w-screen overflow-x-hidden">
	<div
		id="header"
		class="bg-dark-surfaceContainerLow h-1/5 w-5/5 flex-col justify-between ps-2.5"
	>
		<div id="headerNav" class="h-3/5 pt-2.5"></div>
		<div id="headerContent" class="h-2/5 text-4xl text-white">Welcome.</div>
	</div>
	<div id="body" class="px-2.5 pt-5">
		<h1 class="text-2xl text-white">Please Register Below...</h1>
		<div id="signUpCont">
			<form class="flex-row" name="regUser" on:submit|preventDefault={signUpUser}>
				<div>
					<label for="name">Name</label>
					<input name="name" placeholder="Name" required type="text" />
				</div>
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
				{#if Error}
					<p class="text-dark-error">{Error}</p>
				{/if}
				<div>
					<input
						type="submit"
						value="Sign Up"
						class="bg-dark-surfaceContainerHigh text-dark-onSurface hover:bg-dark-surfaceContainerHighest active:bg-dark-surfaceContainer rounded-full p-2"
					/>
					<button
						on:click={() => goto('/login')}
						class="bg-dark-surfaceContainerHigh text-dark-onSurface hover:bg-dark-surfaceContainerHighest active:bg-dark-surfaceContainer my-3 rounded-full p-2"
					>
						Log in
					</button>
				</div>
			</form>
		</div>
	</div>
</div>

<style>
	@reference '../../app.css';

	#signUpCont form > div > input {
		@apply bg-dark-surfaceContainerHigh text-dark-onSurface placeholder-dark-onSurfaceVariant selection:bg-dark-surfaceContainer hover:bg-dark-surfaceContainerHighest my-2 rounded-full;
		border: none;
	}
	#signUpCont form > div > label {
		@apply text-dark-onSurface;
	}
	#signUpCont form > div {
		@apply h-1/5 w-screen flex-row items-center justify-center;
	}
</style>
