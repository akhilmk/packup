// REST API client for Todo backend
const API_BASE_URL = "http://localhost:8080/api";

export interface Todo {
    id: string;
    text: string;
    completed: boolean;
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

    async updateTodo(id: string, updates: { text?: string; completed?: boolean }): Promise<Todo> {
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
};
