<script lang="ts">
	import SettingsModule from '../settings_module.svelte';

	interface Props {
		profile: {
			unsubscribed: string;
		};
	}

	let { profile }: Props = $props();
	let unsubscribed = profile?.unsubscribed;
</script>

<svelte:head>
	<title>Change Email Subscription</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-6">Email Subscription</h1>

<SettingsModule
	editable={true}
	title="Subscription"
	message={unsubscribed
		? 'You are currently unsubscribed from emails'
		: 'You are currently subscribed to emails'}
	saveButtonTitle={unsubscribed ? 'Re-subscribe' : 'Unsubscribe'}
	successBody={unsubscribed
		? 'You have been re-subscribed to emails'
		: 'You have been unsubscribed from emails'}
	formTarget="/account/api?/toggleEmailSubscription"
	fields={[]}
/>
