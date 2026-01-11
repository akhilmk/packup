<script lang="ts">
  import { onMount } from "svelte";
  import { api, type Todo } from "../api";
  import TodoItem from "./TodoItem.svelte";
  import AddTodo from "./AddTodo.svelte";

  let todos: Todo[] = [];
  let loading = true;

  async function fetchTodos() {
    try {
      todos = await api.listTodos();
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  onMount(fetchTodos);
</script>

<div class="max-w-2xl mx-auto p-4 md:p-8">
  <header class="mb-10 text-center">
    <h1 class="text-4xl font-black text-slate-800 tracking-tight mb-2">
      Daily Focus
    </h1>
    <p class="text-slate-500 font-medium">
      {new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}
    </p>
    
    <div class="mt-6 flex justify-center space-x-4">
      <div class="px-4 py-1.5 bg-indigo-50 rounded-full border border-indigo-100/50">
        <span class="text-xs font-bold text-indigo-600 uppercase tracking-wider">
          {todos.filter(t => t.status === 'pending').length} Pending
        </span>
      </div>
      <div class="px-4 py-1.5 bg-amber-50 rounded-full border border-amber-100/50">
        <span class="text-xs font-bold text-amber-600 uppercase tracking-wider">
          {todos.filter(t => t.status === 'in-progress').length} In Progress
        </span>
      </div>
      <div class="px-4 py-1.5 bg-emerald-50 rounded-full border border-emerald-100/50">
        <span class="text-xs font-bold text-emerald-600 uppercase tracking-wider">
          {todos.filter(t => t.status === 'done').length} Done
        </span>
      </div>
    </div>
  </header>
  
  <AddTodo on:added={fetchTodos} />

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
    <div class="glass-card rounded-3xl overflow-hidden border-border divide-y divide-gray-50">
      {#each todos as todo (todo.id)}
        <TodoItem {todo} on:update={fetchTodos} />
      {/each}
    </div>
    
    <p class="text-center text-slate-400 text-sm font-medium mt-8">
      Double-click a task to edit
    </p>
  {/if}
</div>

