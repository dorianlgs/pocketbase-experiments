<script lang="ts">
  import { goto } from "$app/navigation";

  let { isSignUp } = $props();

  function handleClick() {
    goto(`/login/passkey?is_sign_up=${isSignUp}`);
  }
</script>

<button
  aria-label={(isSignUp ? "Sign up" : "Sign in") + " with Passkey"}
  class="btn btn-passkey"
  onclick={handleClick}
>
  <img
    alt="FIDO Passkey logo"
    width="21px"
    height="21px"
    src="/images/FIDO_Passkey_mark_A_black.jpg"
  />
  {isSignUp ? "Sign up" : "Sign in"} with Passkey
</button>

<style>
  .btn-passkey {
    background-color: white;
    border-color: rgb(224, 224, 224);
  }
</style>
