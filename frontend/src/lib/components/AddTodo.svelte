<script lang="ts">
  import { api } from "../api";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  let text = "";
  const LIMIT = 200;

  async function addTodo() {
    if (!text.trim() || text.length > LIMIT) return;
    try {
      await api.createTodo(text);
      text = "";
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
      <p class="text-sm text-slate-400 font-medium">
        {#if text.length === 0}
          Plan your day ahead
        {:else if charsLeft < 0}
          <span class="text-red-500">Message is too long</span>
        {:else}
          <span class="text-indigo-500">Ready to save</span>
        {/if}
      </p>
      
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

