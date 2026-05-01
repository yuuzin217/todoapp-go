package models

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// Todo はデータベースの todos テーブルのレコードを表す構造体です。
// 各タスクの内容とその所有者情報を保持します。
type Todo struct {
	ID        int       // TODO の一意な ID
	Content   string    // タスクの具体的な内容
	UserID    int       // このタスクを作成したユーザーの ID（リレーション用）
	CreatedAt time.Time // タスクが作成された日時
}

// CreateTodo は指定された内容で新しい TODO をデータベースに保存します。
// タスクの所有者として現在のユーザー(User 構造体)が紐付けられます。
func (u *User) CreateTodo(ctx context.Context, db *sql.DB, content string) (err error) {
	cmd := `INSERT INTO todos (content, user_id, created_at) VALUES (?, ?, ?)`
	_, err = db.ExecContext(ctx, cmd, content, u.ID, time.Now())
	if err != nil {
		log.Printf("Failed to insert todo for user %d: %v", u.ID, err)
	}
	return err
}

// GetTodo は指定された ID の TODO をデータベースから一件取得します。
// 取得したタスクの UserID を確認することで、認可（所有権のチェック）が可能です。
func GetTodo(ctx context.Context, db *sql.DB, id int) (todo Todo, err error) {
	todo = Todo{}
	cmd := `SELECT id, content, user_id, created_at FROM todos WHERE id = ?`
	err = db.QueryRowContext(ctx, cmd, id).Scan(
		&todo.ID,
		&todo.Content,
		&todo.UserID,
		&todo.CreatedAt,
	)
	if err != nil {
		log.Printf("Failed to query todo %d: %v", id, err)
	}

	return todo, err
}

// GetTodos はデータベースに保存されているすべての TODO を取得します。
// TODO: 管理者用機能や統計用として想定していますが、現状は全ユーザーのデータが混ざるため注意が必要です。
func GetTodos(ctx context.Context, db *sql.DB) (todos []Todo, err error) {
	cmd := `SELECT id, content, user_id, created_at FROM todos`
	rows, err := db.QueryContext(ctx, cmd)
	if err != nil {
		log.Printf("Failed to list all todos: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		if err = rows.Scan(
			&todo.ID,
			&todo.Content,
			&todo.UserID,
			&todo.CreatedAt,
		); err != nil {
			log.Printf("Failed to scan todo row: %v", err)
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// GetTodosByUser は特定のユーザーが作成した TODO 一覧のみを抽出して取得します。
// ユーザー自身のマイページ（index）を表示する際に主要な役割を果たします。
func (u *User) GetTodosByUser(ctx context.Context, db *sql.DB) (todos []Todo, err error) {
	cmd := `SELECT id, content, user_id, created_at FROM todos WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := db.QueryContext(ctx, cmd, u.ID)
	if err != nil {
		log.Printf("Failed to list todos for user %d: %v", u.ID, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		if err = rows.Scan(
			&todo.ID,
			&todo.Content,
			&todo.UserID,
			&todo.CreatedAt,
		); err != nil {
			log.Printf("Failed to scan todo row for user %d: %v", u.ID, err)
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// UpdateTodo は指定された ID の TODO 内容をデータベース上で更新します。
// 意図しない書き換えを防ぐため、UserID を WHERE 句に含めて所有権を保証します。
func (t *Todo) UpdateTodo(ctx context.Context, db *sql.DB) error {
	cmd := `UPDATE todos SET content = ? WHERE id = ? AND user_id = ?`
	_, err := db.ExecContext(ctx, cmd, t.Content, t.ID, t.UserID)
	if err != nil {
		log.Printf("Failed to update todo %d for user %d: %v", t.ID, t.UserID, err)
	}
	return err
}

// DeleteTodo は TODO をデータベースから物理削除します。
func (t *Todo) DeleteTodo(ctx context.Context, db *sql.DB) error {
	cmd := `DELETE FROM todos WHERE id = ?`
	_, err := db.ExecContext(ctx, cmd, t.ID)
	if err != nil {
		log.Printf("Failed to delete todo %d: %v", t.ID, err)
	}
	return err
}
