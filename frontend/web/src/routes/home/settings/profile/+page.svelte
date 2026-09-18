<script lang="ts">
	/**
	 * Profile settings.
	 *
	 * Was a single unlabelled row of inputs (the password pair rendered under
	 * one field name), with a bare form button that fired a request whose result
	 * was never read. There were no section headings, no field descriptions and
	 * no indication that a save had happened.
	 *
	 * Now: two labelled sections with one-line descriptions, inline field
	 * errors, and explicit idle, dirty, saving, saved and failed feedback. The
	 * request is unchanged: PUT /user/:id/creds with { name, email, password },
	 * where an empty password leaves the existing one alone (see
	 * backend/byteport/routes/auth.go UpdateUser).
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { user, initializeUser } from '../../../../stores/user';
	import { apiFetch, getApiBaseUrl } from '$lib/api';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import { formSchema } from './schema';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');

	let initialName = $state('');
	let initialEmail = $state('');

	let saving = $state(false);
	let fieldErrors = $state<Record<string, string>>({});
	let feedback = $state<{ kind: 'saved' | 'error'; message: string } | null>(null);

	const dirty = $derived(
		name !== initialName ||
			email !== initialEmail ||
			password.length > 0 ||
			confirmPassword.length > 0
	);

	// One place decides what the header reports, so the label and its colour
	// cannot disagree.
	const statusLabel = $derived(
		saving
			? 'Saving'
			: feedback?.kind === 'saved'
				? feedback.message
				: dirty
					? 'Unsaved changes'
					: ''
	);
	const statusClass = $derived(
		`text-[12px] ${!saving && feedback?.kind === 'saved' ? 'text-dark-primary' : 'text-dark-onSurfaceVariant'}`
	);

	function clearFeedback() {
		if (feedback) feedback = null;
	}

	async function save() {
		clearFeedback();
		fieldErrors = {};

		const parsed = formSchema.safeParse({ name, email, password, confirmPassword });
		if (!parsed.success) {
			const errors: Record<string, string> = {};
			for (const issue of parsed.error.issues) {
				const path = issue.path.join('.');
				if (!errors[path]) errors[path] = issue.message;
			}
			fieldErrors = errors;
			feedback = { kind: 'error', message: 'Check the highlighted fields.' };
			return;
		}

		const uuid = $user.data?.uuid;
		if (!uuid) {
			feedback = { kind: 'error', message: 'No signed-in user, so nothing was saved.' };
			return;
		}

		saving = true;
		try {
			await apiFetch(`/user/${uuid}/creds`, {
				method: 'PUT',
				body: JSON.stringify({ name, email, password })
			});
			initialName = name;
			initialEmail = email;
			password = '';
			confirmPassword = '';
			feedback = { kind: 'saved', message: 'Profile updated.' };
		} catch (error) {
			feedback = {
				kind: 'error',
				message:
					error instanceof Error
						? `Could not save your profile: ${error.message}`
						: 'Could not save your profile.'
			};
		} finally {
			saving = false;
		}
	}

	function reset() {
		name = initialName;
		email = initialEmail;
		password = '';
		confirmPassword = '';
		fieldErrors = {};
		clearFeedback();
	}

	onMount(() => {
		let seeded = false;
		const unsubscribe = user.subscribe((value) => {
			if (value.status === 'unauthenticated') {
				goto('/login');
				return;
			}
			if (value.status === 'authenticated' && !seeded) {
				seeded = true;
				name = value.data?.name ?? '';
				email = value.data?.email ?? '';
				initialName = name;
				initialEmail = email;
			}
		});
		void initializeUser(getApiBaseUrl());
		return unsubscribe;
	});
</script>

<div class="mx-auto max-w-3xl">
	<form
		class="space-y-6"
		onsubmit={(event) => {
			event.preventDefault();
			void save();
		}}
	>
		<div class="flex items-center justify-end gap-3">
			<span class={statusClass} aria-live="polite">{statusLabel}</span>
			<Button variant="ghost" size="sm" disabled={!dirty || saving} onclick={reset}
				>Reset</Button
			>
			<Button
				variant="primary"
				size="sm"
				disabled={!dirty}
				loading={saving}
				onclick={() => void save()}
			>
				Save changes
			</Button>
		</div>

		{#if feedback?.kind === 'error'}
			<div
				class="border-dark-error/40 bg-dark-errorContainer/40 rounded-lg border px-4 py-3"
				role="alert"
			>
				<p class="text-dark-onSurface text-[13px]">{feedback.message}</p>
			</div>
		{/if}

		<section class="space-y-2">
			<div class="px-1">
				<h2
					class="text-dark-onSurfaceVariant text-[11px] font-medium tracking-wide uppercase"
				>
					Identity
				</h2>
				<p class="text-dark-onSurfaceVariant mt-0.5 text-[12px]">
					How your account is identified across BytePort.
				</p>
			</div>
			<div class="border-border bg-dark-surfaceContainer space-y-4 rounded-lg border p-4">
				<Input
					label="Display name"
					hint="Shown in the sidebar and on deployment records."
					error={fieldErrors['name'] ?? ''}
					value={name}
					autocomplete="name"
					oninput={(event) => {
						name = event.currentTarget.value;
						clearFeedback();
					}}
				/>
				<Input
					label="Email address"
					hint="Used to sign in. Changing it changes your login."
					error={fieldErrors['email'] ?? ''}
					type="email"
					value={email}
					autocomplete="email"
					oninput={(event) => {
						email = event.currentTarget.value;
						clearFeedback();
					}}
				/>
			</div>
		</section>

		<section class="space-y-2">
			<div class="px-1">
				<h2
					class="text-dark-onSurfaceVariant text-[11px] font-medium tracking-wide uppercase"
				>
					Password
				</h2>
				<p class="text-dark-onSurfaceVariant mt-0.5 text-[12px]">
					Leave both fields empty to keep your current password.
				</p>
			</div>
			<div class="border-border bg-dark-surfaceContainer space-y-4 rounded-lg border p-4">
				<Input
					label="New password"
					hint="At least 8 characters."
					error={fieldErrors['password'] ?? ''}
					type="password"
					value={password}
					autocomplete="new-password"
					oninput={(event) => {
						password = event.currentTarget.value;
						clearFeedback();
					}}
				/>
				<Input
					label="Confirm new password"
					error={fieldErrors['confirmPassword'] ?? ''}
					type="password"
					value={confirmPassword}
					autocomplete="new-password"
					oninput={(event) => {
						confirmPassword = event.currentTarget.value;
						clearFeedback();
					}}
				/>
			</div>
		</section>
	</form>
</div>
