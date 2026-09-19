<script lang="ts">
	/**
	 * Sign in.
	 *
	 * Rewritten from the 2023 view, which had three real defects:
	 *
	 *  1. It called `platform()` from `@tauri-apps/plugin-os` unguarded (twice).
	 *     That plugin is not registered in the Rust shell, so the call threw a
	 *     TypeError while dereferencing `window.__TAURI_OS_PLUGIN_INTERNALS__`.
	 *     Base-URL detection now goes through `$lib/api`, which is total.
	 *  2. It imported the route LAYOUT as a component under the name `tempUser`
	 *     (`import tempUser from '../+layout.svelte'`), which rendered the layout
	 *     a second time and was never used. Removed.
	 *  3. The form had no field validation, no disabled/loading state and no
	 *     visible error surface: failures were only `console.log`ged.
	 */
	import { goto } from '$app/navigation';
	import { ApiError, apiFetch, getApiBaseUrl } from '$lib/api';
	import { initializeUser } from '../../stores/user';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	interface LoginResponse {
		message: string;
		user: { uuid: string; name: string; email: string };
	}

	let email = $state('');
	let password = $state('');
	let submitting = $state(false);
	let formError = $state('');
	let fieldErrors = $state<{ email?: string; password?: string }>({});

	/** Client-side checks only. The backend is the authority on credentials. */
	function validate(): boolean {
		const next: { email?: string; password?: string } = {};
		const trimmed = email.trim();

		if (!trimmed) next.email = 'Enter your email address.';
		else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed))
			next.email = 'Enter a valid email address.';

		// No complexity rule here on purpose: the password on an existing account
		// may predate any policy we would invent, and rejecting it client-side
		// would lock the user out with no way to tell why.
		if (!password) next.password = 'Enter your password.';

		fieldErrors = next;
		return Object.keys(next).length === 0;
	}

	function describeFailure(err: unknown): string {
		if (err instanceof ApiError) {
			// The backend answers 401 with either "Failed, user not found" or
			// "Failed, invalid credentials". Reporting those verbatim would let
			// anyone test whether an email is registered, so collapse them.
			if (err.status === 401) return 'Email or password is incorrect.';
			if (err.status === 400)
				return 'The sign-in request was rejected. Check the email format.';
			if (err.status >= 500)
				return 'The BytePort backend hit an error. Try again in a moment.';

			const body = err.body as { message?: string; error?: string; details?: string } | null;
			const text = body?.message ?? body?.error ?? '';
			return text
				? `${text}${body?.details ? `: ${body.details}` : ''}`
				: `Sign-in failed (HTTP ${err.status}).`;
		}

		if (err instanceof DOMException && err.name === 'AbortError') {
			return 'The backend did not respond in time.';
		}

		return `Could not reach the BytePort backend at ${getApiBaseUrl()}. Check that it is running.`;
	}

	async function onSubmit(event: SubmitEvent) {
		event.preventDefault();
		formError = '';

		if (submitting) return;
		if (!validate()) return;

		submitting = true;
		try {
			// Field names match the Go `LoginRequest` json tags exactly.
			await apiFetch<LoginResponse>('/login', {
				method: 'POST',
				body: JSON.stringify({ email: email.trim(), password })
			});

			// Confirms the session cookie was actually stored before we navigate.
			await initializeUser();

			submitting = false;
			await goto('/home');
		} catch (err) {
			formError = describeFailure(err);
			submitting = false;
		}
	}
</script>

<main class="bg-dark-background h-screen w-full overflow-y-auto">
	<div class="flex min-h-full items-center justify-center px-4 py-10">
		<div class="w-full max-w-[400px]">
			<!-- Product mark. A typographic monogram avoids depending on the
			     brand PNG, which lives in src/assets (Vite-processed) and is not
			     published from static/, so a raw <img> path would 404 in the
			     packaged desktop build. -->
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
						Sign in to BytePort
					</h1>
					<p class="text-dark-onSurfaceVariant mt-1 text-[13px] leading-snug">
						Deploy and manage your portfolio infrastructure.
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
						autocomplete="current-password"
						placeholder="Your password"
						value={password}
						oninput={(e) => (password = e.currentTarget.value)}
						error={fieldErrors.password ?? ''}
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
						{submitting ? 'Signing in' : 'Sign in'}
					</Button>
				</form>
			</Card>

			<p class="text-dark-onSurfaceVariant mt-4 text-center text-[13px]">
				No account yet?
				<a
					href="/signup"
					class="text-dark-primary underline-offset-4 hover:underline focus-visible:underline"
				>
					Create one
				</a>
			</p>
		</div>
	</div>
</main>
