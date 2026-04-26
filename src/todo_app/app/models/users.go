package models

import (
	"context"
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User はデータベースの users テーブルのレコードを表す構造体です。
type User struct {
	ID        int       // ユーザーの一意なID
	UUID      string    // セキュリティや外部参照用のUUID
	Name      string    // ユーザー名
	Email     string    // Eメールアドレス (ログインに使用)
	Password  string    // ハッシュ化されたパスワード
	CreatedAt time.Time // 作成日時
	Todos     []Todo    // ユーザーに紐づくTODOリストのキャッシュ
}

// Session はデータベースの sessions テーブルのレコードを表し、ログイン状態を管理します。
type Session struct {
	ID        int       // セッションの一意なID
	UUID      string    // クッキーとしてブラウザに保存されるUUID
	Email     string    // セッションに紐づくユーザーのEメール
	UserID    int       // セッションに紐づくユーザーID
	CreatedAt time.Time // セッション作成日時
}

// CreateUser は新しいユーザーをデータベースに登録します。
// パスワードは保存前にbcryptでハッシュ化されます。
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

// GetUser は指定されたIDのユーザーをデータベースから取得します。
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

// UpdateUser はユーザー情報(名前とEメール)を更新します。
func (u *User) UpdateUser(ctx context.Context, db *sql.DB) (err error) {
	cmd := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	_, err = db.ExecContext(ctx, cmd, u.Name, u.Email, u.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

// DeleteUser はユーザーをデータベースから削除します。
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

// CreateSession はログインに成功したユーザーのために新しいセッションを発行してデータベースに保存します。
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

// DeleteSessionByUUID は指定されたUUIDのセッションをデータベースから削除します（ログアウト処理）。
func (s *Session) DeleteSessionByUUID(ctx context.Context, db *sql.DB) (err error) {
	cmd := `DELETE FROM sessions WHERE uuid = ?`
	_, err = db.ExecContext(ctx, cmd, s.UUID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetUserBySession はセッション情報から紐づくユーザー情報を取得します。
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
