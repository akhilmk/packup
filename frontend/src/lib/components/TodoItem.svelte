<script lang="ts">
  import { api, type Todo } from "../api";
  import { createEventDispatcher } from "svelte";
  import { fade, slide } from "svelte/transition";

  export let todo: Todo;
  const dispatch = createEventDispatcher();

  let isEditing = false;
  let editText = todo.text;
  const LIMIT = 200;

  async function toggleComplete() {
    try {
      await api.updateTodo(todo.id, { completed: !todo.completed });
      dispatch("update");
    } catch (e) {
      console.error(e);
    }
  }

  async function deleteTodo() {
    try {
      await api.deleteTodo(todo.id);
      dispatch("update");
    } catch (e) {
      console.error(e);
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
  <!-- Custom Checkbox -->
  <button 
    on:click={toggleComplete}
    class="relative flex items-center justify-center w-6 h-6 rounded-full border-2 transition-all duration-200 
           {todo.completed ? 'bg-emerald-500 border-emerald-500' : 'bg-white border-slate-200 hover:border-indigo-400'}"
  >
    {#if todo.completed}
      <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="20 6 9 17 4 12"></polyline>
      </svg>
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
      class="flex-1 text-slate-700 font-medium selection:bg-indigo-100 transition-all duration-300 {todo.completed ? 'line-through text-slate-300 opacity-60' : ''}"
      on:dblclick={() => (isEditing = true)}
    >
      {todo.text}
    </span>
    
    <div class="opacity-0 group-hover:opacity-100 transition-all duration-200 flex items-center space-x-1">
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
      <button 
        on:click={deleteTodo} 
        class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all"
        title="Delete task"
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

