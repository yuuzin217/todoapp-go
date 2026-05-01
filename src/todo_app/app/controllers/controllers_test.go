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

// TestMain は controllers パッケージ全体のテストの前処理と後処理を制御します。
// インメモリの SQLite データベースをセットアップし、テストに必要な依存関係 (Env) を初期化します。
func TestMain(m *testing.M) {
	// インメモリ DB を使用することで、ファイル I/O を排除しテストを高速化します。
	db, _ := sql.Open("sqlite3", ":memory:")
	models.CreateTables(db)

	testEnv = &Env{
		DB:     db,
		Config: &config.ConfigList{Env: "development", Port: "8080", Static: "app/views"},
	}

	// 翻訳ファイルのロード機能をテストするため、一時的なディレクトリ構造とダミーファイルを作成します。
	os.MkdirAll("app/views/i18n", 0755)
	os.WriteFile("app/views/i18n/en.json", []byte(`{"Welcome":"Welcome"}`), 0644)
	os.WriteFile("app/views/i18n/ja.json", []byte(`{"Welcome":"ようこそ"}`), 0644)
	testEnv.LoadTranslations()

	code := m.Run()

	// テスト終了後は、作成した一時的なリソースをクリーンアップして副作用を防ぎます。
	os.RemoveAll("app")
	os.Exit(code)
}

