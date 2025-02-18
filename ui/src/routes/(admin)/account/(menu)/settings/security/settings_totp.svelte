<script lang="ts">
  import { PUBLIC_POCKETBASE_URL } from "$env/static/public";
  import { currentUser } from "$lib/stores/user";
  import { onMount } from "svelte";
  import { pb } from "$lib/pocketbase";

  let title: string = $state("Totp");
  let dangerous: boolean = $state(false);
  let regenerate: boolean = $state(false);
  let localImage: string = $state("");
  let imageUrl: string = $derived(
    `${PUBLIC_POCKETBASE_URL}/api/pb-experiments/get-qr?userId=${$currentUser?.id}&regenerate=${regenerate}`,
  );

  async function getImage() {
    try {
      const response = await fetch(imageUrl, {
        headers: { Authorization: `Bearer ${pb.authStore.token}` },
      });
      const blob = await response.blob();
      const tmp = URL.createObjectURL(blob);
      localImage = tmp;
    } catch (err) {}
  }

  onMount(async () => {
    if ($currentUser?.multiFactorAuth) await getImage();
  });

  async function handleClick() {
    regenerate = true;
    await getImage();
  }
</script>

<div class="card p-6 pb-7 mt-8 max-w-xl flex flex-col md:flex-row shadow">
  {#if title}
    <div class="text-xl font-bold mb-3 w-48 md:pr-8 flex-none">{title}</div>
  {/if}

  <div class="w-full min-w-48">
    {#if localImage}
      <img src={localImage} alt="Dog" />
    {/if}
  </div>
  <div class="w-full min-w-48 paddo">
    <button
      onclick={handleClick}
      class="ml-auto btn btn-sm mt-3 min-w-[145px] {dangerous
        ? 'btn-error'
        : 'btn-primary btn-outline'}"
    >
      Generate New
    </button>
  </div>
</div>

<style>
  .paddo {
    padding-left: 10px;
  }
</style>
