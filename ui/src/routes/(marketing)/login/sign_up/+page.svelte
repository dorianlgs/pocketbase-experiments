<script lang="ts">
	import { pb } from '$lib/pocketbase';
	import GitHubButton from '$lib/components/GitHubButton.svelte';
	import GoogleButton from '$lib/components/GoogleButton.svelte';
	import InputFile from '$lib/components/InputFile.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let errors: { [fieldName: string]: string } = $state({});
	let loading = $state(false);
	let emailInput: HTMLInputElement | undefined = $state();

	const handleSubmit = async (e: SubmitEvent) => {
		try {
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

			const name = formData.get('name')?.toString() ?? '';
			if (name.length < 2) {
				errors['name'] = 'Name is required';
			}
			if (name.length > 500) {
				errors['name'] = 'Name too long';
			}

			const password = formData.get('password')?.toString() ?? '';
			if (password.length > 500) {
				errors['password'] = 'Password too long';
			}

			const avatar = formData.get('avatar') as File;

			if (avatar?.size === 0) {
				errors['avatar'] = 'Avatar is required';
			}

			if (Object.keys(errors).length > 0) {
				return;
			}

			loading = true;
			await pb.collection('users').create({
				email: email,
				name: name,
				password: password,
				passwordConfirm: password,
				avatar: avatar
			});

			await pb.collection('users').requestVerification(email);
			loading = false;
			goto(`/login/sign_in?not_verified=true`);
		} catch (err) {
			loading = false;
		}
	};

	onMount(() => {
		if (emailInput) emailInput.focus();
	});
</script>

<svelte:head>
	<title>Sign up</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-6">Sign Up</h1>
<GoogleButton />
<hr class="solid" />
<br />
<GitHubButton />
<br />
<hr class="solid" />
<br />
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
	<label for={'name'}>
		<div class="flex flex-row">
			<div class="text-base font-bold">{'Name'}</div>
			{#if errors['name']}
				<div class="text-red-600 flex-grow text-sm ml-2 text-right">
					{errors['name']}
				</div>
			{/if}
		</div>
		<input
			id={'name'}
			name={'name'}
			autocomplete={'off'}
			placeholder={'Your name'}
			class="{errors['name']
				? 'input-error'
				: ''} input-md mt-1 input input-bordered w-full mb-3 text-base py-4"
		/>
	</label>
	<label for={'password'}>
		<div class="flex flex-row">
			<div class="text-base font-bold">{'Create a Password'}</div>
			{#if errors['password']}
				<div class="text-red-600 flex-grow text-sm ml-2 text-right">
					{errors['password']}
				</div>
			{/if}
		</div>
		<input
			id={'password'}
			name={'password'}
			type={'password'}
			autocomplete={'off'}
			placeholder={'Your password'}
			class="{errors['email']
				? 'input-error'
				: ''} input-md mt-1 input input-bordered w-full mb-3 text-base py-4"
		/>
	</label>
	<label for={'avatar'}>
		<div class="flex flex-row">
			<div class="text-base font-bold">{'Avatar'}</div>
			{#if errors['avatar']}
				<div class="text-red-600 flex-grow text-sm ml-2 text-right">
					{errors['avatar']}
				</div>
			{/if}
		</div>
		<InputFile {errors} placeholder={'Upload an avatar'} name={'avatar'} />

		<br />
	</label>

	{#if Object.keys(errors).length > 0}
		{#if errors['loginResult']}
			<p class="text-red-600 text-sm mb-2">{errors['loginResult']}</p>
		{:else}
			<p class="text-red-600 text-sm mb-2">Please resolve above issues.</p>
		{/if}
	{/if}

	<button disabled={loading} type={'submit'} class="btn btn-primary {loading ? 'btn-disabled' : ''}"
		>Sign up</button
	>
</form>
<div class="text-l text-slate-800 mt-4 mb-2">
	Have an account? <a class="underline" href="/login/sign_in">Sign in</a>
</div>
