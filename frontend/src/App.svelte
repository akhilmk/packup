<script lang="ts">
  import { onMount } from "svelte";
  import TodoList from "./lib/components/TodoList.svelte";
  import Login from "./lib/components/Login.svelte";
  import { api, type User } from "./lib/api";

  let user: User | null = null;
  let loading = true;

  onMount(async () => {
    try {
      user = await api.getMe();
    } catch (e) {
      console.log("Not authenticated");
    } finally {
      loading = false;
    }
  });

  async function handleLogout() {
    try {
      await api.logout();
      user = null;
    } catch (e) {
      console.error(e);
    }
  }
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800;900&display=swap" rel="stylesheet">
</svelte:head>

<main>
  {#if loading}
    <div class="h-screen flex items-center justify-center bg-slate-50">
      <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
    </div>
  {:else if user}
    <div class="min-h-screen bg-gradient-to-br from-slate-50 to-indigo-50/20">
      <div class="max-w-2xl mx-auto px-4 py-4 flex justify-between items-center">
        <div class="flex items-center space-x-3 bg-white/50 backdrop-blur-sm px-3 py-1.5 rounded-full border border-white/50">
          {#if user.avatar_url}
            <img src={user.avatar_url} alt={user.name} class="w-8 h-8 rounded-full border border-indigo-100" />
          {/if}
          <span class="text-sm font-bold text-slate-700 hidden sm:block">{user.name}</span>
          {#if user.role === 'admin'}
            <span class="text-[10px] font-bold tracking-wider text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded-full border border-indigo-100 uppercase">Admin</span>
          {/if}
        </div>
        
        <button 
          on:click={handleLogout}
          class="text-xs font-bold text-slate-500 hover:text-indigo-600 px-3 py-1.5 rounded-lg hover:bg-white/50 transition-colors"
        >
          Sign Out
        </button>
      </div>
      
      <TodoList />
    </div>
  {:else}
    <Login />
  {/if}
</main>

