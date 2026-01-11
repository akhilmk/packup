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

<div class="mb-6 p-4 bg-white rounded shadow-sm border border-gray-100">
  <form on:submit|preventDefault={addTodo} class="flex gap-2">
    <div class="flex-1">
      <input
        type="text"
        bind:value={text}
        placeholder="What needs to be done?"
        class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors {charsLeft < 0 ? 'border-red-500' : 'border-gray-200'}"
      />
      <div class="text-xs text-right mt-1 {charsLeft < 0 ? 'text-red-500' : 'text-gray-400'}">
        {charsLeft} characters left
      </div>
    </div>
    <button
      type="submit"
      disabled={!text.trim() || charsLeft < 0}
      class="px-6 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors h-fit"
    >
      Add
    </button>
  </form>
</div>
