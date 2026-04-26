package models

import (
	"log"
	"time"
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

func (u *User) CreateUser() (err error) {
	cmd := `INSERT INTO users(
		uuid,
		name,
		email,
		password,
		created_at) values (?, ?, ?, ?, ?)`

	_, err = DB.Exec(
		cmd,
		createUUID(),
		u.Name,
		u.Email,
		Encrypt(u.Password),
		time.Now(),
	)
	if err != nil {
		log.Println(err)
	}
	return err
}

func GetUser(id int) (user User, err error) {
	user = User{}
	cmd := `SELECT id,uuid, name, email, password, created_at FROM users WHERE id = ?`
	err = DB.QueryRow(cmd, id).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	return user, err
}

func (u *User) UpdateUser() (err error) {
	cmd := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	_, err = DB.Exec(cmd, u.Name, u.Email, u.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (u *User) DeleteUser() (err error) {
	cmd := `DELETE FROM users WHERE id = ?`
	_, err = DB.Exec(cmd, u.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

/*
GetUserByEmail は Eメールからユーザー情報を取得します。
*/
func GetUserByEmail(email string) (user User, err error) {
	user = User{}
	cmd := `SELECT id, uuid, name, email, password, created_at
	FROM users WHERE email = ?`
	err = DB.QueryRow(cmd, email).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	return user, err
}

func (u *User) CreateSession() (session Session, err error) {
	session = Session{}
	cmd1 := `INSERT INTO sessions (uuid, email, user_id, created_at) VALUES (?, ?, ?, ?)`
	_, err = DB.Exec(cmd1, createUUID(), u.Email, u.ID, time.Now())
	if err != nil {
		log.Println(err)
	}

	cmd2 := `SELECT id, uuid, email, user_id, created_at FROM sessions WHERE user_id = ? AND email = ?`
	if err = DB.QueryRow(cmd2, u.ID, u.Email).Scan(
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
func (s *Session) CheckSession() (valid bool, err error) {
	cmd := `select id, uuid, email, user_id, created_at
	 from sessions where uuid = ?`

	err = DB.QueryRow(cmd, s.UUID).Scan(
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

func (s *Session) DeleteSessionByUUID() (err error) {
	cmd := `DELETE FROM sessions WHERE uuid = ?`
	_, err = DB.Exec(cmd, s.UUID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Session) GetUserBySession() (user User, err error) {
	user = User{}
	cmd := `SELECT id, uuid, name, email, created_at FROM users WHERE id = ?`
	if err = DB.QueryRow(cmd, s.UserID).Scan(
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
