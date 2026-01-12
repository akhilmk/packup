<script lang="ts">
  import { api, type Todo, type TodoStatus } from "../api";
  import { createEventDispatcher } from "svelte";
  import { fade, slide } from "svelte/transition";

  export let todo: Todo;
  const dispatch = createEventDispatcher();

  let isEditing = false;
  let editText = todo.text;
  const LIMIT = 200;

  async function cycleStatus() {
    let nextStatus: TodoStatus;
    if (todo.status === 'pending') nextStatus = 'in-progress';
    else if (todo.status === 'in-progress') nextStatus = 'done';
    else nextStatus = 'pending';

    try {
      await api.updateTodo(todo.id, { status: nextStatus });
      dispatch("update");
    } catch (e) {
      console.error(e);
    }
  }

  async function deleteTodo() {
    try {
      if (todo.is_default_task) {
        // Should not be reachable due to UI restrictions, but for safety
        return;
      }
      await api.deleteTodo(todo.id);
      dispatch("delete", todo.id);
    } catch (e) {
      console.error("Failed to delete todo", e);
    }
  }

  async function toggleShare() {
    try {
      const updated = await api.updateTodo(todo.id, {
        shared_with_admin: !todo.shared_with_admin
      });
      todo = updated;
    } catch (e) {
      console.error("Failed to toggle share status", e);
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      saveEdit();
    } else if (event.key === "Escape") {
      isEditing = false;
      editText = todo.text;
    }
  }

  async function saveEdit() {
    if (editText.length > LIMIT || !editText.trim()) return;
    try {
      await api.updateTodo(todo.id, { text: editText });
      isEditing = false;
      dispatch("update");
    } catch (e) {
      console.error(e);
    }
  }

  function cancelEdit() {
    isEditing = false;
    editText = todo.text;
  }
</script>

<div 
  transition:slide={{ duration: 300 }}
  class="flex items-center gap-4 p-4 bg-white border-b border-gray-50 last:border-0 hover:bg-indigo-50/30 transition-all group overflow-hidden"
>
  <!-- Status Cycle Button -->
  <button 
    on:click={cycleStatus}
    class="relative flex items-center justify-center w-6 h-6 rounded-full border-2 transition-all duration-200 
           {todo.status === 'done' ? 'bg-emerald-500 border-emerald-500' : 
            todo.status === 'in-progress' ? 'bg-amber-100 border-amber-400' : 
            'bg-white border-slate-200 hover:border-indigo-400'}"
    title="Click to cycle status: Pending -> In Progress -> Done"
  >
    {#if todo.status === 'done'}
      <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="20 6 9 17 4 12"></polyline>
      </svg>
    {:else if todo.status === 'in-progress'}
      <div class="w-2.5 h-2.5 bg-amber-500 rounded-full animate-pulse"></div>
    {/if}
  </button>

  {#if isEditing}
    <form 
      in:fade
      on:submit|preventDefault={saveEdit} 
      class="flex-1 flex gap-2"
    >
      <!-- svelte-ignore a11y-autofocus -->
      <input
        bind:value={editText}
        autofocus
        aria-label="Edit todo text"
        class="flex-1 px-3 py-1.5 bg-white border border-indigo-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
      />
      <button 
        type="submit" 
        aria-label="Save changes"
        class="p-2 text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      </button>
      <button 
        type="button" 
        on:click={cancelEdit} 
        aria-label="Cancel editing"
        class="p-2 text-slate-400 hover:bg-slate-100 rounded-lg transition-colors"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
    </form>
  {:else}
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <span
      class="flex-1 font-medium selection:bg-indigo-100 transition-all duration-300 
             {todo.status === 'done' ? 'line-through text-slate-300 opacity-60' : 'text-slate-700'}
             {todo.status === 'in-progress' ? 'text-indigo-600' : ''}"
      on:dblclick={() => { if (!todo.is_default_task) isEditing = true; }}
    >
      {todo.text}
      {#if todo.is_default_task}
        <span class="inline-flex items-center ml-2 text-[10px] font-bold tracking-wider text-purple-600 bg-purple-50 px-2 py-0.5 rounded-full border border-purple-100 uppercase align-middle transform -translate-y-0.5 gap-1 shadow-sm">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2L2 7l10 5 10-5-10-5z"></path>
            <path d="M2 17l10 5 10-5"></path>
            <path d="M2 12l10 5 10-5"></path>
          </svg>
          Default Task
        </span>
      {/if}
      {#if !todo.is_default_task && todo.shared_with_admin}
        <span class="inline-flex items-center ml-2 text-[10px] font-bold tracking-wider text-green-600 bg-green-50 px-2 py-0.5 rounded-full border border-green-100 uppercase align-middle transform -translate-y-0.5 gap-1 shadow-sm">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
            <circle cx="12" cy="12" r="3"></circle>
          </svg>
          Shared
        </span>
      {/if}
      {#if todo.status === 'in-progress'}
        <span class="inline-flex items-center ml-2 text-[10px] font-bold tracking-wider text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full border border-amber-100 uppercase align-middle transform -translate-y-0.5 gap-1 shadow-sm">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-3 h-3 animate-spin" style="animation-duration: 3s;" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M5 22h14"></path>
            <path d="M5 2h14"></path>
            <path d="M17 22v-4.172a2 2 0 0 0-.586-1.414L12 12l-4.414 4.414A2 2 0 0 0 7 17.828V22"></path>
            <path d="M7 2v4.172a2 2 0 0 0 .586 1.414L12 12l4.414-4.414A2 2 0 0 0 17 6.172V2"></path>
          </svg>
          In Progress
        </span>
      {/if}
    </span>
    
    <div class="opacity-0 group-hover:opacity-100 transition-all duration-200 flex items-center space-x-1">
      {#if !todo.is_default_task}
      <!-- Share Toggle (only for personal tasks) -->
      <button
        on:click={toggleShare}
        class="p-2 transition-all rounded-lg {todo.shared_with_admin ? 'text-green-600 bg-green-50 hover:bg-green-100' : 'text-slate-400 hover:text-green-600 hover:bg-green-50'}"
        title={todo.shared_with_admin ? "Unshare with Admin" : "Share with Admin"}
        aria-label="Toggle share with admin"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          {#if todo.shared_with_admin}
            <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
            <line x1="1" y1="1" x2="23" y2="23"></line>
          {:else}
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
            <circle cx="12" cy="12" r="3"></circle>
          {/if}
        </svg>
      </button>

      <button 
        on:click={() => (isEditing = true)} 
        class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-100/50 rounded-lg transition-all"
        title="Edit task"
        aria-label="Edit task"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
        </svg>
      </button>
      {/if}
      <button 
        on:click={deleteTodo} 
        disabled={todo.is_default_task}
        class="p-2 transition-all rounded-lg
               {todo.is_default_task 
                 ? 'text-slate-300 cursor-not-allowed' 
                 : 'text-slate-400 hover:text-red-500 hover:bg-red-50'}"
        title={todo.is_default_task ? "Default tasks cannot be deleted" : "Delete task"}
        aria-label="Delete task"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="3 6 5 6 21 6"></polyline>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          <line x1="10" y1="11" x2="10" y2="17"></line>
          <line x1="14" y1="11" x2="14" y2="17"></line>
        </svg>
      </button>
    </div>
  {/if}
</div>

