// REST API client for Todo backend
const API_BASE_URL = "/api";

export type TodoStatus = 'pending' | 'in-progress' | 'done';

export interface Todo {
    id: string;
    text: string;
    status: TodoStatus;
    created: string; // ISO date string
    position?: number;
    is_default_task?: boolean;
    shared_with_admin?: boolean;
    created_by_user_id?: string;
}

interface ListTodosResponse {
    todos: Todo[];
}

class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
        this.name = "ApiError";
    }
}

async function handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
        const text = await response.text();
        throw new ApiError(response.status, text || response.statusText);
    }
    return response.json();
}

export const api = {
    async listTodos(): Promise<Todo[]> {
        const response = await fetch(`${API_BASE_URL}/todos`);
        const data = await handleResponse<ListTodosResponse>(response);
        return data.todos;
    },

    async createTodo(text: string): Promise<Todo> {
        const response = await fetch(`${API_BASE_URL}/todos`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ text }),
        });
        return handleResponse<Todo>(response);
    },

    async updateTodo(id: string, updates: { text?: string; status?: TodoStatus; shared_with_admin?: boolean }): Promise<Todo> {
        const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(updates),
        });
        return handleResponse<Todo>(response);
    },

    async deleteTodo(id: string): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/todos/${id}`, {
            method: "DELETE",
        });
        await handleResponse<{ success: boolean }>(response);
    },

    async reorderTodos(ids: string[]): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/todos/reorder`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ ids }),
        });
        await handleResponse<{ success: boolean }>(response);
    },

    // Auth
    async getMe(): Promise<User> {
        const response = await fetch(`${API_BASE_URL}/auth/me`);
        return handleResponse<User>(response);
    },

    async logout(): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/auth/logout`, {
            method: "POST",
        });
        if (!response.ok) throw new Error("Logout failed");
    },

    // Admin endpoints
    async listUsers(): Promise<User[]> {
        const response = await fetch(`${API_BASE_URL}/admin/users`);
        const data = await handleResponse<{ users: User[] }>(response);
        return data.users;
    },

    async listDefaultTasks(): Promise<Todo[]> {
        const response = await fetch(`${API_BASE_URL}/admin/todos`);
        const data = await handleResponse<{ todos: Todo[] }>(response);
        return data.todos;
    },

    async listUserTodos(userId: string): Promise<Todo[]> {
        const response = await fetch(`${API_BASE_URL}/admin/users/${userId}/todos`);
        const data = await handleResponse<{ todos: Todo[] }>(response);
        return data.todos;
    },

    async updateUserTodo(userId: string, todoId: string, updates: { status: TodoStatus }): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/admin/users/${userId}/todos/${todoId}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(updates),
        });
        await handleResponse<{ success: boolean }>(response);
    },

    async createDefaultTask(text: string): Promise<Todo> {
        const response = await fetch(`${API_BASE_URL}/admin/todos`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ text }),
        });
        return handleResponse<Todo>(response);
    },

    async updateDefaultTask(id: string, text: string): Promise<Todo> {
        const response = await fetch(`${API_BASE_URL}/admin/todos/${id}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ text }),
        });
        return handleResponse<Todo>(response);
    },

    async deleteDefaultTask(id: string): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/admin/todos/${id}`, {
            method: "DELETE",
        });
        await handleResponse<{ success: boolean }>(response);
    }
};

export interface User {
    id: string;
    email: string;
    name: string;
    avatar_url: string;
    role: 'admin' | 'user';
    created_at: string;
}
