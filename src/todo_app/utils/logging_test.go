package utils

import (
	"os"
	"testing"
)

// TestLoggingSettings はログの初期化設定が正しく行われ、ログファイルが作成されることを検証します。
func TestLoggingSettings(t *testing.T) {
	logFile := "test_webapp.log"
	LoggingSettings(logFile)
	// テスト終了後にログファイルを削除して環境を汚さないようにします。
	defer os.Remove(logFile)

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file %s was not created", logFile)
	}
}
