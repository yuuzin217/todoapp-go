package models

import (
	"context"
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	UUID      string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	Todos     []Todo
}

type Session struct {
	ID        int
	UUID      string
	Email     string
	UserID    int
	CreatedAt time.Time
}

func (u *User) CreateUser(ctx context.Context, db *sql.DB) (err error) {
	cmd := `INSERT INTO users(
		uuid,
		name,
		email,
		password,
		created_at) values (?, ?, ?, ?, ?)`

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = db.ExecContext(
		ctx,
		cmd,
		createUUID(),
		u.Name,
		u.Email,
		string(hash),
		time.Now(),
	)
	if err != nil {
		log.Println(err)
	}
	return err
}

func GetUser(ctx context.Context, db *sql.DB, id int) (user User, err error) {
	user = User{}
	cmd := `SELECT id,uuid, name, email, password, created_at FROM users WHERE id = ?`
	err = db.QueryRowContext(ctx, cmd, id).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	return user, err
}

func (u *User) UpdateUser(ctx context.Context, db *sql.DB) (err error) {
	cmd := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	_, err = db.ExecContext(ctx, cmd, u.Name, u.Email, u.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (u *User) DeleteUser(ctx context.Context, db *sql.DB) (err error) {
	cmd := `DELETE FROM users WHERE id = ?`
	_, err = db.ExecContext(ctx, cmd, u.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

/*
GetUserByEmail は Eメールからユーザー情報を取得します。
*/
func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (user User, err error) {
	user = User{}
	cmd := `SELECT id, uuid, name, email, password, created_at
	FROM users WHERE email = ?`
	err = db.QueryRowContext(ctx, cmd, email).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	return user, err
}

func (u *User) CreateSession(ctx context.Context, db *sql.DB) (session Session, err error) {
	session = Session{}
	cmd1 := `INSERT INTO sessions (uuid, email, user_id, created_at) VALUES (?, ?, ?, ?)`
	_, err = db.ExecContext(ctx, cmd1, createUUID(), u.Email, u.ID, time.Now())
	if err != nil {
		log.Println(err)
	}

	cmd2 := `SELECT id, uuid, email, user_id, created_at FROM sessions WHERE user_id = ? AND email = ?`
	if err = db.QueryRowContext(ctx, cmd2, u.ID, u.Email).Scan(
		&session.ID,
		&session.UUID,
		&session.Email,
		&session.UserID,
		&session.CreatedAt,
	); err != nil {
		log.Println(err)
	}

	return session, nil
}

/*
CheckSession は セッションを確認します。
*/
func (s *Session) CheckSession(ctx context.Context, db *sql.DB) (valid bool, err error) {
	cmd := `select id, uuid, email, user_id, created_at
	 from sessions where uuid = ?`

	err = db.QueryRowContext(ctx, cmd, s.UUID).Scan(
		&s.ID,
		&s.UUID,
		&s.Email,
		&s.UserID,
		&s.CreatedAt)

	if err != nil {
		valid = false
		return
	}
	if s.ID != 0 {
		valid = true
	}
	return valid, err
}

func (s *Session) DeleteSessionByUUID(ctx context.Context, db *sql.DB) (err error) {
	cmd := `DELETE FROM sessions WHERE uuid = ?`
	_, err = db.ExecContext(ctx, cmd, s.UUID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Session) GetUserBySession(ctx context.Context, db *sql.DB) (user User, err error) {
	user = User{}
	cmd := `SELECT id, uuid, name, email, created_at FROM users WHERE id = ?`
	if err = db.QueryRowContext(ctx, cmd, s.UserID).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	); err != nil {
		log.Println(err)
	}
	return user, nil
}
