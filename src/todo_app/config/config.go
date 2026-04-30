package config

import (
	"log"
	"os"

	"gopkg.in/ini.v1"
)

// ConfigList はアプリケーション全体の設定を保持する構造体です。
// config.iniファイルから読み込まれた値が各フィールドに格納されます。
type ConfigList struct {
	Port      string // Webサーバーがリッスンするポート番号
	SQLDriver string // データベースドライバー名 (例: "sqlite3")
	DBName    string // データベースファイル名
	LogFile   string // ログを出力するファイルパス
	Static    string // 静的ファイル (HTML, CSS等) を提供するディレクトリパス
	Env       string // 実行環境 ("development" または "production")
}

// LoadConfig は config.ini ファイルを読み込み、ConfigList構造体を初期化して返します。
// ファイルの読み込みに失敗した場合はログを出力してプログラムを終了します。
func LoadConfig() *ConfigList {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln(err)
	}

	env := os.Getenv("GO_ENV")
	if env == "" {
		env = cfg.Section("web").Key("env").MustString("production")
	}

	return &ConfigList{
		Port:      cfg.Section("web").Key("Port").MustString("8080"),
		SQLDriver: cfg.Section("db").Key("driver").String(),
		DBName:    cfg.Section("db").Key("name").String(),
		LogFile:   cfg.Section("web").Key("logfile").String(),
		Static:    cfg.Section("web").Key("static").String(),
		Env:       env,
	}
}
