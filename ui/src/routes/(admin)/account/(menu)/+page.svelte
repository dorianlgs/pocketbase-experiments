<script lang="ts">
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";

  let adminSection: Writable<string> = getContext("adminSection");
  adminSection.set("home");
</script>

<svelte:head>
  <title>Home</title>
</svelte:head>

<h1 class="text-2xl font-bold mb-1">Home</h1>

<div class="overflow-x-auto"></div>
