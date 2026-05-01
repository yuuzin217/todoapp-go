package models

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// テスト用のインメモリDBをセットアップ
	var err error
	testDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer testDB.Close()

	// テーブル作成
	CreateTables(testDB)

	// テスト実行
	code := m.Run()
	os.Exit(code)
}

func TestCreateUser(t *testing.T) {
	u := &User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "password",
	}
	err := u.CreateUser(context.Background(), testDB)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// 重複登録のテスト (エラーになるはず)
	err = u.CreateUser(context.Background(), testDB)
	if err == nil {
		t.Error("Expected error for duplicate email, but got nil")
	}
}

func TestGetUser(t *testing.T) {
	u := &User{
		Name:     "getuser",
		Email:    "get@example.com",
		Password: "password",
	}
	u.CreateUser(context.Background(), testDB)

	fetchedUser, err := GetUserByEmail(context.Background(), testDB, u.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	userByID, err := GetUser(context.Background(), testDB, fetchedUser.ID)
	if err != nil {
		t.Fatalf("GetUser by ID failed: %v", err)
	}

	if userByID.Name != u.Name {
		t.Errorf("Expected name %s, but got %s", u.Name, userByID.Name)
	}

	// 存在しないユーザー
	_, err = GetUser(context.Background(), testDB, 999)
	if err == nil {
		t.Error("Expected error for non-existent user ID, but got nil")
	}
}

func TestUpdateUser(t *testing.T) {
	u := &User{
		Name:     "updateuser",
		Email:    "update@example.com",
		Password: "password",
	}
	u.CreateUser(context.Background(), testDB)
	fetched, _ := GetUserByEmail(context.Background(), testDB, u.Email)

	fetched.Name = "updatedname"
	err := fetched.UpdateUser(context.Background(), testDB)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	updated, _ := GetUser(context.Background(), testDB, fetched.ID)
	if updated.Name != "updatedname" {
		t.Errorf("Expected updated name 'updatedname', but got %s", updated.Name)
	}
}

func TestDeleteUser(t *testing.T) {
	u := &User{
		Name:     "deleteuser",
		Email:    "delete@example.com",
		Password: "password",
	}
	u.CreateUser(context.Background(), testDB)
	fetched, _ := GetUserByEmail(context.Background(), testDB, u.Email)

	err := fetched.DeleteUser(context.Background(), testDB)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	_, err = GetUser(context.Background(), testDB, fetched.ID)
	if err == nil {
		t.Error("Expected error for deleted user, but got nil")
	}
}

func TestGetUserByEmailOrName(t *testing.T) {
	u := &User{
		Name:     "identuser",
		Email:    "ident@example.com",
		Password: "password",
	}
	u.CreateUser(context.Background(), testDB)

	// By Email
	_, err := GetUserByEmailOrName(context.Background(), testDB, "ident@example.com")
	if err != nil {
		t.Errorf("GetUserByEmailOrName by email failed: %v", err)
	}

	// By Name
	_, err = GetUserByEmailOrName(context.Background(), testDB, "identuser")
	if err != nil {
		t.Errorf("GetUserByEmailOrName by name failed: %v", err)
	}

	// Not found
	_, err = GetUserByEmailOrName(context.Background(), testDB, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent identifier, but got nil")
	}
}

func TestSession(t *testing.T) {
	u := &User{
		Name:     "sessuser",
		Email:    "sess@example.com",
		Password: "password",
	}
	u.CreateUser(context.Background(), testDB)
	user, _ := GetUserByEmail(context.Background(), testDB, u.Email)

	// CreateSession
	session, err := user.CreateSession(context.Background(), testDB)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// CheckSession
	valid, err := session.CheckSession(context.Background(), testDB)
	if err != nil || !valid {
		t.Errorf("CheckSession failed: valid=%v, err=%v", valid, err)
	}

	// GetUserBySession
	sessUser, err := session.GetUserBySession(context.Background(), testDB)
	if err != nil {
		t.Fatalf("GetUserBySession failed: %v", err)
	}
	if sessUser.ID != user.ID {
		t.Errorf("Expected user ID %d, but got %d", user.ID, sessUser.ID)
	}

	// DeleteSession
	err = session.DeleteSessionByUUID(context.Background(), testDB)
	if err != nil {
		t.Fatalf("DeleteSessionByUUID failed: %v", err)
	}

	valid, _ = session.CheckSession(context.Background(), testDB)
	if valid {
		t.Error("Expected session to be invalid after deletion")
	}
}
