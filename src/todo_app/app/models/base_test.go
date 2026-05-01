package models

import (
	"testing"
)

// TestGetUUID は UUID 生成関数が正常に文字列を返すことを検証します。
// テストコードから内部関数を検証するためにエクスポートされた GetUUID を利用しています。
func TestGetUUID(t *testing.T) {
	uuid := GetUUID()
	if uuid.String() == "" {
		t.Error("Generated UUID string is empty")
	}
}
