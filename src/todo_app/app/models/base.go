package models

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

/*
定数定義
将来的に参照される可能性を考慮し、テーブル名を定数として定義しています。
現在は SQLインジェクション対策のため CREATE TABLE 文で直接ハードコードしています。
*/
const (
	tableNameUser    = "users"
	tableNameTodo    = "todos"
	tableNameSession = "sessions"
)

// CreateTables はアプリケーションで必要なデータベーステーブル (users, todos, sessions) を作成します。
// 既にテーブルが存在する場合は作成をスキップします (IF NOT EXISTS)。
func CreateTables(db *sql.DB) {
	// 外部キー制約を有効化 (接続ごとに必要)
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalln(err)
	}

	// トランザクションの開始
	tx, err := db.Begin()
	if err != nil {
		log.Fatalln(err)
	}

	// エラー発生時にロールバックするように遅延実行を設定
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Fatalf("Transaction failed, rolling back: %v", err)
		}
	}()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			name TEXT,
			email TEXT UNIQUE,
			password TEXT,
			created_at DATETIME)`,

		`CREATE TABLE IF NOT EXISTS todos(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT,
			user_id INTEGER,
			created_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE)`,

		`CREATE TABLE IF NOT EXISTS sessions(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			email TEXT,
			user_id INTEGER,
			created_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE)`,
	}

	for _, q := range queries {
		if _, err = tx.Exec(q); err != nil {
			return // defer内のRollbackが呼ばれる
		}
	}

	// すべて成功したらコミット
	if err = tx.Commit(); err != nil {
		return
	}
}

/*
createUUID は UUID を作成します。
*/
func createUUID() (uuidobj uuid.UUID) {
	uuidobj, _ = uuid.NewUUID()
	return uuidobj
}

// GetUUID はテスト用に公開された UUID 生成関数です
func GetUUID() uuid.UUID {
	return createUUID()
}

