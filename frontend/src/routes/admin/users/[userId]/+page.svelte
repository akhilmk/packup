<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { api, type Todo, type TodoStatus, type User } from "$lib/api";
  import { goto } from "$app/navigation";
  import StatusIndicator from "$lib/components/StatusIndicator.svelte";

  let userId = $derived($page.params.userId);
  let user: User | undefined = $state(undefined);
  // ... (rest of imports and variables same as before)
  let todos: Todo[] = $state([]);
  let loading = $state(true);
  
  // Add Task State
  let newTaskText = $state("");
  let hideFromUser = $state(false);
  let isSubmitting = $state(false);

  // Filter State
  let activeFilter: 'all' | TodoStatus = $state('all');

  // Derived State
  let filteredTodos = $derived(
    activeFilter === 'all' 
      ? todos 
      : todos.filter(t => t.status === activeFilter)
  );

  let userTasks = $derived(filteredTodos.filter(t => !t.is_default_task));
  let defaultTasks = $derived(filteredTodos.filter(t => t.is_default_task));

  // Stats
  let stats = $derived({
    all: todos.length,
    pending: todos.filter(t => t.status === 'pending').length,
    inProgress: todos.filter(t => t.status === 'in-progress').length,
    done: todos.filter(t => t.status === 'done').length
  });

  onMount(async () => {
    if (userId) {
      await loadData();
    }
  });

  async function loadData() {
    loading = true;
    try {
      const [usersList, todosList] = await Promise.all([
        api.listUsers(),
        api.listUserTodos(userId)
      ]);
      user = usersList.find(u => u.id === userId);
      todos = todosList;
    } catch (e) {
      console.error("Failed to load data", e);
    } finally {
      loading = false;
    }
  }

  async function handleAddTask() {
    if (!newTaskText.trim()) return;
    
    isSubmitting = true;
    try {
      await api.createUserTodo(userId, newTaskText, hideFromUser);
      newTaskText = "";
      hideFromUser = false;
      const updatedTodos = await api.listUserTodos(userId);
      todos = updatedTodos;
    } catch (e) {
      console.error("Failed to create task", e);
    } finally {
      isSubmitting = false;
    }
  }

  async function toggleVisibility(todo: Todo) {
    try {
      const newHiddenStatus = !todo.hidden_from_user;
      await api.updateUserTodo(userId, todo.id, { hidden_from_user: newHiddenStatus });
      todos = todos.map(t => t.id === todo.id ? { ...t, hidden_from_user: newHiddenStatus } : t);
    } catch (e) {
      console.error("Failed to toggle visibility", e);
    }
  }

  async function cycleStatus(todo: Todo) {
    let nextStatus: TodoStatus;
    if (todo.status === 'pending') nextStatus = 'in-progress';
    else if (todo.status === 'in-progress') nextStatus = 'done';
    else nextStatus = 'pending';

    try {
      await api.updateUserTodo(userId, todo.id, { status: nextStatus });
      todos = todos.map(t => t.id === todo.id ? { ...t, status: nextStatus } : t);
    } catch (e) {
      console.error("Failed to update status", e);
      // Refresh to ensure sync
      const updatedTodos = await api.listUserTodos(userId);
      todos = updatedTodos;
    }
  }

  // Edit State
  let editingId: string | null = $state(null);
  let editText = $state("");

  async function deleteTodo(todoId: string) {
    if (!confirm('Are you sure you want to delete this task?')) return;
    
    try {
      await api.deleteUserTodo(userId, todoId);
      todos = todos.filter(t => t.id !== todoId);
    } catch (e) {
      console.error("Failed to delete todo", e);
    }
  }

  function startEditing(todo: Todo) {
    editingId = todo.id;
    editText = todo.text;
  }

  function cancelEditing() {
    editingId = null;
    editText = "";
  }

  async function saveEdit(todo: Todo) {
    if (!editText.trim()) return;
    try {
      await api.updateUserTodo(userId, todo.id, { text: editText });
      todos = todos.map(t => t.id === todo.id ? { ...t, text: editText } : t);
      editingId = null;
    } catch (e) {
      console.error("Failed to update task text", e);
    }
  }

  function handleKeydown(event: KeyboardEvent, todo: Todo) {
    if (event.key === "Enter") {
      event.preventDefault();
      saveEdit(todo);
    } else if (event.key === "Escape") {
      cancelEditing();
    }
  }

  function goBack() {
    goto('/admin');
  }

  const filterTabs = [
    { id: 'all', label: 'ALL' },
    { id: 'pending', label: 'PENDING' },
    { id: 'in-progress', label: 'IN PROGRESS' },
    { id: 'done', label: 'DONE' }
  ] as const;
