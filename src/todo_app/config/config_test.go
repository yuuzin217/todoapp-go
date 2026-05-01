package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 期待される config.ini を作成
	content := []byte("[web]\nPort = 8888\nlogfile = test.log\nstatic = test_static\nenv = development\n\n[db]\ndriver = sqlite3\nname = test.sql")
	err := os.WriteFile("config.ini", content, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("config.ini")

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
