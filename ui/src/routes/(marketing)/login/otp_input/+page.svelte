<script lang="ts">
  import { pb } from "$lib/pocketbase";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { page } from "$app/state";

  let errors: { [fieldName: string]: string } = $state({});
  let loading = $state(false);
  let emailCodeInput: HTMLInputElement | undefined = $state();

  const handleSubmit = async (e: SubmitEvent) => {
    e.preventDefault();
    errors = {};

    const formData = new FormData(e.target as HTMLFormElement);

    const emailCode = formData.get("emailCode")?.toString() ?? "";
    if (!emailCode) {
      errors["emailCode"] = "Email code required";
    }

    const mfaId = page.url.searchParams.get("mfaId") as string;

    if (!mfaId) {
      errors["paramsError"] = "Empty mfaId";
    }

    const otpId = page.url.searchParams.get("otpId") as string;

    if (!otpId) {
      errors["paramsError"] = "Empty otpId";
    }

    if (Object.keys(errors).length > 0) {
      return;
    }

    try {
      loading = true;
      await pb
        .collection("users")
        .authWithOTP(otpId, emailCode, { mfaId: mfaId });
      loading = false;

      if (pb.authStore.isValid) {
        goto("/account");
      }
    } catch (err: any) {
      loading = false;
    }
  };

  onMount(() => {
    if (emailCodeInput) emailCodeInput.focus();
  });
</script>

<svelte:head>
  <title>2-Step Verification</title>
</svelte:head>
<h1 class="text-2xl font-bold mb-6">2-Step Verification</h1>
<form class="form-widget flex flex-col" onsubmit={handleSubmit}>
  <label for={"emailCode"}>
    <div class="flex flex-row">
      <div class="text-base font-bold">{"Email code"}</div>
      {#if errors["emailCode"]}
        <div class="text-red-600 flex-grow text-sm ml-2 text-right">
          {errors["emailCode"]}
        </div>
      {/if}
    </div>
    <input
      bind:this={emailCodeInput}
      id={"emailCode"}
      name={"emailCode"}
      autocomplete={"off"}
      placeholder={"Your email code"}
      class="{errors['emailCode']
        ? 'input-error'
        : ''} input-md mt-1 input input-bordered w-full mb-3 text-base py-4"
    />
  </label>
  {#if Object.keys(errors).length > 0}
    {#if errors["paramsError"]}
      <p class="text-red-600 text-sm mb-2">{errors["paramsError"]}</p>
    {:else}
      <p class="text-red-600 text-sm mb-2">Please resolve above issues.</p>
    {/if}
  {/if}

  <button
    disabled={loading}
    class="btn btn-primary {loading ? 'btn-disabled' : ''}">Sign in</button
  >
</form>
