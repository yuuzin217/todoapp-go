package models

import (
	"context"
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User はデータベースの users テーブルのレコードを表す構造体です。
// アプリケーション内のユーザー認証やプロファイル情報の基盤となります。
type User struct {
	ID        int       // ユーザーの一意なID (プライマリキー)
	UUID      string    // セキュリティや外部参照用の不変な識別子
	Name      string    // ユーザー名 (ログイン識別子としても利用可能)
	Email     string    // Eメールアドレス (主要なログイン識別子)
	Password  string    // ハッシュ化されたパスワード。平文は保持しません。
	CreatedAt time.Time // アカウントの作成日時
	Todos     []Todo    // ユーザーに紐づくTODOリストのキャッシュ用スライス
}

// Session はデータベースの sessions テーブルのレコードを表し、ログイン状態を保持します。
// クッキーに保存された UUID と紐づけて認証状態を管理します。
type Session struct {
	ID        int       // セッションの一意なID
	UUID      string    // ブラウザのクッキー (_cookie) に保存される公開用識別子
	Email     string    // セッション作成時のユーザーのEメール。冗長ですが検索効率のために保持。
	UserID    int       // セッションに紐づくユーザーの内部ID
	CreatedAt time.Time // セッションが発行された日時
}

// CreateUser は新しいユーザーをデータベースに登録します。
// セキュリティのため、渡されたパスワードは必ず bcrypt でハッシュ化してから保存されます。
func (u *User) CreateUser(ctx context.Context, db *sql.DB) (err error) {
	// パスワードのハッシュ化。コスト値は標準的な DefaultCost を使用。
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return err
	}

	cmd := `INSERT INTO users(uuid, name, email, password, created_at) values (?, ?, ?, ?, ?)`
	_, err = db.ExecContext(
		ctx,
		cmd,
		createUUID(), // 内部IDとは別に外部公開用の不変なIDを付与
		u.Name,
		u.Email,
		string(hash),
		time.Now(),
	)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
	}
	return err
}

// GetUser は内部的な ID を指定して、一人のユーザー情報を取得します。
// 内部的なリレーションの解決（例: TODO の所有者確認）などで使用されます。
func GetUser(ctx context.Context, db *sql.DB, id int) (user User, err error) {
	user = User{}
	cmd := `SELECT id, uuid, name, email, password, created_at FROM users WHERE id = ?`
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

// UpdateUser は指定されたユーザーの基本情報（名前とEメール）を更新します。
// 現時点ではパスワードの更新は別ルートを想定しているため含まれません。
func (u *User) UpdateUser(ctx context.Context, db *sql.DB) (err error) {
	cmd := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	_, err = db.ExecContext(ctx, cmd, u.Name, u.Email, u.ID)
	if err != nil {
		log.Printf("Failed to update user %d: %v", u.ID, err)
	}
	return err
}

// DeleteUser はユーザー情報を物理削除します。
// データベース側の ON DELETE CASCADE 制約により、紐づく TODO やセッションも自動的に削除されます。
func (u *User) DeleteUser(ctx context.Context, db *sql.DB) (err error) {
	cmd := `DELETE FROM users WHERE id = ?`
	_, err = db.ExecContext(ctx, cmd, u.ID)
	if err != nil {
		log.Printf("Failed to delete user %d: %v", u.ID, err)
	}
	return err
}

// GetUserByEmail は Eメールアドレスに完全一致するユーザーを取得します。
// ユーザー登録時の重複チェックや、初期の認証フローで使用されます。
func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (user User, err error) {
	user = User{}
	cmd := `SELECT id, uuid, name, email, password, created_at FROM users WHERE email = ?`
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

// GetUserByEmailOrName は Eメールアドレスまたはユーザー名のいずれかに一致するユーザーを取得します。
// ユーザーがどちらの識別子でもログインできるようにするために導入されました。
func GetUserByEmailOrName(ctx context.Context, db *sql.DB, identifier string) (user User, err error) {
	user = User{}
	// 同一の入力値(identifier)を email と name 両方のカラムに対して検索します。
	cmd := `SELECT id, uuid, name, email, password, created_at FROM users WHERE email = ? OR name = ?`
	err = db.QueryRowContext(ctx, cmd, identifier, identifier).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	return user, err
}

// CreateSession はユーザーに対して新しいログインセッションを発行します。
// 発行されたセッションの UUID は、ブラウザのクッキーに保存して次回以降の認証に使用します。
func (u *User) CreateSession(ctx context.Context, db *sql.DB) (session Session, err error) {
	session = Session{}
	cmd1 := `INSERT INTO sessions (uuid, email, user_id, created_at) VALUES (?, ?, ?, ?)`
	uuid := createUUID()
	_, err = db.ExecContext(ctx, cmd1, uuid, u.Email, u.ID, time.Now())
	if err != nil {
		log.Printf("Failed to insert session for user %d: %v", u.ID, err)
		return session, err
	}

	// 挿入されたレコードを再度取得して ID 等を確定させます。
	cmd2 := `SELECT id, uuid, email, user_id, created_at FROM sessions WHERE uuid = ?`
	err = db.QueryRowContext(ctx, cmd2, uuid).Scan(
		&session.ID,
		&session.UUID,
		&session.Email,
		&session.UserID,
		&session.CreatedAt,
	)
	if err != nil {
		log.Printf("Failed to retrieve created session: %v", err)
	}

	return session, err
}

// CheckSession は提供されたクッキー値（UUID）が有効なセッションとして存在するか確認します。
// セッションが見つかれば true を、見つからないかエラーが発生すれば false を返します。
func (s *Session) CheckSession(ctx context.Context, db *sql.DB) (valid bool, err error) {
	cmd := `SELECT id, uuid, email, user_id, created_at FROM sessions WHERE uuid = ?`

	err = db.QueryRowContext(ctx, cmd, s.UUID).Scan(
		&s.ID,
		&s.UUID,
		&s.Email,
		&s.UserID,
		&s.CreatedAt)

	if err != nil {
		// データが見つからない場合も Scan はエラー(sql.ErrNoRows)を返すため、
		// 呼び出し側は err の有無で有効性を判断できます。
		return false, err
	}

	return s.ID != 0, nil
}

// DeleteSessionByUUID は指定された UUID に紐づくセッションレコードを削除します。
// 主にログアウト処理で使用され、これによりクッキーが有効でもサーバー側で認可されなくなります。
func (s *Session) DeleteSessionByUUID(ctx context.Context, db *sql.DB) (err error) {
	cmd := `DELETE FROM sessions WHERE uuid = ?`
	_, err = db.ExecContext(ctx, cmd, s.UUID)
	if err != nil {
		log.Printf("Failed to delete session %s: %v", s.UUID, err)
	}
	return err
}

// GetUserBySession は現在のセッションに紐づいているユーザーの情報を取得します。
// 認証済みリクエストにおいて「誰が」リクエストを送っているかを特定するために必須の処理です。
func (s *Session) GetUserBySession(ctx context.Context, db *sql.DB) (user User, err error) {
	user = User{}
	cmd := `SELECT id, uuid, name, email, created_at FROM users WHERE id = ?`
	err = db.QueryRowContext(ctx, cmd, s.UserID).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)
	if err != nil {
		log.Printf("Failed to get user by session: %v", err)
	}
	return user, err
}
