<script lang="ts">
  import { page } from "$app/stores";
  import UserList from "$lib/components/admin/UserList.svelte";
  import DefaultTodoList from "$lib/components/admin/DefaultTodoList.svelte";

  let activeTab = $state('users');

  function setActiveTab(tab: string) {
    activeTab = tab;
  }
</script>

<div class="max-w-6xl mx-auto p-4 md:p-8">
  <header class="mb-8">
    <h1 class="text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 tracking-tight mb-2">
      Admin Dashboard
    </h1>
    <p class="text-sm text-slate-500">Manage users and default settings</p>
  </header>

  <div class="mb-8 border-b border-slate-200">
    <nav class="-mb-px flex space-x-8" aria-label="Tabs">
      <button
        onclick={() => setActiveTab('users')}
        class="{activeTab === 'users'
          ? 'border-indigo-500 text-indigo-600'
          : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
          whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
      >
        Users
      </button>

      <button
        onclick={() => setActiveTab('default-todos')}
        class="{activeTab === 'default-todos'
          ? 'border-indigo-500 text-indigo-600'
          : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
          whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors"
      >
        Default Tasks
      </button>
    </nav>
  </div>

  {#if activeTab === 'users'}
    <UserList />
  {:else if activeTab === 'default-todos'}
    <DefaultTodoList />
  {/if}
</div>
