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
	cmdU := `CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid STRING NOT NULL UNIQUE,
		name STRING,
		email STRING,
		password STRING,
		created_at DATETIME)`

	if _, err := db.Exec(cmdU); err != nil {
		log.Fatalln(err)
	}

	cmdT := `CREATE TABLE IF NOT EXISTS todos(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT,
		user_id INTEGER,
		created_at DATETIME)`

	if _, err := db.Exec(cmdT); err != nil {
		log.Fatalln(err)
	}

	cmdS := `CREATE TABLE IF NOT EXISTS sessions(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid STRING NOT NULL UNIQUE,
		email STRING,
		user_id INTEGER,
		created_at DATETIME)`

	if _, err := db.Exec(cmdS); err != nil {
		log.Fatalln(err)
	}
}

/*
createUUID は UUID を作成します。
*/
func createUUID() (uuidobj uuid.UUID) {
	uuidobj, _ = uuid.NewUUID()
	return uuidobj
}

