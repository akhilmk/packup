import { writable } from 'svelte/store';
import { api } from './api';

export interface ChatbotConfig {
    enabled: boolean;
    apiUrl: string;
    apiToken: string;
}

const STORAGE_KEY = 'packup_chatbot_config';

// Initial state
const initialConfig: ChatbotConfig = {
    enabled: false,
    apiUrl: '',
    apiToken: '',
};

export const chatbotConfigStore = writable<ChatbotConfig>(initialConfig);

let isFetching = false;

// Function to fetch config from backend
export const loadRuntimeConfig = async () => {
    // 1. Check if we already have it in local state
    let current: ChatbotConfig = { enabled: false, apiUrl: '', apiToken: '' };
    chatbotConfigStore.subscribe(val => current = val)();
    if (current.apiUrl) return;

    // 2. Check if we have it in sessionStorage (avoid network hit on refresh)
    const cached = sessionStorage.getItem(STORAGE_KEY);
    if (cached) {
        try {
            const parsed = JSON.parse(cached);
            chatbotConfigStore.set(parsed);
            return;
        } catch (e) {
            sessionStorage.removeItem(STORAGE_KEY);
        }
    }

    // 3. Only if not cached, fetch from server
    if (isFetching) return;
    isFetching = true;

    try {
        const config = await api.getConfig();
        const newConfig = {
            enabled: config.chatbot_enabled,
            apiUrl: config.chatbot_api_url,
            apiToken: config.chatbot_api_token,
        };

        chatbotConfigStore.set(newConfig);
        // Cache for this session
        sessionStorage.setItem(STORAGE_KEY, JSON.stringify(newConfig));
    } catch (error) {
        console.warn('Chatbot config unavailable');
    } finally {
        isFetching = false;
    }
};

// Clear cache on logout
export const clearConfigCache = () => {
    sessionStorage.removeItem(STORAGE_KEY);
    chatbotConfigStore.set(initialConfig);
};
