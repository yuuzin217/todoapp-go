package config

import (
	"os"
	"testing"
)

// TestLoadConfig は設定ファイル (config.ini) および環境変数からの設定読み込みを検証します。
func TestLoadConfig(t *testing.T) {
	// テスト用のダミー設定ファイルを作成します。
	// テスト実行環境に依存せず、一貫した期待値でテストできるようにするためです。
	content := []byte("[web]\nPort = 8888\nlogfile = test.log\nstatic = test_static\nenv = development\n\n[db]\ndriver = sqlite3\nname = test.sql")
	err := os.WriteFile("config.ini", content, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("config.ini")

	// 環境変数が設定ファイルの設定を上書きできるか検証するため、一時的にセットします。
	os.Setenv("GO_ENV", "test")
	defer os.Unsetenv("GO_ENV")

	cfg := LoadConfig()

	if cfg.Port != "8888" {
		t.Errorf("Expected Port 8888, but got %s", cfg.Port)
	}
	if cfg.Env != "test" {
		t.Errorf("Expected Env 'test', but got %s", cfg.Env)
	}
	if cfg.LogFile != "test.log" {
		t.Errorf("Expected LogFile 'test.log', but got %s", cfg.LogFile)
	}
}
