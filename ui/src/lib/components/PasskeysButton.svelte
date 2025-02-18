<script lang="ts">
  import { PUBLIC_POCKETBASE_URL } from "$env/static/public";
  import { pb } from "$lib/pocketbase";
  import { goto } from "$app/navigation";
  import {
    startRegistration,
    startAuthentication,
  } from "@simplewebauthn/browser";

  let loading = $state(false);

  let { isSignUp } = $props();

  async function handleClick() {
    try {
      loading = true;

      const userName = "lol@gmail.com";

      if (isSignUp) {
        const response = await fetch(
          `${PUBLIC_POCKETBASE_URL}/api/pb-experiments/passkey/registerStart`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: userName }),
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

        if (verificationResponse.ok) {
          goto("/login/sign_in");
        } else {
        }
      } else {
        const response = await fetch(
          `${PUBLIC_POCKETBASE_URL}/api/pb-experiments/passkey/loginStart`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: userName }),
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
        if (verificationResponse.ok) {
          if (result.token) {
            pb.authStore.save(result.token, result.record);

            if (pb.authStore.isValid) {
              goto("/account");
            }
          }
        } else {
        }
      }

      loading = false;
    } catch (err) {
      loading = false;
    }
  }
</script>

<button
  disabled={loading}
  aria-label={(isSignUp ? "Sign up" : "Sign in") + " with Passkey"}
  class="btn btn-github {loading ? 'btn-disabled' : ''}"
  onclick={handleClick}
>
  <img
    alt="FIDO Passkey logo"
    width="21px"
    height="21px"
    src="/images/FIDO_Passkey_mark_A_black.jpg"
  />
  {isSignUp ? "Sign up" : "Sign in"} with Passkeys
</button>

<style>
  .btn-github {
    background-color: white;
    border-color: rgb(224, 224, 224);
  }
</style>
