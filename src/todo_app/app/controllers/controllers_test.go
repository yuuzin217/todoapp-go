package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"todo_app/app/models"
	"todo_app/config"

	_ "github.com/mattn/go-sqlite3"
)

var testEnv *Env

func TestMain(m *testing.M) {
	// テスト用DBセットアップ
	db, _ := sql.Open("sqlite3", ":memory:")
	models.CreateTables(db)

	testEnv = &Env{
		DB:     db,
		Config: &config.ConfigList{Env: "development", Port: "8080", Static: "app/views"},
	}

	// 翻訳ファイルのロード (ダミーディレクトリ作成が必要)
	os.MkdirAll("app/views/i18n", 0755)
	os.WriteFile("app/views/i18n/en.json", []byte(`{"Welcome":"Welcome"}`), 0644)
	os.WriteFile("app/views/i18n/ja.json", []byte(`{"Welcome":"ようこそ"}`), 0644)
	testEnv.LoadTranslations()

	code := m.Run()

	os.RemoveAll("app")
	os.Exit(code)
}

func TestTop(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	// generateHTMLが実ファイルを探しに行くので、ディレクトリ構造が必要
	os.MkdirAll("app/views/templates", 0755)
	os.WriteFile("app/views/templates/layout.html", []byte(`{{define "layout"}}{{template "navbar" .}}{{template "content" .}}{{end}}`), 0644)
	os.WriteFile("app/views/templates/public_navbar.html", []byte(`{{define "navbar"}}Navbar{{end}}`), 0644)
	os.WriteFile("app/views/templates/top.html", []byte(`{{define "content"}}Top{{end}}`), 0644)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		top(testEnv, w, r)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Top (Logged in)
	u := models.User{Name: "topuser", Email: "top@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)
	user, _ := models.GetUserByEmail(context.Background(), testEnv.DB, u.Email)
	session, _ := user.CreateSession(context.Background(), testEnv.DB)
	req, _ = http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "_cookie", Value: session.UUID})
	rr = httptest.NewRecorder()
	top(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Top logged in should redirect, got %d", rr.Code)
	}
}

func TestGenerateHTMLCache(t *testing.T) {
	// Setup production env
	originalEnv := testEnv.Config.Env
	testEnv.Config.Env = "production"
	defer func() { testEnv.Config.Env = originalEnv }()

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	// First call - should parse and cache
	testEnv.generateHTML(rr, req, nil, "layout", "public_navbar", "top")
	if rr.Code != http.StatusOK {
		t.Errorf("First call failed: %d", rr.Code)
	}

	// Second call - should use cache
	rr2 := httptest.NewRecorder()
	testEnv.generateHTML(rr2, req, nil, "layout", "public_navbar", "top")
	if rr2.Code != http.StatusOK {
		t.Errorf("Second call (cached) failed: %d", rr2.Code)
	}
}

