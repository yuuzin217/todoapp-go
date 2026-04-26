package models

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// Todo はデータベースの todos テーブルのレコードを表す構造体です。
type Todo struct {
	ID        int       // TODOの一意なID
	Content   string    // TODOの内容
	UserID    int       // 作成したユーザーのID
	CreatedAt time.Time // 作成日時
}

// CreateTodo は指定された内容で新しいTODOをデータベースに保存します。
func (u *User) CreateTodo(ctx context.Context, db *sql.DB, content string) (err error) {
	cmd := `INSERT INTO todos (content, user_id, created_at) VALUES (?, ?, ?)`
	_, err = db.ExecContext(ctx, cmd, content, u.ID, time.Now())
	if err != nil {
		log.Println(err)
	}
	return err
}

// GetTodo は指定されたIDのTODOをデータベースから取得します。
func GetTodo(ctx context.Context, db *sql.DB, id int) (todo Todo, err error) {
	todo = Todo{}
	cmd := `SELECT id, content, user_id, created_at FROM todos WHERE id = ?`
	if err = db.QueryRowContext(ctx, cmd, id).Scan(
		&todo.ID,
		&todo.Content,
		&todo.UserID,
		&todo.CreatedAt,
	); err != nil {
		log.Println(err)
	}

	return todo, nil
}

// GetTodos はデータベースに保存されているすべてのTODOを取得します。
func GetTodos(ctx context.Context, db *sql.DB) (todos []Todo, err error) {

	cmd := `SELECT id, content, user_id, created_at FROM todos`
	rows, err := db.QueryContext(ctx, cmd)
	if err != nil {
		log.Println(err)
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
			log.Println(err)
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, err
}

// GetTodosByUser は特定のユーザーが作成したTODO一覧を取得します。
func (u *User) GetTodosByUser(ctx context.Context, db *sql.DB) (todos []Todo, err error) {
	cmd := `SELECT id, content, user_id, created_at FROM todos WHERE user_id = ?`
	rows, err := db.QueryContext(ctx, cmd, u.ID)
	if err != nil {
		log.Println(err)
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
			log.Println(err)
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, err
}

// UpdateTodo はTODOの内容をデータベース上で更新します。
func (t *Todo) UpdateTodo(ctx context.Context, db *sql.DB) error {
	cmd := `UPDATE todos SET content = ?, user_id = ? WHERE id = ?`
	_, err := db.ExecContext(ctx, cmd, t.Content, t.UserID, t.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

// DeleteTodo はTODOをデータベースから削除します。
func (t *Todo) DeleteTodo(ctx context.Context, db *sql.DB) error {
	cmd := `DELETE FROM todos WHERE id = ?`
	_, err := db.ExecContext(ctx, cmd, t.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}
