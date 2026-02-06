<script lang="ts">
  import { onMount, type Snippet } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { api, type User } from "$lib/api";
  import { loadRuntimeConfig, clearConfigCache } from "$lib/config";
  import Logo from "$lib/components/Logo.svelte";
  import ChatAssistant from "$lib/components/ChatAssistant.svelte";
  import "../app.css";

  let { children }: { children: Snippet } = $props();

  let user: User | null = $state(null);
  let loading = $state(true);

  // Reactive chat max width based on route
  const isAdminRoute = $derived($page.url.pathname.startsWith('/admin'));
  const chatMaxWidth = $derived(isAdminRoute ? "max-w-6xl" : "max-w-2xl");

  onMount(async () => {
    try {
      user = await api.getMe();
      // Load runtime configuration from backend (now that we are authenticated)
      loadRuntimeConfig();
      
      // Redirect based on role if on login page and authenticated
      if ($page.url.pathname === '/' && user) {
        if (user.role === 'admin') {
          goto('/admin');
        } else {
          goto('/todos');
        }
      }
    } catch (e) {
      console.log("Not authenticated");
      // Redirect to login if not on login page and not authenticated
      if ($page.url.pathname !== '/') {
        goto('/');
      }
    } finally {
      loading = false;
    }
  });

  async function handleLogout() {
    try {
      await api.logout();
      clearConfigCache();
      sessionStorage.removeItem('packup_chat_messages');
      user = null;
      goto('/');
    } catch (e) {
      console.error(e);
    }
  }
</script>

{#if loading}
  <div class="h-screen flex items-center justify-center bg-slate-50">
    <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
  </div>
{:else if user}
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-indigo-50/20">
    <div class="max-w-6xl mx-auto px-4 py-4 flex justify-between items-center">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-lg flex items-center justify-center shadow-md">
          <Logo className="w-4 h-4 text-white" />
        </div>
        <div class="hidden sm:block">
          <div class="flex items-center gap-1.5 leading-none">
            <span class="text-lg font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 tracking-tight">PackUp</span>
            <div class="flex items-center mt-0.5">
              <span class="text-[7px] font-bold uppercase tracking-widest text-slate-400 mr-1">by</span>
              <span class="text-[9px] font-bold tracking-wide text-slate-600" style="font-family: 'Montserrat', sans-serif;">
                Orama Holidays
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center space-x-3 bg-white/50 backdrop-blur-sm px-3 py-1.5 rounded-full border border-white/50">
        {#if user.avatar_url}
          <img src={user.avatar_url} alt={user.name} class="w-8 h-8 rounded-full border border-indigo-100" />
        {/if}
        <span class="text-sm font-bold text-slate-700 hidden sm:block">{user.name}</span>
        {#if user.role === 'admin'}
          <span class="text-[10px] font-bold tracking-wider text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded-full border border-indigo-100 uppercase">Admin</span>
        {/if}
      </div>
      
      <div class="flex items-center gap-3">
        {#if user.role === 'admin'}
          <nav class="flex items-center gap-2">
            <a 
              href="/todos"
              class="text-xs font-bold px-3 py-1.5 rounded-lg transition-colors {$page.url.pathname === '/todos' ? 'bg-indigo-100 text-indigo-700' : 'bg-white/50 text-slate-600 hover:bg-white'}"
            >
              My Todos
            </a>
            <a 
              href="/admin"
              class="text-xs font-bold px-3 py-1.5 rounded-lg transition-colors {$page.url.pathname.startsWith('/admin') ? 'bg-indigo-100 text-indigo-700' : 'bg-white/50 text-slate-600 hover:bg-white'}"
            >
              Admin Hub
            </a>
          </nav>
        {/if}
        <button 
          onclick={handleLogout}
          class="text-xs font-bold text-slate-500 hover:text-indigo-600 px-3 py-1.5 rounded-lg hover:bg-white/50 transition-colors"
        >
          Sign Out
        </button>
      </div>
    </div>
    
    {@render children()}

    <ChatAssistant maxWidth={chatMaxWidth} />
  </div>
{:else}
  {@render children()}
{/if}
