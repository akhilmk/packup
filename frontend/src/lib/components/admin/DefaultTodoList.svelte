<script lang="ts">
  import { onMount } from "svelte";
  import { api, type Todo } from "$lib/api";
  import ConfirmModal from "$lib/components/ConfirmModal.svelte";

  function focus(node: HTMLInputElement) {
    node.focus();
  }

  let adminTodos: Todo[] = $state([]);
  let loading = $state(true);
  let newAdminTodoText = $state("");
  let editingTodoId: string | null = $state(null);
  let editingTodoText = $state("");

  // Confirmation modal state
  let showConfirmModal = $state(false);
  let confirmModalConfig = $state({
    title: '',
    message: '',
    onConfirm: () => {},
    danger: false
  });

  onMount(async () => {
    await loadAdminTodos();
  });

  async function loadAdminTodos() {
    loading = true;
    try {
      adminTodos = await api.listDefaultTasks();
    } catch (e) {
      console.error("Failed to load default tasks", e);
    } finally {
      loading = false;
    }
  }

  async function handleCreateAdminTodo() {
    if (!newAdminTodoText.trim()) return;
    
    try {
      await api.createDefaultTask(newAdminTodoText);
      newAdminTodoText = "";
      await loadAdminTodos();
    } catch (e) {
      console.error("Failed to create default task", e);
    }
  }

  function startEdit(todo: Todo) {
    editingTodoId = todo.id;
    editingTodoText = todo.text;
  }

  function cancelEdit() {
    editingTodoId = null;
    editingTodoText = "";
  }

  async function saveEdit(id: string) {
    if (!editingTodoText.trim()) return;
    
    try {
      await api.updateDefaultTask(id, editingTodoText);
      editingTodoId = null;
      editingTodoText = "";
      await loadAdminTodos();
    } catch (e) {
      console.error("Failed to update default task", e);
    }
  }

  async function handleDelete(id: string) {
    confirmModalConfig = {
      title: 'Delete Default Task',
      message: 'Are you sure you want to delete this default task? It will be removed for all users and cannot be undone.',
      danger: true,
      onConfirm: async () => {
        try {
          await api.deleteDefaultTask(id);
          await loadAdminTodos();
        } catch (e) {
          console.error("Failed to delete default task", e);
        }
        showConfirmModal = false;
      }
    };
    showConfirmModal = true;
  }
</script>

<div class="max-w-2xl mx-auto">
  <header class="mb-4">
    <h2 class="text-2xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 tracking-tight mb-1">
      Default Tasks
    </h2>
    <p class="text-sm text-slate-500">Manage tasks that are automatically assigned to all users</p>
  </header>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin"></div>
    </div>
  {:else}
    <div class="space-y-6">
      <!-- Add Admin Todo Form -->
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-bold text-slate-800 mb-2">Add Default Task</h2>
        <p class="text-sm text-slate-500 mb-4">
          This task will be automatically assigned to <strong>all users</strong>. Each user tracks their own status for this task.
        </p>
        <form onsubmit={(e) => { e.preventDefault(); handleCreateAdminTodo(); }} class="flex gap-3">
          <input
            bind:value={newAdminTodoText}
            placeholder="Enter task for all users..."
            maxlength="200"
            class="flex-1 px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500"
          />
          <button
            type="submit"
            class="px-6 py-2 bg-indigo-600 text-white font-bold rounded-lg hover:bg-indigo-700 transition-colors"
          >
            Add Task
          </button>
        </form>
      </div>

      <!-- Admin Todos List -->
      <div class="glass-card rounded-2xl overflow-hidden divide-y divide-slate-100">
        {#if adminTodos.length === 0}
          <div class="p-12 text-center">
            <p class="text-slate-500">No default tasks yet. Create one above!</p>
          </div>
        {:else}
          {#each adminTodos as todo}
            <div class="p-4 flex items-center gap-4 hover:bg-slate-50 transition-colors">
              {#if editingTodoId === todo.id}
                <!-- Edit Mode -->
                <form onsubmit={(e) => { e.preventDefault(); saveEdit(todo.id); }} class="flex-1 flex gap-2">
                  <input
                    bind:value={editingTodoText}
                    maxlength="200"
                    class="flex-1 px-3 py-1.5 border border-indigo-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
                    use:focus
                  />
                  <button
                    type="submit"
                    class="p-2 text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors"
                    title="Save"
                    aria-label="Save task"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                  </button>
                  <button
                    type="button"
                    onclick={cancelEdit}
                    class="p-2 text-slate-400 hover:bg-slate-100 rounded-lg transition-colors"
                    title="Cancel"
                    aria-label="Cancel editing"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <line x1="18" y1="6" x2="6" y2="18"></line>
                      <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                  </button>
                </form>
              {:else}
                <!-- View Mode -->
                <span class="flex-1 font-medium text-slate-700">
                  {todo.text}
                </span>
                <span class="text-xs text-slate-400">
                  {new Date(todo.created).toLocaleDateString()}
                </span>
                <div class="flex items-center gap-1">
                  <button
                    onclick={() => startEdit(todo)}
                    class="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                    title="Edit"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"></path>
                      <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                    </svg>
                  </button>
                  <button
                    onclick={() => handleDelete(todo.id)}
                    class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                    title="Delete"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="3 6 5 6 21 6"></polyline>
                      <path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"></path>
                    </svg>
                  </button>
                </div>
              {/if}
            </div>
          {/each}
        {/if}
      </div>
    </div>
  {/if}
</div>

{#if showConfirmModal}
  <ConfirmModal
    title={confirmModalConfig.title}
    message={confirmModalConfig.message}
    danger={confirmModalConfig.danger}
    on:confirm={confirmModalConfig.onConfirm}
    on:cancel={() => showConfirmModal = false}
  />
{/if}
