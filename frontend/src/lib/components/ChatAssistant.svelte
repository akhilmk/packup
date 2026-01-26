<script lang="ts">
  import { onMount } from 'svelte';
  import { chatbotConfigStore } from '../config';

  export let maxWidth = "max-w-2xl"; 

  let isOpen = false;
  let messages: { role: 'user' | 'assistant', text: string }[] = [];
  let inputText = "";
  let loading = false;
  let chatContainer: HTMLDivElement;

  const CONVO_STORAGE_KEY = 'packup_chat_messages';

  onMount(() => {
    const savedConvo = sessionStorage.getItem(CONVO_STORAGE_KEY);
    if (savedConvo) {
      try {
        messages = JSON.parse(savedConvo);
      } catch (e) {
        console.warn('Failed to restore chat history');
      }
    }
  });

  $: if (messages.length > 0) {
    sessionStorage.setItem(CONVO_STORAGE_KEY, JSON.stringify(messages));
    // Scroll down whenever messages change and window is open
    if (isOpen) {
      setTimeout(() => {
        if (chatContainer) chatContainer.scrollTop = chatContainer.scrollHeight;
      }, 50);
    }
  }

  async function query(data: { question: string }) {
    if (!$chatbotConfigStore.enabled) return;

    const response = await fetch(
      $chatbotConfigStore.apiUrl,
      {
        headers: {
          Authorization: `Bearer ${$chatbotConfigStore.apiToken}`,
          "Content-Type": "application/json"
        },
        method: "POST",
        body: JSON.stringify(data)
      }
    );

    if (!response.ok) {
       throw new Error(`Chat API error: ${response.status}`);
    }

    const result = await response.json();
    return result;
  }

  async function handleSendMessage() {
    if (!inputText.trim() || loading) return;

    const userMessage = inputText.trim();
    inputText = "";

    messages = [...messages, { role: 'user', text: userMessage }];
    loading = true;

    try {
      const response = await query({ question: userMessage });
      const assistantText = response?.text || response?.answer || "I'm here to help with your travel plans!";
      messages = [...messages, { role: 'assistant', text: assistantText }];
    } catch (error) {
      console.error("Chat error:", error);
      messages = [...messages, { 
        role: 'assistant', 
        text: "I'm sorry, I'm currently unavailable to help with travel plans. Please try again later." 
      }];
    } finally {
      loading = false;
    }
  }

  function toggleChat() {
    isOpen = !isOpen;
    if (isOpen && messages.length === 0) {
      messages = [{
        role: 'assistant',
        text: "Hi! I'm your travel assistant 'Orama Buddy'. How can I help you today?"
      }];
    }
  }

  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSendMessage();
    }
  }

  // Export a clear function for logout
  export function clearHistory() {
    messages = [];
    sessionStorage.removeItem(CONVO_STORAGE_KEY);
  }
</script>

{#if $chatbotConfigStore.enabled && $chatbotConfigStore.apiUrl}
  <div class="fixed bottom-6 left-0 right-0 pointer-events-none z-50 px-6 sm:px-0">
    <div class="mx-auto {maxWidth} relative h-0">
      
      <div class="absolute bottom-0 right-0 sm:-right-4 translate-y-0 sm:translate-x-1/2">
        <button
          on:click={toggleChat}
          class="pointer-events-auto w-14 h-14 bg-gradient-to-br from-indigo-600 to-purple-600 text-white rounded-full shadow-lg hover:shadow-xl transition-all duration-300 flex items-center justify-center hover:scale-110 group relative"
          aria-label="Open chat assistant"
        >
          {#if isOpen}
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          {:else}
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
            </svg>
          {/if}
          
          {#if !isOpen}
            <span class="absolute inset-0 rounded-full bg-indigo-400 opacity-75 animate-ping group-hover:animate-none -z-10"></span>
          {/if}
        </button>
      </div>

      {#if isOpen}
        <div class="absolute bottom-20 right-0 sm:right-0 translate-x-0 pointer-events-auto w-[calc(100vw-3rem)] sm:w-96 h-[500px] bg-white rounded-2xl shadow-2xl flex flex-col overflow-hidden border border-slate-200 animate-slideUp">
          <!-- Header -->
          <div class="bg-gradient-to-r from-indigo-600 to-purple-600 text-white p-4 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 bg-white/20 rounded-full flex items-center justify-center backdrop-blur-sm">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                </svg>
              </div>
              <div>
                <h3 class="font-bold text-sm">Travel Assistant</h3>
                <p class="text-xs text-indigo-100">Ask me about your travel plans</p>
              </div>
            </div>
            <button 
              on:click={() => isOpen = false}
              class="text-white/80 hover:text-white transition-colors"
              aria-label="Close chat"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          <div 
            bind:this={chatContainer}
            class="flex-1 overflow-y-auto p-4 space-y-4 bg-slate-50"
          >
            {#each messages as message}
              <div class="flex {message.role === 'user' ? 'justify-end' : 'justify-start'}">
                <div class="max-w-[80%] {message.role === 'user' 
                  ? 'bg-gradient-to-br from-indigo-600 to-purple-600 text-white' 
                  : 'bg-white text-slate-800 border border-slate-200'} 
                  rounded-2xl px-4 py-2 shadow-sm"
                >
                  <p class="text-sm leading-relaxed">{message.text}</p>
                </div>
              </div>
            {/each}
            
            {#if loading}
              <div class="flex justify-start">
                <div class="bg-white border border-slate-200 rounded-2xl px-4 py-3 shadow-sm">
                  <div class="flex gap-1">
                    <div class="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style="animation-delay: 0ms"></div>
                    <div class="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style="animation-delay: 150ms"></div>
                    <div class="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style="animation-delay: 300ms"></div>
                  </div>
                </div>
              </div>
            {/if}
          </div>

          <div class="p-4 bg-white border-t border-slate-200">
            <form on:submit|preventDefault={handleSendMessage} class="flex gap-2">
              <input
                bind:value={inputText}
                on:keypress={handleKeyPress}
                placeholder="Ask about your travel plans..."
                disabled={loading}
                class="flex-1 px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 text-sm disabled:bg-slate-50 disabled:text-slate-400"
              />
              <button
                type="submit"
                disabled={!inputText.trim() || loading}
                class="px-4 py-2 bg-gradient-to-br from-indigo-600 to-purple-600 text-white rounded-lg hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                aria-label="Send message"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="22" y1="2" x2="11" y2="13"></line>
                  <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                </svg>
              </button>
            </form>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-slideUp {
    animation: slideUp 0.3s ease-out;
  }

  div::-webkit-scrollbar {
    width: 6px;
  }

  div::-webkit-scrollbar-track {
    background: transparent;
  }

  div::-webkit-scrollbar-thumb {
    background: #cbd5e1;
    border-radius: 3px;
  }

  div::-webkit-scrollbar-thumb:hover {
    background: #94a3b8;
  }
</style>
