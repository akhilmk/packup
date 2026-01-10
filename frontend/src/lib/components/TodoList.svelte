<script lang="ts">
  import { onMount } from "svelte";
  import { request } from "../client";
  import type { Todo } from "../../gen/proto/todo_pb";
  import TodoItem from "./TodoItem.svelte";
  import AddTodo from "./AddTodo.svelte";

  let todos: Todo[] = [];
  let loading = true;

  async function fetchTodos() {
    try {
      const res = await request.listTodos({});
      todos = res.todos;
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  onMount(fetchTodos);
</script>

<div class="max-w-2xl mx-auto p-6">
  <h1 class="text-3xl font-bold text-gray-900 mb-8 text-center">Tasks</h1>
  
  <AddTodo on:added={fetchTodos} />

  {#if loading}
    <div class="text-center text-gray-400 py-8">Loading...</div>
  {:else if todos.length === 0}
    <div class="text-center text-gray-400 py-8">No tasks yet. Add one above!</div>
  {:else}
    <div class="bg-white rounded-xl shadow-lg border border-gray-100 overflow-hidden">
      {#each todos as todo (todo.id)}
        <TodoItem {todo} on:update={fetchTodos} />
      {/each}
    </div>
  {/if}
</div>
