package models

import (
	"log"
	"time"
)

type Todo struct {
	ID        int
	Content   string
	UserID    int
	CreatedAt time.Time
}

func (u *User) CreateTodo(content string) (err error) {
	cmd := `INSERT INTO todos (content, user_id, created_at) VALUES (?, ?, ?)`
	_, err = DB.Exec(cmd, content, u.ID, time.Now())
	if err != nil {
		log.Println(err)
	}
	return err
}

func GetTodo(id int) (todo Todo, err error) {
	todo = Todo{}
	cmd := `SELECT id, content, user_id, created_at FROM todos WHERE id = ?`
	if err = DB.QueryRow(cmd, id).Scan(
		&todo.ID,
		&todo.Content,
		&todo.UserID,
		&todo.CreatedAt,
	); err != nil {
		log.Println(err)
	}

	return todo, nil
}

func GetTodos() (todos []Todo, err error) {

	cmd := `SELECT id, content, user_id, created_at FROM todos`
	rows, err := DB.Query(cmd)
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

func (u *User) GetTodosByUser() (todos []Todo, err error) {
	cmd := `SELECT id, content, user_id, created_at FROM todos WHERE user_id = ?`
	rows, err := DB.Query(cmd, u.ID)
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

func (t *Todo) UpdateTodo() error {
	cmd := `UPDATE todos SET content = ?, user_id = ? WHERE id = ?`
	_, err = DB.Exec(cmd, t.Content, t.UserID, t.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (t *Todo) DeleteTodo() error {
	cmd := `DELETE FROM todos WHERE id = ?`
	_, err = DB.Exec(cmd, t.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}
