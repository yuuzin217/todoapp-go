package main

import (
	"database/sql"
	"log"
	"todo_app/app/controllers"
	"todo_app/app/models"
	"todo_app/config"
	"todo_app/utils"
)

func main() {
	cfg := config.LoadConfig()
	utils.LoggingSettings(cfg.LogFile)

	db, err := sql.Open(cfg.SQLDriver, cfg.DBName)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	models.CreateTables(db)

	env := &controllers.Env{
		DB:     db,
		Config: cfg,
	}

	controllers.StartMainServer(env)
}
