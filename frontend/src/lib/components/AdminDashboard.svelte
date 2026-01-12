<script lang="ts">
  import { onMount } from "svelte";
  import { api, type Todo, type User } from "../api";
  import StatusIndicator from "./StatusIndicator.svelte";

  type View = 'users' | 'admin-todos' | 'user-todos';
  
  let currentView: View = 'users';
  let users: User[] = [];
  let adminTodos: Todo[] = [];
  let userTodos: Todo[] = [];
  let selectedUser: User | null = null;
  let loading = true;
  
  // Admin todo management
  let newAdminTodoText = "";
  let editingTodoId: string | null = null;
  let editingTodoText = "";

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
      adminTodos = await api.listDefaultTasks();
    } catch (e) {
      console.error("Failed to load default tasks", e);
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

  async function handleCreateAdminTodo() {
    if (!newAdminTodoText.trim()) return;
    
    try {
      await api.createDefaultTask(newAdminTodoText);
      newAdminTodoText = "";
      await loadAdminTodos();
    } catch (e) {
      console.error("Failed to create default task", e);
    }
  }

  function startEdit(todo: Todo) {
    editingTodoId = todo.id;
    editingTodoText = todo.text;
  }

  function cancelEdit() {
    editingTodoId = null;
    editingTodoText = "";
  }

  async function saveEdit(id: string) {
    if (!editingTodoText.trim()) return;
    
    try {
      await api.updateDefaultTask(id, editingTodoText);
      editingTodoId = null;
      editingTodoText = "";
      await loadAdminTodos();
    } catch (e) {
      console.error("Failed to update default task", e);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this default task? It will be removed for all users.")) return;
    
    try {
      await api.deleteDefaultTask(id);
      await loadAdminTodos();
    } catch (e) {
      console.error("Failed to delete default task", e);
    }

  }

  async function handleCycleUserTodoStatus(todo: Todo) {
    if (!selectedUser) return;
    
    // Determine next status
    let nextStatus: 'pending' | 'in-progress' | 'done';
    if (todo.status === 'pending') nextStatus = 'in-progress';
    else if (todo.status === 'in-progress') nextStatus = 'done';
    else nextStatus = 'pending';

    try {
      if (todo.is_default_task) {
        await api.updateUserTodo(selectedUser.id, todo.id, { status: nextStatus });
        // Refresh list
        await loadUserTodos(selectedUser);
      } else {
        alert("Admins cannot change status of personal tasks.");
      }
    } catch (e) {
      console.error("Failed to update user todo status", e);
    }
  }

  // Admin user task creation
  let newUserTodoText = "";
  let newUserTodoHidden = false;

  async function handleCreateUserTodo() {
    if (!newUserTodoText.trim() || !selectedUser) return;

    try {
      await api.createUserTodo(selectedUser.id, newUserTodoText, newUserTodoHidden);
      newUserTodoText = "";
      newUserTodoHidden = false;
      await loadUserTodos(selectedUser);
    } catch (e) {
      console.error("Failed to create user todo", e);
    }
  }

  async function toggleUserTodoHidden(todo: Todo) {
    if (!selectedUser) return;
    try {
      await api.updateUserTodo(selectedUser.id, todo.id, { hidden_from_user: !todo.hidden_from_user });
      await loadUserTodos(selectedUser);
    } catch (e) {
      console.error("Failed to toggle hidden status", e);
    }
  }

  // User todo editing
  let editingUserTodoId: string | null = null;
  let editingUserTodoText = "";

  function startEditUserTodo(todo: Todo) {
    editingUserTodoId = todo.id;
    editingUserTodoText = todo.text;
  }

  function cancelEditUserTodo() {
    editingUserTodoId = null;
    editingUserTodoText = "";
  }

  async function saveEditUserTodo(todo: Todo) {
    if (!selectedUser || !editingUserTodoText.trim()) return;
    try {
      await api.updateUserTodo(selectedUser.id, todo.id, { text: editingUserTodoText });
      editingUserTodoId = null;
      editingUserTodoText = "";
      await loadUserTodos(selectedUser);
    } catch (e) {
      console.error("Failed to update user todo", e);
    }
  }

  async function handleDeleteUserTodo(todo: Todo) {
    if (!selectedUser) return;
    if (!confirm("Delete this task?")) return;
    
    try {
      await api.deleteUserTodo(selectedUser.id, todo.id);
      await loadUserTodos(selectedUser);
    } catch (e) {
      console.error("Failed to delete user todo", e);
    }
  }
</script>

<div class="max-w-6xl mx-auto p-4 md:p-8">
  <header class="mb-8">
    <!-- Itinera Icon and Title -->
    <div class="flex items-center gap-4 mb-6">
      <div class="w-12 h-12 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 11l1 1 2-2" />
        </svg>
      </div>
      <div>
        <h1 class="text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600">Itinera</h1>
        <p class="text-sm text-slate-500 font-medium">Admin Dashboard</p>
      </div>
    </div>
    
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
        Default Tasks
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
    <!-- Admin Todos Management -->
    <div class="space-y-6">
      <!-- Add Admin Todo Form -->
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-bold text-slate-800 mb-2">Add Default Task</h2>
        <p class="text-sm text-slate-500 mb-4">
          This task will be automatically assigned to <strong>all users</strong>. Each user tracks their own status for this task.
        </p>
        <form on:submit|preventDefault={handleCreateAdminTodo} class="flex gap-3">
          <input
            bind:value={newAdminTodoText}
            placeholder="Enter task for all users..."
            maxlength="200"
            class="flex-1 px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500"
          />
          <button
            type="submit"
            class="px-6 py-2 bg-indigo-600 text-white font-bold rounded-lg hover:bg-indigo-700 transition-colors"
          >
            Add Task
          </button>
        </form>
      </div>

      <!-- Admin Todos List -->
      <div class="glass-card rounded-2xl overflow-hidden divide-y divide-slate-100">
        {#if adminTodos.length === 0}
          <div class="p-12 text-center">
            <p class="text-slate-500">No default tasks yet. Create one above!</p>
          </div>
        {:else}
          {#each adminTodos as todo}
            <div class="p-4 flex items-center gap-4 hover:bg-slate-50 transition-colors">
              {#if editingTodoId === todo.id}
                <!-- Edit Mode -->
                <form on:submit|preventDefault={() => saveEdit(todo.id)} class="flex-1 flex gap-2">
                  <input
                    bind:value={editingTodoText}
                    maxlength="200"
                    class="flex-1 px-3 py-1.5 border border-indigo-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                    autofocus
                  />
                  <button
                    type="submit"
                    class="p-2 text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors"
                    title="Save"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                  </button>
                  <button
                    type="button"
                    on:click={cancelEdit}
                    class="p-2 text-slate-400 hover:bg-slate-100 rounded-lg transition-colors"
                    title="Cancel"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="18" y1="6" x2="6" y2="18"></line>
                      <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                  </button>
                </form>
              {:else}
                <!-- View Mode -->
                <span class="flex-1 font-medium text-slate-700">
                  {todo.text}
                </span>
                <span class="text-xs text-slate-400">
                  {new Date(todo.created).toLocaleDateString()}
                </span>
                <div class="flex items-center gap-1">
                  <button
                    on:click={() => startEdit(todo)}
                    class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                    title="Edit"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"></path>
                      <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                    </svg>
                  </button>
                  <button
                    on:click={() => handleDelete(todo.id)}
                    class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                    title="Delete"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="3 6 5 6 21 6"></polyline>
                      <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"></path>
                    </svg>
                  </button>
                </div>
              {/if}
            </div>
          {/each}
        {/if}
      </div>
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

    <!-- Add Task for User -->
    <div class="glass-card rounded-2xl p-6 mb-6">
      <h3 class="text-lg font-bold text-slate-800 mb-4">Add Task for {selectedUser.name}</h3>
      <form on:submit|preventDefault={handleCreateUserTodo} class="flex gap-3">
        <input
          bind:value={newUserTodoText}
          placeholder="Enter task text..."
          maxlength="200"
          class="flex-1 px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500"
        />
        <label class="flex items-center gap-2 cursor-pointer select-none group relative">
          <input type="checkbox" bind:checked={newUserTodoHidden} class="sr-only peer" />
          <div class="relative w-11 h-6 bg-slate-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600"></div>
          <span class="text-sm font-medium text-slate-600 group-hover:text-slate-900 transition-colors">Hide from user</span>
        </label>
        <button
          type="submit"
          class="px-6 py-2 bg-indigo-600 text-white font-bold rounded-lg hover:bg-indigo-700 transition-colors"
        >
          Add
        </button>
      </form>
    </div>

    <div class="glass-card rounded-3xl overflow-hidden border-border divide-y divide-gray-50">
      {#if userTodos.length === 0}
        <div class="p-12 text-center">
          <p class="text-slate-500">No todos yet</p>
        </div>
      {:else}
        {#each userTodos as todo}
          <div class="p-4 flex items-center gap-4">
            <!-- Status Indicator / Initial for Admin View -->
            {#if todo.is_default_task}
              <StatusIndicator 
                status={todo.status} 
                clickable={true} 
                on:click={() => handleCycleUserTodoStatus(todo)}
              />
            {:else}
               <!-- Read-only indicator for shared personal tasks -->
               <StatusIndicator 
                 status={todo.status} 
                 clickable={false}
               />
            {/if}
            {#if editingUserTodoId === todo.id}
              <form on:submit|preventDefault={() => saveEditUserTodo(todo)} class="flex-1 flex gap-2">
                <input
                  bind:value={editingUserTodoText}
                  maxlength="200"
                  class="flex-1 px-3 py-1.5 border border-indigo-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                  autofocus
                />
                <button
                  type="submit"
                  class="p-2 text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors"
                  title="Save"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                </button>
                <button
                  type="button"
                  on:click={cancelEditUserTodo}
                  class="p-2 text-slate-400 hover:bg-slate-100 rounded-lg transition-colors"
                  title="Cancel"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                </button>
              </form>
            {:else}
              <span class="flex-1 font-medium text-slate-700 {todo.status === 'done' ? 'line-through opacity-60' : ''}">
                {todo.text}
                {#if todo.is_default_task}
                  <span class="inline-flex items-center ml-2 text-[10px] font-bold tracking-wider text-purple-600 bg-purple-50 px-2 py-0.5 rounded-full border border-purple-100 uppercase align-middle transform -translate-y-0.5">
                    Default Task
                  </span>
                {/if}
                {#if !todo.is_default_task && todo.user_id === selectedUser?.id}
                  {#if todo.created_by_user_id === selectedUser?.id}
                    <!-- User-created task (shared with admin) -->
                    <span class="inline-flex items-center ml-2 text-[10px] font-bold tracking-wider text-green-600 bg-green-50 px-2 py-0.5 rounded-full border border-green-100 uppercase align-middle transform -translate-y-0.5">
                      User Created
                    </span>
                  {:else}
                    <!-- Admin-created task for user -->
                    <span class="inline-flex items-center ml-2 text-[10px] font-bold tracking-wider text-blue-600 bg-blue-50 px-2 py-0.5 rounded-full border border-blue-100 uppercase align-middle transform -translate-y-0.5">
                      Admin Shared
                    </span>
                    {#if todo.hidden_from_user}
                      <span class="inline-flex items-center ml-1 text-[10px] font-bold tracking-wider text-slate-500 bg-slate-100 px-2 py-0.5 rounded-full border border-slate-200 uppercase align-middle transform -translate-y-0.5">
                        Hidden
                      </span>
                    {/if}
                  {/if}
                {/if}
              </span>
            {/if}
            <div class="flex items-center gap-2">
              {#if !todo.is_default_task && editingUserTodoId !== todo.id && todo.created_by_user_id !== selectedUser?.id}
                <!-- Only show edit/hide/delete for admin-created tasks -->
                <button
                  on:click={() => startEditUserTodo(todo)}
                  class="p-1 rounded hover:bg-slate-100 text-slate-400 hover:text-indigo-600"
                  title="Edit task"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"></path>
                    <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                  </svg>
                </button>
                <button 
                  on:click={() => toggleUserTodoHidden(todo)}
                  class="p-1 rounded hover:bg-slate-100 text-slate-400 hover:text-indigo-600"
                  title={todo.hidden_from_user ? "Show to user" : "Hide from user"}
                >
                  {#if todo.hidden_from_user}
                   <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                      <circle cx="12" cy="12" r="3"></circle>
                    </svg>
                  {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
                      <line x1="1" y1="1" x2="23" y2="23"></line>
                    </svg>
                  {/if}
                </button>
                <button 
                  on:click={() => handleDeleteUserTodo(todo)}
                  class="p-1 rounded hover:bg-red-50 text-slate-400 hover:text-red-500"
                  title="Delete task"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"></path>
                  </svg>
                </button>
              {/if}
              <span class="text-xs text-slate-400">
                {new Date(todo.created).toLocaleDateString()}
              </span>
            </div>
          </div>
        {/each}
      {/if}
    </div>
  {/if}
</div>
