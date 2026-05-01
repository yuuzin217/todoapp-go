package models

import (
	"testing"
)

func TestGetUUID(t *testing.T) {
	uuid := GetUUID()
	if uuid.String() == "" {
		t.Error("Generated UUID string is empty")
	}
}
