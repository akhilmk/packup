<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { api, type User } from "$lib/api";

  let users: User[] = $state([]);
  let loading = $state(true);

  onMount(async () => {
    await loadUsers();
  });

  async function loadUsers() {
    loading = true;
    try {
      users = await api.listUsers();
    } catch (e) {
      console.error("Failed to load users", e);
    } finally {
      loading = false;
    }
  }

  function viewUserTodos(userId: string) {
    goto(`/admin/users/${userId}`);
  }
</script>

<div class="max-w-6xl mx-auto">
  <header class="mb-4">
    <h2 class="text-2xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 tracking-tight mb-1">
      Users Management
    </h2>
    <p class="text-sm text-slate-500">View and manage all users in the system</p>
  </header>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
    </div>
  {:else}
    <div class="glass-card rounded-2xl overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-slate-50 border-b border-slate-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-bold text-slate-600 uppercase tracking-wider">Name</th>
              <th class="px-6 py-3 text-left text-xs font-bold text-slate-600 uppercase tracking-wider">Email</th>
              <th class="px-6 py-3 text-left text-xs font-bold text-slate-600 uppercase tracking-wider">Role</th>
              <th class="px-6 py-3 text-left text-xs font-bold text-slate-600 uppercase tracking-wider">Joined</th>
              <th class="px-6 py-3 text-left text-xs font-bold text-slate-600 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            {#each users as user}
              <tr class="hover:bg-indigo-50/30 transition-colors">
                <td class="px-6 py-4">
                  <div class="flex items-center space-x-3">
                    {#if user.avatar_url}
                      <img src={user.avatar_url} alt={user.name} class="w-8 h-8 rounded-full" />
                    {/if}
                    <span class="font-medium text-slate-800">{user.name}</span>
                  </div>
                </td>
                <td class="px-6 py-4 text-sm text-slate-600">{user.email}</td>
                <td class="px-6 py-4">
                  <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold {user.role === 'admin' ? 'bg-indigo-100 text-indigo-700' : 'bg-slate-100 text-slate-700'}">
                    {user.role}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-slate-600">
                  {new Date(user.created_at).toLocaleDateString()}
                </td>
                <td class="px-6 py-4">
                  <button
                    onclick={() => viewUserTodos(user.id)}
                    class="text-sm font-bold text-indigo-600 hover:text-indigo-700 hover:underline"
                  >
                    View Todos
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>
