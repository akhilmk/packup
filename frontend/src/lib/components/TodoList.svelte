<script lang="ts">
  import { onMount } from "svelte";
  import { api, type Todo, type User } from "../api";
  import TodoItem from "./TodoItem.svelte";
  import AddTodo from "./AddTodo.svelte";
  import { flip } from "svelte/animate";

  export let user: User | null = null;

  let todos: Todo[] = [];
  let loading = true;
  let filter: 'all' | 'pending' | 'in-progress' | 'done' = 'all';

  // Derived stores for filtering
  $: displayTodos = todos;
  $: filteredTodos = filter === 'all' ? displayTodos : displayTodos.filter(t => t.status === filter);

  // Derived stores for sectioned view (for regular users)
  $: defaultTodos = filteredTodos.filter(t => t.is_default_task);
  $: adminAssignedTodos = filteredTodos.filter(t => !t.is_default_task && t.created_by_user_id !== user?.id);
  $: personalTodos = filteredTodos.filter(t => !t.is_default_task && t.created_by_user_id === user?.id);

  async function fetchTodos() {
    try {
      const options = {};
      todos = await api.listTodos(options);
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  let draggedItemIndex: number | null = null;
  // Track which section the drag started in to restrict movement if needed, 
  // or just disable DnD for sectioned view for simplicity unless requested.
  // Requirement: "remember my todos of admin not required any change" -> Admin DnD should work.
  // For User: "have sections ... make it looks good". DnD might be confusing across sections.
  // Let's enable DnD only for Admin "My Todos" (flat view) for now to be safe and avoid bugs.
  // Or enable it within Personal section if robust. TodoItem triggers events.
  // Given the complexity of reordering a partitioned list backed by a single array, 
  // I will restrict Drag and Drop to the Admin (Flat) view for now.

  function handleDragStart(event: DragEvent, index: number) {
    if (filter !== 'all' || user?.role !== 'admin') return; 
    draggedItemIndex = index;
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.dropEffect = 'move';
    }
  }

  function handleDragOver(event: DragEvent, index: number) {
    event.preventDefault();
    if (draggedItemIndex === null || draggedItemIndex === index) return;
    
    // Only allow for Admin flat view
    if (user?.role !== 'admin') return;

    const newTodos = [...todos];
    const [item] = newTodos.splice(draggedItemIndex, 1);
    newTodos.splice(index, 0, item);
    
    todos = newTodos;
    draggedItemIndex = index;
  }

  async function handleDragEnd() {
    draggedItemIndex = null;
    if (user?.role !== 'admin') return;

    try {
      const ids = todos.map(t => t.id);
      await api.reorderTodos(ids);
    } catch (e) {
      console.error("Failed to save order", e);
    }
  }

  onMount(fetchTodos);
</script>

<div class="max-w-2xl mx-auto p-4 md:p-8">
  <header class="mb-10 text-center">
    <!-- Packup Icon -->
    <div class="mb-4 flex justify-center">
      <div class="w-16 h-16 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 2C7 2 3 6 3 11c0 6 9 13 9 13s9-7 9-13c0-5-4-9-9-9z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8l1.5 1.5 3-3" />
          <line x1="13" y1="8" x2="17" y2="8" stroke-width="2" stroke-linecap="round" />
          <rect x="7" y="12" width="3" height="3" rx="0.5" stroke-width="2" />
          <line x1="12" y1="13.5" x2="17" y2="13.5" stroke-width="2" stroke-linecap="round" />
          <rect x="7" y="16" width="3" height="3" rx="0.5" stroke-width="2" />
          <line x1="12" y1="17.5" x2="17" y2="17.5" stroke-width="2" stroke-linecap="round" />
        </svg>
      </div>
    </div>
    
    <h1 class="text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 tracking-tight mb-2">
      PackUp
    </h1>
    <p class="text-slate-500 font-medium">
      {new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}
    </p>
    
    <div class="mt-6 flex justify-center flex-wrap gap-4">
      <button 
        class="px-4 py-1.5 rounded-full border transition-all duration-200 {filter === 'all' ? 'bg-indigo-600 text-white border-indigo-600 shadow-md transform scale-105' : 'bg-white text-slate-500 border-slate-200 hover:border-indigo-300 hover:text-indigo-600'}"
        on:click={() => filter = 'all'}
      >
        <span class="text-xs font-bold uppercase tracking-wider">
          All ({displayTodos.length})
        </span>
      </button>

      <button 
        class="px-4 py-1.5 rounded-full border transition-all duration-200 {filter === 'pending' ? 'bg-indigo-50 border-indigo-200 ring-2 ring-indigo-100 shadow-sm' : 'bg-white border-slate-200 hover:bg-slate-50'}"
        on:click={() => filter = 'pending'}
      >
        <span class="text-xs font-bold {filter === 'pending' ? 'text-indigo-700' : 'text-slate-500'} uppercase tracking-wider">
          Pending ({displayTodos.filter(t => t.status === 'pending').length})
        </span>
      </button>

      <button 
        class="px-4 py-1.5 rounded-full border transition-all duration-200 {filter === 'in-progress' ? 'bg-amber-50 border-amber-200 ring-2 ring-amber-100 shadow-sm' : 'bg-white border-slate-200 hover:bg-slate-50'}"
        on:click={() => filter = 'in-progress'}
      >
        <span class="text-xs font-bold {filter === 'in-progress' ? 'text-amber-700' : 'text-slate-500'} uppercase tracking-wider">
          In Progress ({displayTodos.filter(t => t.status === 'in-progress').length})
        </span>
      </button>

      <button 
        class="px-4 py-1.5 rounded-full border transition-all duration-200 {filter === 'done' ? 'bg-emerald-50 border-emerald-200 ring-2 ring-emerald-100 shadow-sm' : 'bg-white border-slate-200 hover:bg-slate-50'}"
        on:click={() => filter = 'done'}
      >
        <span class="text-xs font-bold {filter === 'done' ? 'text-emerald-700' : 'text-slate-500'} uppercase tracking-wider">
          Done ({displayTodos.filter(t => t.status === 'done').length})
        </span>
      </button>
    </div>
  </header>
  
  <AddTodo {user} on:added={fetchTodos} />

  {#if loading}
    <div class="flex flex-col items-center justify-center py-20 space-y-4">
      <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
      <p class="text-slate-400 font-medium animate-pulse">Syncing tasks...</p>
    </div>
  {:else if todos.length === 0}
    <div class="glass-card rounded-2xl p-12 text-center">
      <div class="w-20 h-20 bg-indigo-50 text-indigo-500 rounded-2xl flex items-center justify-center mx-auto mb-6">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-10 h-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
      </div>
      <h3 class="text-xl font-bold text-slate-800 mb-2">Clarity starts here</h3>
      <p class="text-slate-500 max-w-sm mx-auto">Your task list is empty. Add your first goal to get momentum started.</p>
    </div>
  {:else}
    {#if user?.role === 'admin'}
      <!-- Admin View (Flat List) -->
      <div class="glass-card rounded-3xl overflow-hidden border-border divide-y divide-gray-50">
        {#each filteredTodos as todo, index (todo.id)}
          <div
            role="listitem"
            draggable={filter === 'all'}
            animate:flip={{ duration: 300 }}
            on:dragstart={(e) => handleDragStart(e, index)}
            on:dragover={(e) => handleDragOver(e, index)}
            on:dragend={handleDragEnd}
            class="cursor-grab active:cursor-grabbing {draggedItemIndex === index ? 'opacity-50 bg-indigo-50/50' : ''} {filter !== 'all' ? 'cursor-default active:cursor-default' : ''}"
          >
            <TodoItem {todo} {user} on:update={fetchTodos} on:delete={fetchTodos} />
          </div>
        {/each}
      </div>
    {:else}
      <!-- User View (Sectioned List) -->
      <div class="space-y-6">
        <!-- Default Tasks -->
        {#if defaultTodos.length > 0}
          <div>
            <h3 class="text-sm font-bold text-slate-500 uppercase tracking-widest px-4 mb-2">Default</h3>
            <div class="glass-card rounded-2xl overflow-hidden border-border divide-y divide-gray-50">
              {#each defaultTodos as todo (todo.id)}
                <TodoItem {todo} {user} showSourceLabels={false} on:update={fetchTodos} on:delete={fetchTodos} />
              {/each}
            </div>
          </div>
        {/if}

        <!-- Admin Assigned Tasks -->
        {#if adminAssignedTodos.length > 0}
          <div>
            <h3 class="text-sm font-bold text-slate-500 uppercase tracking-widest px-4 mb-2">Admin Added</h3>
            <div class="glass-card rounded-2xl overflow-hidden border-border divide-y divide-gray-50">
              {#each adminAssignedTodos as todo (todo.id)}
                <TodoItem {todo} {user} showSourceLabels={false} on:update={fetchTodos} on:delete={fetchTodos} />
              {/each}
            </div>
          </div>
        {/if}

        <!-- Personal Tasks -->
        {#if personalTodos.length > 0}
          <div>
            <h3 class="text-sm font-bold text-slate-500 uppercase tracking-widest px-4 mb-2">Customer Added</h3>
            <div class="glass-card rounded-2xl overflow-hidden border-border divide-y divide-gray-50">
              {#each personalTodos as todo (todo.id)}
                <TodoItem {todo} {user} showSourceLabels={false} on:update={fetchTodos} on:delete={fetchTodos} />
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
    
    <p class="text-center text-slate-400 text-sm font-medium mt-8">
      Double-click a task to edit
    </p>
  {/if}
</div>
