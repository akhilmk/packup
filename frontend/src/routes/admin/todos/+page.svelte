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

<div class="max-w-2xl mx-auto p-4 md:p-8">
  <header class="mb-8">
    <h1 class="text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 tracking-tight mb-2">
      My Personal Todos
    </h1>
    <p class="text-sm text-slate-500">Admin's personal task list</p>
  </header>

  {#if user}
    <TodoList {user} />
  {/if}
</div>
