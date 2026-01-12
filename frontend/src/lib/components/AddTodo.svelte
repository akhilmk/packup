<script lang="ts">
  import { api, type User } from "../api";
  import { createEventDispatcher } from "svelte";

  export let user: User | null = null;
  const dispatch = createEventDispatcher();
  let text = "";
  let hideFromAdmin = false;
  const LIMIT = 200;

  async function addTodo() {
    if (!text.trim() || text.length > LIMIT) return;
    try {
      await api.createTodo(text, !hideFromAdmin);
      text = "";
      hideFromAdmin = false; // Reset to default (shared)
      dispatch("added");
    } catch (e) {
      console.error(e);
      alert("Failed to add todo");
    }
  }

  $: charsLeft = LIMIT - text.length;
</script>

<div class="glass-card rounded-2xl p-6 mb-8 transform transition-all duration-300 hover:shadow-2xl">
  <form on:submit|preventDefault={addTodo} class="space-y-4">
    <div class="relative group">
      <input
        type="text"
        bind:value={text}
        placeholder="What's on your mind?"
        class="input-modern pr-12 {charsLeft < 0 ? 'border-red-400 focus:ring-red-500/10' : ''}"
      />
      
      <div class="absolute right-4 top-1/2 -translate-y-1/2 flex items-center space-x-2">
        {#if text.length > 0}
          <div 
            class="text-[10px] font-bold px-1.5 py-0.5 rounded-full {charsLeft < 0 ? 'bg-red-100 text-red-600' : 'bg-indigo-50 text-indigo-500'}"
          >
            {charsLeft}
          </div>
        {/if}
      </div>
    </div>

    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-4">
        {#if user && user.role !== 'admin'}
          <label class="flex items-center gap-2 cursor-pointer select-none group relative">
            <input type="checkbox" bind:checked={hideFromAdmin} class="sr-only peer" />
            <div class="relative w-11 h-6 bg-slate-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-indigo-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600"></div>
            <span class="text-xs font-medium text-slate-500 group-hover:text-slate-700 transition-colors">Hide from Admin</span>
          </label>
        {/if}
      </div>
      
      <button
        type="submit"
        disabled={!text.trim() || charsLeft < 0}
        class="btn-primary flex items-center space-x-2"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        <span>Add Task</span>
      </button>
    </div>
  </form>
</div>
