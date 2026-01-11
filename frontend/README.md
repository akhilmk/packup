# REST API Migration - Frontend

This document describes the migration from ConnectRPC to pure REST API.

## Changes Made

### 1. New REST API Client (`src/lib/api.ts`)
Created a new TypeScript module that provides typed REST API calls:
- `listTodos()` - GET /api/todos
- `createTodo(text)` - POST /api/todos
- `updateTodo(id, updates)` - PUT /api/todos/{id}
- `deleteTodo(id)` - DELETE /api/todos/{id}

### 2. Updated Components
All Svelte components now use the REST API client:
- `TodoList.svelte` - Fetches todos using REST
- `AddTodo.svelte` - Creates todos using REST
- `TodoItem.svelte` - Updates and deletes todos using REST

### 3. Removed Dependencies
- `@bufbuild/protoc-gen-es`
- `@connectrpc/protoc-gen-connect-es`
- Removed `src/gen/` directory (generated protobuf files)
- Removed `src/lib/client.ts` (old ConnectRPC client)

### 4. API Configuration
The frontend now connects to `http://localhost:8080/api` (configurable in `api.ts`)

## Building the Frontend

```bash
# Install dependencies
npm install

# Build for production (creates dist/ folder)
npm run build

# Run development server
npm run dev
```

## Using Make Commands

```bash
# Install dependencies
make frontend-install

# Build static files
make frontend-build

# Clean dist folder
make frontend-clean

# Build both frontend and backend
make build-all
```

## API Endpoints

The backend exposes the following REST endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/todos` | List all todos |
| POST | `/api/todos` | Create a new todo |
| PUT | `/api/todos/{id}` | Update a todo |
| DELETE | `/api/todos/{id}` | Delete a todo |

## Todo Type Definition

```typescript
interface Todo {
  id: string;
  text: string;
  completed: boolean;
}
```
