<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount, untrack } from "svelte";
  import { PUBLIC_POCKETBASE_URL } from "$env/static/public";
  import { pb } from "$lib/pocketbase";
  import { page } from "$app/state";

  import {
    startRegistration,
    startAuthentication,
  } from "@simplewebauthn/browser";

  const isSignUpParam = page.url.searchParams.get("is_sign_up");
  const isSignUp: boolean = isSignUpParam == "true";

  console.log({ isSignUp });

  let errors: { [fieldName: string]: string } = $state({});
  let loading = $state(false);
  let emailInput: HTMLInputElement | undefined = $state();

  let descriptionText: string = $derived(
    (isSignUp ? "Create" : "Sign in with") + " Passkey",
  );

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

    if (Object.keys(errors).length > 0) {
      return;
    }

    try {
      loading = true;

      if (isSignUp) {
        const response = await fetch(
          `${PUBLIC_POCKETBASE_URL}/api/pb-experiments/passkey/registerStart`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email: email }),
          },
        );

        if (!response.ok) {
          const msg = await response.json();
          throw new Error(
            "User already exists or failed to get registration options from server: " +
              msg,
          );
        }

        const options = await response.json();

        const attestationResponse = await startRegistration({
          optionsJSON: options.publicKey,
        });

        const sessionKey = response.headers.get("Session-Key") as string;

        // Send attestationResponse back to server for verification and storage.
        const verificationResponse = await fetch(
          `${PUBLIC_POCKETBASE_URL}/api/pb-experiments/passkey/registerFinish`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "Session-Key": sessionKey,
            },
            body: JSON.stringify(attestationResponse),
          },
        );

        const msg = await verificationResponse.json();

        if (!verificationResponse.ok) {
          errors["createPasskeyResult"] = msg;
          return;
        }

        goto("/login/sign_in");
      } else {
        const response = await fetch(
          `${PUBLIC_POCKETBASE_URL}/api/pb-experiments/passkey/loginStart`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email: email }),
          },
        );
        // Check if the login options are ok.
        if (!response.ok) {
          const msg = await response.json();
          throw new Error("Failed to get login options from server: " + msg);
        }
        // Convert the login options to JSON.
        const options = await response.json();

        // This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
        // A new assertionResponse is created. This also means that the challenge has been signed.
        const assertionResponse = await startAuthentication({
          optionsJSON: options.publicKey,
        });

        const loginKey = response.headers.get("Login-Key") as string;

        // Send assertionResponse back to server for verification.
        const verificationResponse = await fetch(
          `${PUBLIC_POCKETBASE_URL}/api/pb-experiments/passkey/loginFinish`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "Login-Key": loginKey,
            },
            body: JSON.stringify(assertionResponse),
          },
        );

        const result = await verificationResponse.json();

        if (!verificationResponse.ok) {
          errors["createPasskeyResult"] = result;
          return;
        }

        if (result.token) {
          pb.authStore.save(result.token, result.record);

          if (pb.authStore.isValid) {
            goto("/account");
          }
        }
      }

      loading = false;
    } catch (err: any) {
      loading = false;
      errors["createPasskeyResult"] = err.toString();
    }
  };

  onMount(() => {
    if (emailInput) emailInput.focus();
  });
</script>

<svelte:head>
  <title>{descriptionText}</title>
</svelte:head>
<h1 class="text-2xl font-bold mb-6">
  {descriptionText}
</h1>
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
  {#if Object.keys(errors).length > 0}
    {#if errors["createPasskeyResult"]}
      <p class="text-red-600 text-sm mb-2">{errors["createPasskeyResult"]}</p>
    {:else}
      <p class="text-red-600 text-sm mb-2">Please resolve above issues.</p>
    {/if}
  {/if}

  <button aria-label={descriptionText} class="btn btn-passkey">
    <img
      alt="FIDO Passkey logo"
      width="21px"
      height="21px"
      src="/images/FIDO_Passkey_mark_A_black.jpg"
    />
    {descriptionText}
  </button>
</form>

<style>
  .btn-passkey {
    background-color: white;
    border-color: rgb(224, 224, 224);
  }
</style>
