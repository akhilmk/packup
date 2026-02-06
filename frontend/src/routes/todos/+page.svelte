<script lang="ts">
  import { onMount } from "svelte";
  import { api, type User } from "$lib/api";
  import TodoList from "$lib/components/TodoList.svelte";

  let user: User | null = $state(null);

  onMount(async () => {
    try {
      user = await api.getMe();
    } catch (e) {
      console.error("Failed to get user", e);
    }
  });
</script>

{#if user}
  <TodoList {user} />
{/if}