</script>

<div class="max-w-4xl mx-auto p-4 md:p-8">
  <button 
    onclick={goBack}
    class="mb-6 flex items-center text-sm font-bold text-indigo-600 hover:text-indigo-800 transition-colors"
  >
    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M19 12H5M12 19l-7-7 7-7"/>
    </svg>
    Back to Users
  </button>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
    </div>
  {:else if user}
    <!-- User Profile & Stats Card -->
    <div class="glass-card rounded-2xl p-6 mb-6 flex flex-col md:flex-row items-center justify-between gap-6">
      <div class="flex items-center gap-4">
        {#if user.avatar_url}
          <img src={user.avatar_url} alt={user.name} class="w-16 h-16 rounded-full border-2 border-indigo-100" />
        {:else}
          <div class="w-16 h-16 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 text-xl font-bold">
            {user.name.charAt(0).toUpperCase()}
          </div>
        {/if}
        <div>
          <h1 class="text-xl font-bold text-slate-800">{user.name}</h1>
          <p class="text-sm text-slate-500">{user.email}</p>
        </div>
      </div>

      <div class="flex flex-wrap gap-2">
        {#each filterTabs as tab}
          <button
            onclick={() => activeFilter = tab.id}
            class="px-4 py-2 rounded-full text-xs font-bold transition-all uppercase tracking-wide
              {activeFilter === tab.id 
                ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-200' 
                : 'bg-white border border-slate-200 text-slate-500 hover:border-indigo-300 hover:text-indigo-600'}"
          >
            {tab.label} ({stats[tab.id === 'in-progress' ? 'inProgress' : tab.id]})
          </button>
        {/each}
      </div>
    </div>

    <!-- Add Task Card -->
    <div class="glass-card rounded-2xl p-6 mb-8">
      <h2 class="text-lg font-bold text-slate-800 mb-4">Add Task for {user.name}</h2>
      <form onsubmit={(e) => { e.preventDefault(); handleAddTask(); }} class="flex flex-col md:flex-row gap-4 items-center">
        <input
          bind:value={newTaskText}
          placeholder="Enter task text..."
          class="flex-1 w-full px-4 py-3 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 bg-slate-50/50"
        />
        
        <label class="flex items-center gap-2 cursor-pointer select-none">
          <div class="relative">
            <input type="checkbox" bind:checked={hideFromUser} class="sr-only peer" />
            <div class="w-11 h-6 bg-slate-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600"></div>
          </div>
          <span class="text-sm font-medium text-slate-600">Hide from user</span>
        </label>

        <button
          type="submit"
          disabled={!newTaskText.trim() || isSubmitting}
          class="w-full md:w-auto px-8 py-3 bg-indigo-600 text-white font-bold rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-indigo-200"
        >
          {isSubmitting ? 'Adding...' : 'Add'}
        </button>
      </form>
    </div>

    <!-- Admin Added Tasks -->
    {#if todos.filter(t => !t.is_default_task && t.created_by_user_id !== user?.id).length > 0}
      <div class="mb-4">
        <h3 class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-2 px-2">Admin Added</h3>
        <div class="space-y-3">
          {#each todos.filter(t => !t.is_default_task && t.created_by_user_id !== user?.id && (activeFilter === 'all' || t.status === activeFilter)) as todo (todo.id)}
            <div class="bg-white p-4 rounded-xl border border-slate-100 shadow-sm flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <StatusIndicator 
                  status={todo.status} 
                  clickable={true} 
                  on:click={() => cycleStatus(todo)}
                />
                <div class="flex-1 min-w-0 flex items-center gap-2">
                  {#if editingId === todo.id}
                     <div class="flex-1 flex gap-2">
                      <input
                        bind:value={editText}
                        onkeydown={(e) => handleKeydown(e, todo)}
                        onclick={(e) => e.stopPropagation()} 
                        class="flex-1 px-2 py-1 text-sm border border-indigo-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                        autofocus
                      />
                      <button 
                        onclick={(e) => { e.stopPropagation(); saveEdit(todo); }}
                        class="p-1 text-emerald-600 hover:bg-emerald-50 rounded"
                        track="Save"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                      </button>
                      <button 
                        onclick={(e) => { e.stopPropagation(); cancelEditing(); }}
                        class="p-1 text-slate-400 hover:bg-slate-100 rounded"
                        track="Cancel"
                      >
                         <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                      </button>
                    </div>
                  {:else}
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <span 
                      class="font-medium text-slate-700 truncate cursor-text hover:text-indigo-600 transition-colors"
                      ondblclick={() => startEditing(todo)}
                      title="Double click to edit"
                    >
                      {todo.text}
                    </span>
                    <!-- Labels -->
                    {#if todo.hidden_from_user}
                      <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-slate-500 bg-slate-100 px-2 py-0.5 rounded-full border border-slate-200 uppercase gap-1 shadow-sm whitespace-nowrap">
                        Hidden from User
                      </span>
                    {:else}
                      <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-green-600 bg-green-50 px-2 py-0.5 rounded-full border border-green-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                        Shared
                      </span>
                    {/if}
                    <!-- Status Badges -->
                    {#if todo.status === 'in-progress'}
                      <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full border border-amber-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                        In Progress
                      </span>
                    {/if}
                    {#if todo.status === 'done'}
                      <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                        Done
                      </span>
                    {/if}
                  {/if}
                </div>
              </div>

              <div class="flex items-center gap-2 shrink-0">
                <button 
                  onclick={() => toggleVisibility(todo)}
                  class="p-2 text-slate-400 hover:bg-slate-100 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                  title={todo.hidden_from_user ? "Hidden from user (Click to show)" : "Visible to user (Click to hide)"}
                  aria-label={todo.hidden_from_user ? "Show to user" : "Hide from user"}
                >
                  {#if todo.hidden_from_user}
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-indigo-400/50 hover:text-indigo-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                  {/if}
                </button>
                <button
                  onclick={() => startEditing(todo)}
                  aria-label="Edit Task"
                  class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                  </svg>
                </button>
                <button
                  onclick={() => deleteTodo(todo.id)}
                  aria-label="Delete Task"
                  class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"></path>
                  </svg>
                </button>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Customer Added Tasks -->
    {#if todos.filter(t => !t.is_default_task && t.created_by_user_id === user?.id).length > 0}
      <div class="mb-4">
        <h3 class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-2 px-2">Customer Added</h3>
        <div class="space-y-3">
          {#each todos.filter(t => !t.is_default_task && t.created_by_user_id === user?.id && (activeFilter === 'all' || t.status === activeFilter)) as todo (todo.id)}
            <div class="bg-white p-4 rounded-xl border border-slate-100 shadow-sm flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <StatusIndicator 
                  status={todo.status} 
                  clickable={true} 
                  on:click={() => cycleStatus(todo)}
                />
                <div class="flex-1 min-w-0 flex items-center gap-2">
                  <span class="font-medium text-slate-700 truncate">
                    {todo.text}
                  </span>
                  <!-- Status Badges -->
                  {#if todo.status === 'in-progress'}
                    <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full border border-amber-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                      In Progress
                    </span>
                  {/if}
                  {#if todo.status === 'done'}
                    <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                      Done
                    </span>
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Default Tasks -->
    {#if defaultTasks.length > 0}
      <div>
        <h3 class="text-xs font-bold text-slate-500 uppercase tracking-wider mb-4 px-2">Default</h3>
        <div class="space-y-3">
          {#each defaultTasks as todo (todo.id)}
            <div class="bg-white/50 p-4 rounded-xl border border-slate-100 flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <StatusIndicator 
                  status={todo.status} 
                  clickable={true} 
                  on:click={() => cycleStatus(todo)}
                />
                <span class="font-medium text-slate-600 flex-1 flex items-center gap-2">
                  <span class="truncate">{todo.text}</span>
                  <!-- Labels -->
                  <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-purple-600 bg-purple-50 px-2 py-0.5 rounded-full border border-purple-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                    Default Task
                  </span>
                  {#if todo.status === 'in-progress'}
                    <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full border border-amber-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                      In Progress
                    </span>
                  {/if}
                  {#if todo.status === 'done'}
                    <span class="inline-flex items-center text-[10px] font-bold tracking-wider text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-100 uppercase gap-1 shadow-sm whitespace-nowrap">
                      Done
                    </span>
                  {/if}
                </span>
              </div>
              
              <div class="flex items-center gap-4">
                <span class="text-xs text-slate-400 font-medium">
                  {new Date(todo.created).toLocaleDateString()}
                </span>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

  {:else}
    <div class="text-center py-20 text-slate-500">
      User not found
    </div>
  {/if}
</div>
