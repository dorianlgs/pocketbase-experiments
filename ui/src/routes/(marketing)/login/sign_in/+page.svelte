<script lang="ts">
  import { ClientResponseError } from "pocketbase";
  import { pb } from "$lib/pocketbase";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import GitHubButton from "$lib/components/GitHubButton.svelte";
  import { onMount } from "svelte";
  import GoogleButton from "$lib/components/GoogleButton.svelte";
  import PasskeysButton from "$lib/components/PasskeysButton.svelte";

  let errors: { [fieldName: string]: string } = $state({});
  let loading = $state(false);
  let emailInput: HTMLInputElement | undefined = $state();

  const handleSubmit = async (e: SubmitEvent) => {
    e.preventDefault();
    errors = {};

    const formData = new FormData(e.target as HTMLFormElement);

    const email = formData.get("email")?.toString() ?? "";
    if (email.length < 6) {
      errors["email"] = "Email is required";
    } else if (email.length > 500) {
      errors["email"] = "Email too long";
    } else if (!email.includes("@") || !email.includes(".")) {
      errors["email"] = "Invalid email";
    }
    const password = formData.get("password")?.toString() ?? "";
    if (password.length > 500) {
      errors["password"] = "Password too long";
    }

    if (Object.keys(errors).length > 0) {
      return;
    }

    try {
      loading = true;
      await pb.collection("users").authWithPassword(email, password);
      loading = false;

      if (pb.authStore.isValid) {
        goto("/account");
      }
    } catch (err: any) {
      loading = false;
      if (err instanceof ClientResponseError) {
        const mfaId = err.response?.mfaId;

        if (mfaId) {
          //const result = await pb.collection("users").requestOTP(email);

          // if (result.otpId) {
          //  goto(`/login/otp_input?otpId=${result.otpId}&mfaId=${mfaId}`);
          //}

          goto(`/login/totp?mfaId=${mfaId}`);
        }

        if (err.response.message === "Failed to authenticate.") {
          errors["loginResult"] =
            "The Email or Password is Incorrect. Try again.";
        }
      }
    }
  };

  onMount(() => {
    if (emailInput) emailInput.focus();
  });
</script>

<svelte:head>
  <title>Sign in</title>
</svelte:head>

{#if page.url.searchParams.get("password_changed") === "true"}
  <div role="alert" class="alert alert-success mb-5">
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="stroke-current shrink-0 h-6 w-6"
      fill="none"
      viewBox="0 0 24 24"
      ><path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
      /></svg
    >
    <span>Password updated successfully</span>
  </div>
{/if}

{#if page.url.searchParams.get("verified") === "true"}
  <div role="alert" class="alert alert-success mb-5">
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="stroke-current shrink-0 h-6 w-6"
      fill="none"
      viewBox="0 0 24 24"
      ><path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
      /></svg
    >
    <span>Email verified! Please sign in.</span>
  </div>
{/if}
{#if page.url.searchParams.get("not_verified") === "true"}
  <div role="alert" class="alert alert-warning mb-5">
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="stroke-current shrink-0 h-6 w-6"
      fill="none"
      viewBox="0 0 24 24"
      ><path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
      /></svg
    >
    <span>Please verify your email.</span>
  </div>
{/if}
{#if page.url.searchParams.get("forgot_password") === "true"}
  <div role="alert" class="alert alert-success mb-5">
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="stroke-current shrink-0 h-6 w-6"
      fill="none"
      viewBox="0 0 24 24"
      ><path
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
      /></svg
    >
    <span>Please check your email in order to make the password reset.</span>
  </div>
{/if}
<h1 class="text-2xl font-bold mb-6">Sign In</h1>
<GoogleButton />
<br />
<hr class="solid" />
<br />
<GitHubButton />
<br />
<hr class="solid" />
<br />
<PasskeysButton isSignUp={false} />
<br />
<hr class="solid" />
<br />
<form class="form-widget flex flex-col" onsubmit={handleSubmit}>
  <label for={"email"}>
    <div class="flex flex-row">
      <div class="text-base font-bold">{"Email address"}</div>
      {#if errors["email"]}
        <div class="text-red-600 flex-grow text-sm ml-2 text-right">
          {errors["email"]}
        </div>
      {/if}
    </div>
    <input
      bind:this={emailInput}
      id={"email"}
      name={"email"}
      type={"email"}
      autocomplete={"email"}
      placeholder={"Your email address"}
      class="{errors['email']
        ? 'input-error'
        : ''} input-md mt-1 input input-bordered w-full mb-3 text-base py-4"
    />
  </label>
  <label for={"password"}>
    <div class="flex flex-row">
      <div class="text-base font-bold">{"Password"}</div>
      {#if errors["password"]}
        <div class="text-red-600 flex-grow text-sm ml-2 text-right">
          {errors["password"]}
        </div>
      {/if}
    </div>
    <input
      id={"password"}
      name={"password"}
      type={"password"}
      autocomplete={"off"}
      placeholder={"Your password"}
      class="{errors['email']
        ? 'input-error'
        : ''} input-md mt-1 input input-bordered w-full mb-3 text-base py-4"
    />
  </label>

  {#if Object.keys(errors).length > 0}
    {#if errors["loginResult"]}
      <p class="text-red-600 text-sm mb-2">{errors["loginResult"]}</p>
    {:else}
      <p class="text-red-600 text-sm mb-2">Please resolve above issues.</p>
    {/if}
  {/if}

  <button
    disabled={loading}
    class="btn btn-primary {loading ? 'btn-disabled' : ''}">Sign in</button
  >
</form>

<div class="text-l text-slate-800 mt-4">
  <a class="underline" href="/login/forgot_password">Forgot password?</a>
</div>
<div class="text-l text-slate-800 mt-3">
  Don't have an account? <a class="underline" href="/login/sign_up">Sign up</a>.
</div>
