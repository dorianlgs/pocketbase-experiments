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

    const totpCode = formData.get("totpCode")?.toString() ?? "";
    if (!totpCode) {
      errors["totpCode"] = "Totp code required";
    }

    const mfaId = page.url.searchParams.get("mfaId") as string;

    if (!mfaId) {
      errors["paramsError"] = "Empty mfaId";
    }

    if (Object.keys(errors).length > 0) {
      return;
    }

    try {
      loading = true;

      const result = await pb.send("/api/pb-experiments/totp-login", {
        method: "POST",
        body: JSON.stringify({
          passcode: totpCode,
          mfaId: mfaId,
        }),
      });

      loading = false;

      if (result.token) {
        pb.authStore.save(result.token, result.record);

        if (pb.authStore.isValid) {
          goto("/account");
        }
      }
    } catch (err: any) {
      if (err.status === 401) {
        errors["totpCode"] = "Invalid code";
      }

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
  <label for={"totpCode"}>
    <div class="flex flex-row">
      <div class="text-base font-bold">{"Totp code"}</div>
      {#if errors["totpCode"]}
        <div class="text-red-600 flex-grow text-sm ml-2 text-right">
          {errors["totpCode"]}
        </div>
      {/if}
    </div>
    <input
      bind:this={emailCodeInput}
      id={"totpCode"}
      name={"totpCode"}
      autocomplete={"off"}
      placeholder={"Your totp code"}
      class="{errors['totpCode']
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
