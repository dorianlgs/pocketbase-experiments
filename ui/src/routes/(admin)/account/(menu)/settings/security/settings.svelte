<script lang="ts">
	import { pb } from '$lib/pocketbase';
	import { currentUser } from '$lib/stores/user';

	let title: string = $state('Security Settings');
	let loading: boolean = $state(false);
	let showSuccess: boolean = $state(false);
	let message: string = $state('mensaje');
	let dangerous: boolean = $state(false);
	let currentValue: boolean = $state($currentUser?.multiFactorAuth);

	const handleSubmit = async (e: SubmitEvent) => {
		e.preventDefault();

		const formData = new FormData(e.target as HTMLFormElement);

		const newValueTwoFactorAuth = formData.get('twoFactorAuth')?.toString() ?? '';

		if ($currentUser) {
			loading = true;
			await pb.collection('users').update($currentUser?.id, {
				multiFactorAuth: newValueTwoFactorAuth === 'on' ? true : false
			});
			loading = false;
			showSuccess = true;
		}
	};
</script>

<div class="card p-6 pb-7 mt-8 max-w-xl flex flex-col md:flex-row shadow">
	{#if title}
		<div class="text-xl font-bold mb-3 w-48 md:pr-8 flex-none">{title}</div>
	{/if}

	<div class="w-full min-w-48">
		{#if !showSuccess}
			{#if message}
				<div class="mb-6 {dangerous ? 'alert alert-warning' : ''}">
					{#if dangerous}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="stroke-current shrink-0 h-6 w-6"
							fill="none"
							viewBox="0 0 24 24"
							><path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
							/></svg
						>
						<span>{message}</span>
					{/if}
				</div>
			{/if}
			<form class="form-widget flex flex-col" onsubmit={handleSubmit}>
				<label for={'twoFactorAuth'}>
					<span class="text-sm text-gray-500">{'2-Step Authentication'}</span>
				</label>
				<input
					class="checkbox"
					id={'twoFactorAuth'}
					name={'twoFactorAuth'}
					type={'checkbox'}
					placeholder={'2-Step Authentication'}
					defaultChecked={currentValue}
				/>
				<button
					type="submit"
					class="ml-auto btn btn-sm mt-3 min-w-[145px] {dangerous
						? 'btn-error'
						: 'btn-primary btn-outline'}"
					disabled={loading}
				>
					{#if loading}
						<span class="loading loading-spinner loading-md align-middle mx-3"></span>
					{:else}
						{'Save'}
					{/if}
				</button>
			</form>
		{:else}
			<div>
				<div class="text-base">{'Updated successfully'}</div>
			</div>
			<a href="/account/settings">
				<button class="btn btn-outline btn-sm mt-3 min-w-[145px]"> Return to Settings </button>
			</a>
		{/if}
	</div>
</div>
