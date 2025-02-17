<script lang="ts">
	import { pb } from '$lib/pocketbase';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	let errors: { [fieldName: string]: string } = $state({});
	let loading = $state(false);
	let emailInput: HTMLInputElement | undefined = $state();

	const handleSubmit = async (e: SubmitEvent) => {
		e.preventDefault();
		errors = {};

		const formData = new FormData(e.target as HTMLFormElement);

		const newPassword = formData.get('newPassword')?.toString() ?? '';
		if (newPassword.length > 500) {
			errors['newPassword'] = 'New Password too long';
		}

		const newPasswordConfirm = formData.get('newPasswordConfirm')?.toString() ?? '';
		if (newPasswordConfirm.length > 500) {
			errors['newPasswordConfirm'] = 'New Password Confirm too long';
		}

		if (newPassword !== newPasswordConfirm) {
			errors['newPasswordConfirm'] = 'Passwords are different';
		}

		const resetToken = page.url.searchParams.get('reset_token') as string;

		if (!resetToken) {
			errors['newPasswordConfirm'] = 'Empty token';
		}

		if (Object.keys(errors).length > 0) {
			return;
		}

		try {
			loading = true;
			await pb
				.collection('users')
				.confirmPasswordReset(resetToken, newPassword, newPasswordConfirm);
			loading = false;
			goto(`/login/sign_in?password_changed=true`);
		} catch (err: any) {
			loading = false;
		}
	};

	onMount(() => {
		if (emailInput) emailInput.focus();
	});
</script>

<svelte:head>
	<title>Password Reset</title>
</svelte:head>
<h1 class="text-2xl font-bold mb-6">Password Reset</h1>
<form class="form-widget flex flex-col" onsubmit={handleSubmit}>
	<label for={'newPassword'}>
		<div class="flex flex-row">
			<div class="text-base font-bold">{'New Password'}</div>
			{#if errors['newPassword']}
				<div class="text-red-600 flex-grow text-sm ml-2 text-right">
					{errors['newPassword']}
				</div>
			{/if}
		</div>
		<input
			id={'newPassword'}
			name={'newPassword'}
			type={'password'}
			autocomplete={'off'}
			placeholder={'Your new password'}
			class="{errors['email']
				? 'input-error'
				: ''} input-md mt-1 input input-bordered w-full mb-3 text-base py-4"
		/>
	</label>
	<label for={'newPasswordConfirm'}>
		<div class="flex flex-row">
			<div class="text-base font-bold">{'New Password Confirm'}</div>
			{#if errors['newPasswordConfirm']}
				<div class="text-red-600 flex-grow text-sm ml-2 text-right">
					{errors['newPasswordConfirm']}
				</div>
			{/if}
		</div>
		<input
			id={'newPasswordConfirm'}
			name={'newPasswordConfirm'}
			type={'password'}
			autocomplete={'off'}
			placeholder={'Your new password confirm'}
			class="{errors['email']
				? 'input-error'
				: ''} input-md mt-1 input input-bordered w-full mb-3 text-base py-4"
		/>
	</label>
	{#if Object.keys(errors).length > 0}
		{#if errors['loginResult']}
			<p class="text-red-600 text-sm mb-2">{errors['loginResult']}</p>
		{:else}
			<p class="text-red-600 text-sm mb-2">Please resolve above issues.</p>
		{/if}
	{/if}

	<button disabled={loading} class="btn btn-primary {loading ? 'btn-disabled' : ''}"
		>Set new password</button
	>
</form>