func TestSignup(t *testing.T) {
	// Signup GET
	req, _ := http.NewRequest("GET", "/signup", nil)
	rr := httptest.NewRecorder()
	os.WriteFile("app/views/templates/signup.html", []byte(`{{define "content"}}Signup{{end}}`), 0644)
	
	signup(testEnv, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Signup GET failed: %d", rr.Code)
	}

	// Signup GET (Logged in)
	u := models.User{Name: "signupuser", Email: "signup@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)
	user, _ := models.GetUserByEmail(context.Background(), testEnv.DB, u.Email)
	session, _ := user.CreateSession(context.Background(), testEnv.DB)
	req, _ = http.NewRequest("GET", "/signup", nil)
	req.AddCookie(&http.Cookie{Name: "_cookie", Value: session.UUID})
	rr = httptest.NewRecorder()
	signup(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Signup logged in should redirect, got %d", rr.Code)
	}

	// Signup POST
	data := "name=newuser&email=new@example.com&password=password"
	req, _ = http.NewRequest("POST", "/signup", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	
	signup(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Signup POST failed: %d", rr.Code)
	}
	
	// Check if user is created and session cookie is set
	if cookie := rr.Result().Cookies(); len(cookie) == 0 || cookie[0].Name != "_cookie" {
		t.Error("Signup POST did not set session cookie")
	}
}

func TestAuthenticate(t *testing.T) {
	// Pre-create user
	u := models.User{Name: "authuser", Email: "auth@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)

	// Authenticate Success
	data := "identifier=auth@example.com&password=password"
	req, _ := http.NewRequest("POST", "/authenticate", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	
	authenticate(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Authenticate success should redirect, got %d", rr.Code)
	}

	// Authenticate Failure
	data = "identifier=auth@example.com&password=wrong"
	req, _ = http.NewRequest("POST", "/authenticate", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	
	authenticate(testEnv, rr, req)
	if rr.Code != http.StatusFound { // Redirects back to /login
		location := rr.Header().Get("Location")
		if location != "/login" {
			t.Errorf("Authenticate failure should redirect to /login, got %s", location)
		}
	}
}

func TestTodoFlow(t *testing.T) {
	// 1. Setup User and Session
	u := models.User{Name: "todouser", Email: "todo@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)
	user, _ := models.GetUserByEmail(context.Background(), testEnv.DB, u.Email)
	session, _ := user.CreateSession(context.Background(), testEnv.DB)
	cookie := &http.Cookie{Name: "_cookie", Value: session.UUID}

	// 2. Index (GET /todos)
	req, _ := http.NewRequest("GET", "/todos", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	os.WriteFile("app/views/templates/private_navbar.html", []byte(`{{define "navbar"}}PrivateNavbar{{end}}`), 0644)
	os.WriteFile("app/views/templates/index.html", []byte(`{{define "content"}}Index{{range .Todos}}{{.Content}}{{end}}{{end}}`), 0644)
	
	index(testEnv, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Index failed: %d", rr.Code)
	}

	// 3. New (GET /todos/new)
	req, _ = http.NewRequest("GET", "/todos/new", nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	os.WriteFile("app/views/templates/todo_new.html", []byte(`{{define "content"}}NewTodo{{end}}`), 0644)
	todoNew(testEnv, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("TodoNew failed: %d", rr.Code)
	}

	// 4. Save (POST /todos/save)
	data := "content=TestTask"
	req, _ = http.NewRequest("POST", "/todos/save", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	todoSave(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("TodoSave failed: %d", rr.Code)
	}

	// 5. Edit (GET /todos/edit/1)
	todos, _ := user.GetTodosByUser(context.Background(), testEnv.DB)
	todoID := todos[0].ID
	req, _ = http.NewRequest("GET", fmt.Sprintf("/todos/edit/%d", todoID), nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	os.WriteFile("app/views/templates/todo_edit.html", []byte(`{{define "content"}}EditTodo{{.Content}}{{end}}`), 0644)
	todoEdit(testEnv, rr, req, todoID)
	if rr.Code != http.StatusOK {
		t.Errorf("TodoEdit failed: %d", rr.Code)
	}

	// 6. Update (POST /todos/update/1)
	data = "content=UpdatedTask"
	req, _ = http.NewRequest("POST", fmt.Sprintf("/todos/update/%d", todoID), strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	todoUpdate(testEnv, rr, req, todoID)
	if rr.Code != http.StatusFound {
		t.Errorf("TodoUpdate failed: %d", rr.Code)
	}

	// 7. Delete (GET /todos/delete/1)
	req, _ = http.NewRequest("GET", fmt.Sprintf("/todos/delete/%d", todoID), nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	todoDelete(testEnv, rr, req, todoID)
	if rr.Code != http.StatusFound {
		t.Errorf("TodoDelete failed: %d", rr.Code)
	}
}

func TestI18n(t *testing.T) {
	// Test setLang
	req, _ := http.NewRequest("GET", "/set-lang?l=ja", nil)
	rr := httptest.NewRecorder()
	setLang(testEnv, rr, req)
	if cookie := rr.Result().Cookies(); len(cookie) == 0 || cookie[0].Value != "ja" {
		t.Errorf("setLang failed to set cookie to ja, got %v", cookie)
	}

	// Test getLang with Cookie
	req, _ = http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "ja"})
	lang := testEnv.getLang(req)
	if lang != "ja" {
		t.Errorf("getLang with cookie failed: expected ja, got %s", lang)
	}

	// Test getLang with Header
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8")
	lang = testEnv.getLang(req)
	if lang != "ja" {
		t.Errorf("getLang with header failed: expected ja, got %s", lang)
	}
}

func TestLogout(t *testing.T) {
	// Setup user and session
	u := models.User{Name: "logoutuser", Email: "logout@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)
	user, _ := models.GetUserByEmail(context.Background(), testEnv.DB, u.Email)
	session, _ := user.CreateSession(context.Background(), testEnv.DB)
	
	// Logout with session
	req, _ := http.NewRequest("GET", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "_cookie", Value: session.UUID})
	rr := httptest.NewRecorder()
	
	logout(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Logout with session failed: %d", rr.Code)
	}
	
	// Check if session is deleted
	valid, _ := session.CheckSession(context.Background(), testEnv.DB)
	if valid {
		t.Error("Session should be invalid after logout")
	}

	// Logout without session
	req, _ = http.NewRequest("GET", "/logout", nil)
	rr = httptest.NewRecorder()
	logout(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Logout without session failed: %d", rr.Code)
	}
}
