<script lang="ts">
	import { getContext } from 'svelte';
	import type { Writable } from 'svelte/store';
	import SettingsModule from '../settings/settings_module.svelte';
	import PricingModule from '../../../../(marketing)/pricing/pricing_module.svelte';
	import { pricingPlans, defaultPlanId } from '../../../../(marketing)/pricing/pricing_plans';

	let adminSection: Writable<string> = getContext('adminSection');
	adminSection.set('billing');

	interface Props {
		currentPlanId: string;
		isActiveCustomer: boolean;
		hasEverHadSubscription: boolean;
	}

	let {
		currentPlanId: currentPlanIdProps,
		isActiveCustomer,
		hasEverHadSubscription
	}: Props = $props();

	let currentPlanId = currentPlanIdProps ?? defaultPlanId;
	let currentPlanName = pricingPlans.find((x) => x.id === currentPlanIdProps)?.name;
</script>

<svelte:head>
	<title>Billing</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-2">
	{isActiveCustomer ? 'Billing' : 'Select a Plan'}
</h1>
<div>
	View our <a href="/pricing" target="_blank" class="link">pricing page</a> for details.
</div>

{#if !isActiveCustomer}
	<div class="mt-8">
		<PricingModule {currentPlanId} callToAction="Select Plan" center={false} />
	</div>

	{#if hasEverHadSubscription}
		<div class="mt-10">
			<a href="/account/billing/manage" class="link">View past invoices</a>
		</div>
	{/if}
{:else}
	<SettingsModule
		title="Subscription"
		editable={false}
		fields={[
			{
				id: 'plan',
				label: 'Current Plan',
				initialValue: currentPlanName || ''
			}
		]}
		editButtonTitle="Manage Subscription"
		editLink="/account/billing/manage"
	/>
{/if}
