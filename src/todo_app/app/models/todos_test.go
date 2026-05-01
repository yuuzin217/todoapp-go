package models

import (
	"context"
	"testing"
)

// TestCreateTodo は TODO の作成機能が正常に動作し、DB に保存されることを検証します。
func TestCreateTodo(t *testing.T) {
	u := &User{Name: "todouser", Email: "todo@example.com", Password: "password"}
	u.CreateUser(context.Background(), testDB)
	user, _ := GetUserByEmail(context.Background(), testDB, u.Email)

	err := user.CreateTodo(context.Background(), testDB, "Test Content")
	if err != nil {
		t.Fatalf("CreateTodo failed: %v", err)
	}
}

// TestGetTodos は特定のユーザーに関連付けられた TODO の取得、
// および ID による単体取得が正しく機能することを検証します。
func TestGetTodos(t *testing.T) {
	u := &User{Name: "getuser", Email: "gettodo@example.com", Password: "password"}
	u.CreateUser(context.Background(), testDB)
	user, _ := GetUserByEmail(context.Background(), testDB, u.Email)

	user.CreateTodo(context.Background(), testDB, "Task 1")
	user.CreateTodo(context.Background(), testDB, "Task 2")

	// GetTodosByUser: ユーザーに紐づく TODO リストが正しく取得できるか。
	todos, err := user.GetTodosByUser(context.Background(), testDB)
	if err != nil {
		t.Fatalf("GetTodosByUser failed: %v", err)
	}
	if len(todos) < 2 {
		t.Errorf("Expected at least 2 todos, but got %d", len(todos))
	}

	// GetTodo: 指定した ID の TODO が期待通りの内容で取得できるか。
	todo, err := GetTodo(context.Background(), testDB, todos[0].ID)
	if err != nil {
		t.Fatalf("GetTodo failed: %v", err)
	}
	if todo.Content != todos[0].Content {
		t.Errorf("Expected content %s, but got %s", todos[0].Content, todo.Content)
	}

	// GetTodos: システム全体の TODO が取得できるか (管理者用途等を想定)。
	allTodos, err := GetTodos(context.Background(), testDB)
	if err != nil {
		t.Fatalf("GetTodos failed: %v", err)
	}
	if len(allTodos) < 2 {
		t.Errorf("Expected at least 2 total todos, but got %d", len(allTodos))
	}

	// 存在しない ID を指定した際のエラーハンドリングを検証。
	_, err = GetTodo(context.Background(), testDB, 999)
	if err == nil {
		t.Error("Expected error for non-existent todo ID, but got nil")
	}
}

// TestUpdateDeleteTodo は TODO の更新と削除のライフサイクルを検証します。
func TestUpdateDeleteTodo(t *testing.T) {
	u := &User{Name: "upduser", Email: "updtodo@example.com", Password: "password"}
	u.CreateUser(context.Background(), testDB)
	user, _ := GetUserByEmail(context.Background(), testDB, u.Email)

	user.CreateTodo(context.Background(), testDB, "Old Content")
	todos, _ := user.GetTodosByUser(context.Background(), testDB)
	todo := todos[0]

	// Update: 内容の変更が DB に反映されるか。
	todo.Content = "New Content"
	err := todo.UpdateTodo(context.Background(), testDB)
	if err != nil {
		t.Fatalf("UpdateTodo failed: %v", err)
	}

	updated, _ := GetTodo(context.Background(), testDB, todo.ID)
	if updated.Content != "New Content" {
		t.Errorf("Expected New Content, but got %s", updated.Content)
	}

	// Delete: 削除後にデータが取得不能になるか。
	err = todo.DeleteTodo(context.Background(), testDB)
	if err != nil {
		t.Fatalf("DeleteTodo failed: %v", err)
	}

	_, err = GetTodo(context.Background(), testDB, todo.ID)
	if err == nil {
		t.Error("Expected error for deleted todo, but got nil")
	}
}
