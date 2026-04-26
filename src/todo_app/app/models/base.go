package models

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

/*
テーブル名の一蘭
*/
const (
	tableNameUser    = "users"
	tableNameTodo    = "todos"
	tableNameSession = "sessions"
)

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

