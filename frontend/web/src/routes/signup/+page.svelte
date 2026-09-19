<script lang="ts">
	/**
	 * Create an account.
	 *
	 * The previous version of this route had NO markup at all: just a script
	 * block. The form it referenced (`document.forms.namedItem('regUser')`) did
	 * not exist, so every submit failed with "Signup form was not found". It also
	 * imported the route layout as a component named `tempUser` and never used it.
	 *
	 * A confirm-password field is included because there is no password-reset
	 * route in the backend, so a typo here would lock the account out for good.
	 */
	import { goto } from '$app/navigation';
	import { ApiError, apiFetch, getApiBaseUrl } from '$lib/api';
	import { initializeUser } from '../../stores/user';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	interface SignupResponse {
		uuid: string;
		name: string;
		email: string;
	}

	const MIN_PASSWORD_LENGTH = 8;

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let submitting = $state(false);
	let formError = $state('');
	let emailTaken = $state(false);
	let fieldErrors = $state<{
		name?: string;
		email?: string;
		password?: string;
		confirmPassword?: string;
	}>({});

	function validate(): boolean {
		const next: typeof fieldErrors = {};
		const trimmedName = name.trim();
		const trimmedEmail = email.trim();

		if (!trimmedName) next.name = 'Enter your name.';
		else if (trimmedName.length > 100) next.name = 'Keep your name under 100 characters.';

		if (!trimmedEmail) next.email = 'Enter your email address.';
		else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmedEmail))
			next.email = 'Enter a valid email address.';

		if (!password) next.password = 'Choose a password.';
		else if (password.length < MIN_PASSWORD_LENGTH)
			next.password = `Use at least ${MIN_PASSWORD_LENGTH} characters.`;

		if (!confirmPassword) next.confirmPassword = 'Repeat your password.';
		else if (confirmPassword !== password) next.confirmPassword = 'The passwords do not match.';

		fieldErrors = next;
		return Object.keys(next).length === 0;
	}

	function describeFailure(err: unknown): string {
		if (err instanceof ApiError) {
			if (err.status === 409) {
				emailTaken = true;
				return 'An account with this email already exists. Sign in instead.';
			}
			if (err.status === 400) {
				// Gin's binding text ("Key: 'SignupRequest.Name' Error:Field
				// validation ...") is not user-facing copy.
				return 'The sign-up request was rejected. Check the name, email and password fields.';
			}
			if (err.status >= 500) {
				const body = err.body as { error?: string; details?: string } | null;
				const text = body?.error ?? '';
				return text
					? `${text}${body?.details ? `: ${body.details}` : ''}`
					: 'The BytePort backend hit an error. Try again in a moment.';
			}

			const body = err.body as { message?: string; error?: string; details?: string } | null;
			const text = body?.message ?? body?.error ?? '';
			return text
				? `${text}${body?.details ? `: ${body.details}` : ''}`
				: `Sign-up failed (HTTP ${err.status}).`;
		}

		if (err instanceof DOMException && err.name === 'AbortError') {
			return 'The backend did not respond in time.';
		}

		return `Could not reach the BytePort backend at ${getApiBaseUrl()}. Check that it is running.`;
	}

	async function onSubmit(event: SubmitEvent) {
		event.preventDefault();
		formError = '';
		emailTaken = false;

		if (submitting) return;
		if (!validate()) return;

		submitting = true;
		try {
			// Field names match the Go `SignupRequest` json tags exactly.
			await apiFetch<SignupResponse>('/signup', {
				method: 'POST',
				body: JSON.stringify({
					name: name.trim(),
					email: email.trim(),
					password
				})
			});

			// Signup sets the session cookie, so the first-time-setup wizard can
			// read the new session before it runs.
			await initializeUser();

			submitting = false;
			await goto('/fts');
		} catch (err) {
			formError = describeFailure(err);
			submitting = false;
		}
	}
</script>

<main class="bg-dark-background h-screen w-full overflow-y-auto">
	<div class="flex min-h-full items-center justify-center px-4 py-10">
		<div class="w-full max-w-[400px]">
			<div class="mb-6 flex flex-col items-center gap-3">
				<div
					class="border-border bg-dark-surfaceContainer flex h-10 w-10 items-center justify-center rounded-[10px] border"
					aria-hidden="true"
				>
					<span class="text-dark-primary text-[13px] font-semibold tracking-tight"
						>BP</span
					>
				</div>
				<div class="text-center">
					<h1 class="text-dark-onSurface text-[15px] font-semibold tracking-tight">
						Create your account
					</h1>
					<p class="text-dark-onSurfaceVariant mt-1 text-[13px] leading-snug">
						Set up BytePort, then connect your cloud credentials.
					</p>
				</div>
			</div>

			<Card padding="lg">
				<form class="flex flex-col gap-4" onsubmit={onSubmit} novalidate>
					<p
						class="text-dark-onSurfaceVariant text-[11px] font-medium tracking-wide uppercase"
					>
						Account
					</p>

					<Input
						label="Name"
						name="name"
						type="text"
						autocomplete="name"
						placeholder="Ada Lovelace"
						value={name}
						oninput={(e) => (name = e.currentTarget.value)}
						error={fieldErrors.name ?? ''}
						disabled={submitting}
						required
					/>

					<Input
						label="Email"
						name="email"
						type="email"
						autocomplete="username"
						autocapitalize="none"
						spellcheck={false}
						placeholder="you@example.com"
						value={email}
						oninput={(e) => (email = e.currentTarget.value)}
						error={fieldErrors.email ?? ''}
						disabled={submitting}
						required
					/>

					<Input
						label="Password"
						name="password"
						type="password"
						autocomplete="new-password"
						placeholder="At least 8 characters"
						hint={`At least ${MIN_PASSWORD_LENGTH} characters.`}
						value={password}
						oninput={(e) => (password = e.currentTarget.value)}
						error={fieldErrors.password ?? ''}
						disabled={submitting}
						required
					/>

					<Input
						label="Confirm password"
						name="confirmPassword"
						type="password"
						autocomplete="new-password"
						placeholder="Repeat your password"
						value={confirmPassword}
						oninput={(e) => (confirmPassword = e.currentTarget.value)}
						error={fieldErrors.confirmPassword ?? ''}
						disabled={submitting}
						required
					/>

					{#if formError}
						<p
							role="alert"
							class="border-dark-error/40 bg-dark-errorContainer/40 text-dark-onErrorContainer rounded-md border px-3 py-2 text-[12px] leading-snug"
						>
							{formError}
						</p>
					{/if}

					<Button type="submit" variant="primary" loading={submitting} class="w-full">
						{submitting ? 'Creating account' : 'Create account'}
					</Button>
				</form>
			</Card>

			<p class="text-dark-onSurfaceVariant mt-4 text-center text-[13px]">
				{#if emailTaken}
					<a
						href="/login"
						class="text-dark-primary underline-offset-4 hover:underline focus-visible:underline"
					>
						Sign in to your existing account
					</a>
				{:else}
					Already have an account?
					<a
						href="/login"
						class="text-dark-primary underline-offset-4 hover:underline focus-visible:underline"
					>
						Sign in
					</a>
				{/if}
			</p>
		</div>
	</div>
</main>
