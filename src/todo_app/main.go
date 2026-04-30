package main

import (
	"database/sql"
	"html/template"
	"log"
	"sync"
	"todo_app/app/controllers"
	"todo_app/app/models"
	"todo_app/config"
	"todo_app/utils"
)

func main() {
	// 設定ファイルのロードとログ設定の初期化
	cfg := config.LoadConfig()
	utils.LoggingSettings(cfg.LogFile)

	// データベース接続の初期化
	db, err := sql.Open(cfg.SQLDriver, cfg.DBName)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// 必要なデータベーステーブルの作成
	models.CreateTables(db)

	// コントローラー層へ依存(DB, 設定)を注入するためのEnv構造体を初期化
	env := &controllers.Env{
		DB:            db,
		Config:        cfg,
		TemplateCache: make(map[string]*template.Template),
		Mu:            sync.RWMutex{},
	}

	// Webサーバーの起動
	controllers.StartMainServer(env)
}
