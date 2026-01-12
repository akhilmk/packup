<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export let status: 'pending' | 'in-progress' | 'done' = 'pending';
  export let clickable = false;
  
  const dispatch = createEventDispatcher();
  
  function handleClick() {
    if (clickable) {
      dispatch('click');
    }
  }
</script>

<button 
  on:click={handleClick}
  disabled={!clickable}
  class="relative flex items-center justify-center w-6 h-6 rounded-full border-2 transition-all duration-200 
         {clickable ? 'cursor-pointer hover:scale-110 active:scale-95 hover:ring-2 hover:ring-offset-1 hover:ring-indigo-200' : 'cursor-default'}
         {status === 'done' ? 'bg-emerald-500 border-emerald-500' : 
          status === 'in-progress' ? 'bg-amber-100 border-amber-400' : 
          'bg-white border-slate-200 ' + (clickable ? 'hover:border-indigo-400' : '')}"
  title={clickable ? "Click to cycle status" : "Status"}
  type="button"
>
  {#if status === 'done'}
    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="20 6 9 17 4 12"></polyline>
    </svg>
  {:else if status === 'in-progress'}
    <div class="w-2.5 h-2.5 bg-amber-500 rounded-full animate-pulse"></div>
  {/if}
</button>
