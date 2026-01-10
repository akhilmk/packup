import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TodoService } from "../gen/proto/todo_connect";

// The backend runs on localhost:8090 by default
const transport = createConnectTransport({
    baseUrl: "http://127.0.0.1:8090",
});

export const request = createClient(TodoService, transport);
