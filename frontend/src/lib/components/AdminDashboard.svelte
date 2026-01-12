<script lang="ts">
  import { onMount } from "svelte";
  import { api, type Todo, type User } from "../api";

  type View = 'users' | 'admin-todos' | 'user-todos';
  
  let currentView: View = 'users';
  let users: User[] = [];
  let adminTodos: Todo[] = [];
  let userTodos: Todo[] = [];
  let selectedUser: User | null = null;
  let loading = true;

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

  async function loadAdminTodos() {
    loading = true;
    try {
      adminTodos = await api.listAdminTodos();
    } catch (e) {
      console.error("Failed to load admin todos", e);
    } finally {
      loading = false;
    }
  }

  async function loadUserTodos(user: User) {
    selectedUser = user;
    loading = true;
    try {
      userTodos = await api.listUserTodos(user.id);
    } catch (e) {
      console.error("Failed to load user todos", e);
    } finally {
      loading = false;
    }
  }

  function switchView(view: View) {
    currentView = view;
    selectedUser = null;
    if (view === 'users') {
      loadUsers();
    } else if (view === 'admin-todos') {
      loadAdminTodos();
    }
  }
</script>

<div class="max-w-6xl mx-auto p-4 md:p-8">
  <header class="mb-8">
    <h1 class="text-3xl font-black text-slate-800 mb-6">Admin Dashboard</h1>
    
    <!-- View Tabs -->
    <div class="flex gap-2 border-b border-slate-200">
      <button
        class="px-6 py-3 font-bold text-sm transition-all {currentView === 'users' ? 'text-indigo-600 border-b-2 border-indigo-600' : 'text-slate-500 hover:text-slate-700'}"
        on:click={() => switchView('users')}
      >
        Users
      </button>
      <button
        class="px-6 py-3 font-bold text-sm transition-all {currentView === 'admin-todos' ? 'text-indigo-600 border-b-2 border-indigo-600' : 'text-slate-500 hover:text-slate-700'}"
        on:click={() => switchView('admin-todos')}
      >
        Admin Todos
      </button>
    </div>
  </header>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
    </div>
  {:else if currentView === 'users'}
    <!-- Users List -->
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
                    on:click={() => { currentView = 'user-todos'; loadUserTodos(user); }}
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
  {:else if currentView === 'admin-todos'}
    <!-- Admin Todos List -->
    <div class="glass-card rounded-3xl overflow-hidden border-border divide-y divide-gray-50">
      {#if adminTodos.length === 0}
        <div class="p-12 text-center">
          <p class="text-slate-500">No admin todos yet</p>
        </div>
      {:else}
        {#each adminTodos as todo}
          <div class="p-4 flex items-center gap-4">
            <div class="flex items-center justify-center w-6 h-6 rounded-full border-2 {todo.status === 'done' ? 'bg-emerald-500 border-emerald-500' : todo.status === 'in-progress' ? 'bg-amber-100 border-amber-400' : 'bg-white border-slate-200'}">
              {#if todo.status === 'done'}
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
              {:else if todo.status === 'in-progress'}
                <div class="w-2.5 h-2.5 bg-amber-500 rounded-full animate-pulse"></div>
              {/if}
            </div>
            <span class="flex-1 font-medium text-slate-700 {todo.status === 'done' ? 'line-through opacity-60' : ''}">
              {todo.text}
            </span>
            <span class="text-xs text-slate-400">
              {new Date(todo.created).toLocaleDateString()}
            </span>
          </div>
        {/each}
      {/if}
    </div>
  {:else if currentView === 'user-todos' && selectedUser}
    <!-- User Todos View -->
    <div class="mb-6">
      <button
        on:click={() => switchView('users')}
        class="text-sm font-bold text-indigo-600 hover:text-indigo-700 flex items-center gap-2"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        Back to Users
      </button>
    </div>

    <div class="glass-card rounded-2xl p-6 mb-6">
      <div class="flex items-center space-x-4">
        {#if selectedUser.avatar_url}
          <img src={selectedUser.avatar_url} alt={selectedUser.name} class="w-16 h-16 rounded-full" />
        {/if}
        <div>
          <h2 class="text-2xl font-bold text-slate-800">{selectedUser.name}</h2>
          <p class="text-slate-600">{selectedUser.email}</p>
        </div>
      </div>
    </div>

    <div class="glass-card rounded-3xl overflow-hidden border-border divide-y divide-gray-50">
      {#if userTodos.length === 0}
        <div class="p-12 text-center">
          <p class="text-slate-500">No todos yet</p>
        </div>
      {:else}
        {#each userTodos as todo}
          <div class="p-4 flex items-center gap-4">
            <div class="flex items-center justify-center w-6 h-6 rounded-full border-2 {todo.status === 'done' ? 'bg-emerald-500 border-emerald-500' : todo.status === 'in-progress' ? 'bg-amber-100 border-amber-400' : 'bg-white border-slate-200'}">
              {#if todo.status === 'done'}
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4">
                  <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
              {:else if todo.status === 'in-progress'}
                <div class="w-2.5 h-2.5 bg-amber-500 rounded-full animate-pulse"></div>
              {/if}
            </div>
            <span class="flex-1 font-medium text-slate-700 {todo.status === 'done' ? 'line-through opacity-60' : ''}">
              {todo.text}
              {#if todo.is_admin_todo}
                <span class="inline-flex items-center ml-2 text-[10px] font-bold tracking-wider text-purple-600 bg-purple-50 px-2 py-0.5 rounded-full border border-purple-100 uppercase align-middle transform -translate-y-0.5">
                  Admin Task
                </span>
              {/if}
            </span>
            <span class="text-xs text-slate-400">
              {new Date(todo.created).toLocaleDateString()}
            </span>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>
