package main

import (
	"context"
	"errors"
	"log"
	"os"

	proto "todo-app/gen/proto"
	"todo-app/gen/proto/todov1connect"

	"connectrpc.com/connect"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type TodoServer struct {
	app *pocketbase.PocketBase
}

func (s *TodoServer) AddTodo(ctx context.Context, req *connect.Request[proto.AddTodoRequest]) (*connect.Response[proto.Todo], error) {
	text := req.Msg.Text
	if len(text) > 200 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("text limit of 200 characters exceeded"))
	}
	if text == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("text cannot be empty"))
	}

	collection, err := s.app.FindCollectionByNameOrId("todos")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	record := core.NewRecord(collection)
	record.Set("text", text)
	record.Set("completed", false)

	if err := s.app.Save(record); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&proto.Todo{
		Id:        record.Id,
		Text:      record.GetString("text"),
		Completed: record.GetBool("completed"),
	}), nil
}

func (s *TodoServer) ListTodos(ctx context.Context, req *connect.Request[proto.ListTodosRequest]) (*connect.Response[proto.ListTodosResponse], error) {
	records, err := s.app.FindRecordsByFilter(
		"todos",
		"created != ''", // filter
		"-created",      // sort
		100,             // limit
		0,               // offset
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	todos := make([]*proto.Todo, len(records))
	for i, r := range records {
		todos[i] = &proto.Todo{
			Id:        r.Id,
			Text:      r.GetString("text"),
			Completed: r.GetBool("completed"),
		}
	}

	return connect.NewResponse(&proto.ListTodosResponse{Todos: todos}), nil
}

func (s *TodoServer) UpdateTodo(ctx context.Context, req *connect.Request[proto.UpdateTodoRequest]) (*connect.Response[proto.Todo], error) {
	record, err := s.app.FindRecordById("todos", req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("todo not found"))
	}

	if req.Msg.Text != "" {
		if len(req.Msg.Text) > 200 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("text limit of 200 characters exceeded"))
		}
		record.Set("text", req.Msg.Text)
	}
	record.Set("completed", req.Msg.Completed)

	if err := s.app.Save(record); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&proto.Todo{
		Id:        record.Id,
		Text:      record.GetString("text"),
		Completed: record.GetBool("completed"),
	}), nil
}

func (s *TodoServer) DeleteTodo(ctx context.Context, req *connect.Request[proto.DeleteTodoRequest]) (*connect.Response[proto.DeleteTodoResponse], error) {
	record, err := s.app.FindRecordById("todos", req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("todo not found"))
	}

	if err := s.app.Delete(record); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&proto.DeleteTodoResponse{Id: req.Msg.Id}), nil
}

func main() {
	app := pocketbase.New()

	// Configure "todos" collection creation
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		_, err := app.FindCollectionByNameOrId("todos")
		if err != nil {
			// Create collection
			collection := core.NewBaseCollection("todos")
			collection.Fields.Add(&core.TextField{
				Name:     "text",
				Required: true,
				Max:      200,
			})
			collection.Fields.Add(&core.BoolField{
				Name: "completed",
			})

			// Save collection
			if err := app.Save(collection); err != nil {
				// Don't fail if already exists (race condition)
				log.Printf("Warning: failed to create collection: %v", err)
			} else {
				log.Println("Created 'todos' collection")
			}
		}

		// Register ConnectRPC handler
		path, handler := todov1connect.NewTodoServiceHandler(&TodoServer{app: app})

		// Mount handler
		// Note: PocketBase router uses slightly different signature or method
		e.Router.POST(path+"/*", apis.WrapStdHandler(handler))
		e.Router.GET(path+"/*", apis.WrapStdHandler(handler))

		// Serve static frontend files
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("../frontend/dist"), true))

		log.Printf("ConnectRPC service mounted at %s", path)
		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
