<script lang="ts">
  import { request } from "../client";
  import type { Todo } from "../../gen/proto/todo_pb";
  import { createEventDispatcher } from "svelte";

  export let todo: Todo;
  const dispatch = createEventDispatcher();

  let isEditing = false;
  let editText = todo.text;
  const LIMIT = 200;

  async function toggleComplete() {
    try {
      await request.updateTodo({ ...todo, completed: !todo.completed });
      dispatch("update");
    } catch (e) {
      console.error(e);
    }
  }

  async function deleteTodo() {
    try {
      await request.deleteTodo({ id: todo.id });
      dispatch("update");
    } catch (e) {
      console.error(e);
    }
  }

  async function saveEdit() {
    if (editText.length > LIMIT || !editText.trim()) return;
    try {
      await request.updateTodo({ ...todo, text: editText });
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

<div class="flex items-center gap-3 p-3 bg-white border-b border-gray-100 last:border-0 hover:bg-gray-50 transition-colors group">
  <input
    type="checkbox"
    checked={todo.completed}
    on:change={toggleComplete}
    class="w-5 h-5 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
  />

  {#if isEditing}
    <form on:submit|preventDefault={saveEdit} class="flex-1 flex gap-2">
      <input
        bind:value={editText}
        class="flex-1 px-2 py-1 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      <button type="submit" class="text-green-600 hover:text-green-700">Save</button>
      <button type="button" on:click={cancelEdit} class="text-gray-500 hover:text-gray-700">Cancel</button>
    </form>
  {:else}
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <span
      class="flex-1 text-gray-800 {todo.completed ? 'line-through text-gray-400' : ''}"
      on:dblclick={() => (isEditing = true)}
    >
      {todo.text}
    </span>
    <div class="opacity-0 group-hover:opacity-100 transition-opacity flex gap-2">
      <button on:click={() => (isEditing = true)} class="text-gray-400 hover:text-blue-500">
        Edit
      </button>
      <button on:click={deleteTodo} class="text-gray-400 hover:text-red-500">
        Delete
      </button>
    </div>
  {/if}
</div>
