<script lang="ts">
	import { pb } from '$lib/pocketbase';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let errors: { [fieldName: string]: string } = $state({});
	let loading = $state(false);
	let emailInput: HTMLInputElement | undefined = $state();

	const handleSubmit = async (e: SubmitEvent) => {
		e.preventDefault();
		errors = {};

		const formData = new FormData(e.target as HTMLFormElement);

		const email = formData.get('email')?.toString() ?? '';
		if (email.length < 6) {
			errors['email'] = 'Email is required';
		} else if (email.length > 500) {
			errors['email'] = 'Email too long';
		} else if (!email.includes('@') || !email.includes('.')) {
			errors['email'] = 'Invalid email';
		}

		if (Object.keys(errors).length > 0) {
			return;
		}

		try {
			loading = true;
			await pb.collection('users').requestPasswordReset(email);
			loading = false;
		} catch (err: any) {
			loading = false;
		}

		goto(`/login/sign_in?forgot_password=true`);
	};

	onMount(() => {
		if (emailInput) emailInput.focus();
	});
</script>

<svelte:head>
	<title>Forgot Password</title>
</svelte:head>
<h1 class="text-2xl font-bold mb-6">Forgot Password</h1>
<form class="form-widget flex flex-col" onsubmit={handleSubmit}>
	<label for={'email'}>
		<div class="flex flex-row">
			<div class="text-base font-bold">{'Email address'}</div>
			{#if errors['email']}
				<div class="text-red-600 flex-grow text-sm ml-2 text-right">
					{errors['email']}
				</div>
			{/if}
		</div>
		<input
			bind:this={emailInput}
			id={'email'}
			name={'email'}
			type={'email'}
			autocomplete={'email'}
			placeholder={'Your email address'}
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
		>Send reset password instructions</button
	>
</form>
<div class="text-l text-slate-800 mt-4">
	Remember your password? <a class="underline" href="/login/sign_in">Sign in</a>.
</div>