// TestTop はルートパスへのアクセスをテストします。
// 未ログイン時の紹介ページ表示と、ログイン時のタスク一覧へのリダイレクトの両方のパスを検証します。
func TestTop(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	// generateHTML は物理的なテンプレートファイルを探しに行くため、テスト用のダミーテンプレートを用意します。
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

	// ログイン済みユーザーがルートにアクセスした場合のリダイレクト処理を検証。
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

// TestGenerateHTMLCache は本番環境におけるテンプレートキャッシュ機能をテストします。
// 2回目以降の呼び出しでパース処理がスキップされ、キャッシュが利用されるパスを通過させることを目的としています。
func TestGenerateHTMLCache(t *testing.T) {
	// 一時的に本番環境設定を偽装します。
	originalEnv := testEnv.Config.Env
	testEnv.Config.Env = "production"
	defer func() { testEnv.Config.Env = originalEnv }()

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	// 1回目の呼び出し: パースとキャッシュへの格納。
	testEnv.generateHTML(rr, req, nil, "layout", "public_navbar", "top")
	if rr.Code != http.StatusOK {
		t.Errorf("First call failed: %d", rr.Code)
	}

	// 2回目の呼び出し: キャッシュからの取得。
	rr2 := httptest.NewRecorder()
	testEnv.generateHTML(rr2, req, nil, "layout", "public_navbar", "top")
	if rr2.Code != http.StatusOK {
		t.Errorf("Second call (cached) failed: %d", rr2.Code)
	}
}

// TestSignup は新規ユーザー登録処理をテストします。
func TestSignup(t *testing.T) {
	// GET リクエストによるフォーム表示。
	req, _ := http.NewRequest("GET", "/signup", nil)
	rr := httptest.NewRecorder()
	os.WriteFile("app/views/templates/signup.html", []byte(`{{define "content"}}Signup{{end}}`), 0644)
	
	signup(testEnv, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Signup GET failed: %d", rr.Code)
	}

	// ログイン済みユーザーが登録画面へアクセスした場合のリダイレクトを検証。
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

	// POST リクエストによる新規登録と、登録後の自動ログイン (クッキーセット) を検証。
	data := "name=newuser&email=new@example.com&password=password"
	req, _ = http.NewRequest("POST", "/signup", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	
	signup(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Signup POST failed: %d", rr.Code)
	}
	
	if cookie := rr.Result().Cookies(); len(cookie) == 0 || cookie[0].Name != "_cookie" {
		t.Error("Signup POST did not set session cookie")
	}
}

// TestAuthenticate はログイン認証処理をテストします。
func TestAuthenticate(t *testing.T) {
	// 認証対象のユーザーをあらかじめ作成。
	u := models.User{Name: "authuser", Email: "auth@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)

	// 正しい資格情報での認証成功。
	data := "identifier=auth@example.com&password=password"
	req, _ := http.NewRequest("POST", "/authenticate", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	
	authenticate(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Authenticate success should redirect, got %d", rr.Code)
	}

	// 誤ったパスワードによる認証失敗と、ログイン画面への差し戻しを検証。
	data = "identifier=auth@example.com&password=wrong"
	req, _ = http.NewRequest("POST", "/authenticate", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	
	authenticate(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		location := rr.Header().Get("Location")
		if location != "/login" {
			t.Errorf("Authenticate failure should redirect to /login, got %s", location)
		}
	}
}

// TestTodoFlow は TODO の作成、表示、編集、削除の一連の流れ (CRUD) をテストします。
func TestTodoFlow(t *testing.T) {
	// 1. テストユーザーとセッションの準備
	u := models.User{Name: "todouser", Email: "todo@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)
	user, _ := models.GetUserByEmail(context.Background(), testEnv.DB, u.Email)
	session, _ := user.CreateSession(context.Background(), testEnv.DB)
	cookie := &http.Cookie{Name: "_cookie", Value: session.UUID}

	// 2. TODO 一覧表示 (index)
	req, _ := http.NewRequest("GET", "/todos", nil)
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	os.WriteFile("app/views/templates/private_navbar.html", []byte(`{{define "navbar"}}PrivateNavbar{{end}}`), 0644)
	os.WriteFile("app/views/templates/index.html", []byte(`{{define "content"}}Index{{range .Todos}}{{.Content}}{{end}}{{end}}`), 0644)
	
	index(testEnv, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Index failed: %d", rr.Code)
	}

	// 3. 新規作成フォーム表示 (todoNew)
	req, _ = http.NewRequest("GET", "/todos/new", nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	os.WriteFile("app/views/templates/todo_new.html", []byte(`{{define "content"}}NewTodo{{end}}`), 0644)
	todoNew(testEnv, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("TodoNew failed: %d", rr.Code)
	}

	// 4. TODO の保存 (todoSave)
	data := "content=TestTask"
	req, _ = http.NewRequest("POST", "/todos/save", strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	todoSave(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("TodoSave failed: %d", rr.Code)
	}

	// 5. 編集フォーム表示 (todoEdit)
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

	// 6. TODO の更新 (todoUpdate)
	data = "content=UpdatedTask"
	req, _ = http.NewRequest("POST", fmt.Sprintf("/todos/update/%d", todoID), strings.NewReader(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	todoUpdate(testEnv, rr, req, todoID)
	if rr.Code != http.StatusFound {
		t.Errorf("TodoUpdate failed: %d", rr.Code)
	}

	// 7. TODO の削除 (todoDelete)
	req, _ = http.NewRequest("GET", fmt.Sprintf("/todos/delete/%d", todoID), nil)
	req.AddCookie(cookie)
	rr = httptest.NewRecorder()
	todoDelete(testEnv, rr, req, todoID)
	if rr.Code != http.StatusFound {
		t.Errorf("TodoDelete failed: %d", rr.Code)
	}
}

// TestI18n は多言語対応機能をテストします。
func TestI18n(t *testing.T) {
	// 言語設定の変更 (set-lang) を検証。
	req, _ := http.NewRequest("GET", "/set-lang?l=ja", nil)
	rr := httptest.NewRecorder()
	setLang(testEnv, rr, req)
	if cookie := rr.Result().Cookies(); len(cookie) == 0 || cookie[0].Value != "ja" {
		t.Errorf("setLang failed to set cookie to ja, got %v", cookie)
	}

	// クッキーに基づく言語判定を検証。
	req, _ = http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "ja"})
	lang := testEnv.getLang(req)
	if lang != "ja" {
		t.Errorf("getLang with cookie failed: expected ja, got %s", lang)
	}

	// Accept-Language ヘッダーに基づく言語判定を検証。
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8")
	lang = testEnv.getLang(req)
	if lang != "ja" {
		t.Errorf("getLang with header failed: expected ja, got %s", lang)
	}
}

// TestLogout はログアウト処理をテストします。
func TestLogout(t *testing.T) {
	// セッションが存在する場合のログアウトと、クッキーの無効化を検証。
	u := models.User{Name: "logoutuser", Email: "logout@example.com", Password: "password"}
	u.CreateUser(context.Background(), testEnv.DB)
	user, _ := models.GetUserByEmail(context.Background(), testEnv.DB, u.Email)
	session, _ := user.CreateSession(context.Background(), testEnv.DB)
	
	req, _ := http.NewRequest("GET", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "_cookie", Value: session.UUID})
	rr := httptest.NewRecorder()
	
	logout(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Logout with session failed: %d", rr.Code)
	}
	
	// セッションが実際に DB から削除されたことを確認。
	valid, _ := session.CheckSession(context.Background(), testEnv.DB)
	if valid {
		t.Error("Session should be invalid after logout")
	}

	// すでにログアウトしている状態で再度ログアウトを試みた場合の安全性を検証。
	req, _ = http.NewRequest("GET", "/logout", nil)
	rr = httptest.NewRecorder()
	logout(testEnv, rr, req)
	if rr.Code != http.StatusFound {
		t.Errorf("Logout without session failed: %d", rr.Code)
	}
}
