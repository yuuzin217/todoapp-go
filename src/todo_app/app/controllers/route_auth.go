package controllers

import (
	"log"
	"net/http"
	"todo_app/app/models"

	"golang.org/x/crypto/bcrypt"
)

// signup は新規ユーザー登録処理を担当します。
// GET リクエスト時は登録フォームを表示し、POST リクエスト時は送信されたデータに基づいてユーザーを作成します。
func signup(env *Env, w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// すでにログインしているユーザーが登録画面にアクセスした場合、タスク一覧へ逃がします。
		_, err := env.checkSession(w, r)
		if err != nil {
			env.generateHTML(w, r, nil, "layout", "public_navbar", "signup")
		} else {
			http.Redirect(w, r, "/todos", MovedTemporarily)
		}
	} else if r.Method == "POST" {
		// フォームデータのパース。
		if err := r.ParseForm(); err != nil {
			log.Printf("Failed to parse signup form: %v", err)
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		user := models.User{
			Name:     r.PostFormValue("name"),
			Email:    r.PostFormValue("email"),
			Password: r.PostFormValue("password"),
		}

		// データベースへのユーザー登録。パスワードのハッシュ化は内部で行われます。
		if err := user.CreateUser(r.Context(), env.DB); err != nil {
			log.Printf("User registration failed: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// ユーザー体験（UX）向上のため、登録完了後にログイン画面へ戻さず、そのまま自動ログインさせます。
		// そのため、作成されたユーザー情報を再取得してセッションを発行します。
		registeredUser, err := models.GetUserByEmail(r.Context(), env.DB, user.Email)
		if err != nil {
			log.Printf("Failed to retrieve user for auto-login: %v", err)
			http.Redirect(w, r, "/login", MovedTemporarily)
			return
		}

		session, err := registeredUser.CreateSession(r.Context(), env.DB)
		if err != nil {
			log.Printf("Session creation failed: %v", err)
			http.Error(w, "Internal server error during login", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     "_cookie",
			Value:    session.UUID,
			Path:     "/",
			HttpOnly: true,
			// ローカル開発環境(HTTP)での動作を妨げないよう、本番環境のみ Secure 属性を付与します。
			Secure: env.Config.Env == "production",
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", MovedTemporarily)
	}
}

// login はログイン画面を表示します。
// ログイン済みであれば、タスク一覧画面へリダイレクトします。
func login(env *Env, w http.ResponseWriter, r *http.Request) {
	_, err := env.checkSession(w, r)
	if err != nil {
		env.generateHTML(w, r, nil, "layout", "public_navbar", "login")
	} else {
		// ログイン済みであることを示すため、一時的リダイレクトを使用します。
		http.Redirect(w, r, "/todos", MovedTemporarily)
	}
}

// authenticate はログインフォームから送信された識別子（Eメールまたはユーザー名）とパスワードを検証します。
// 成功した場合は新規セッションを発行し、クッキーを設定します。
func authenticate(env *Env, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Failed to parse login form: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// ユーザー名でもメールアドレスでもログインできるように、共通の識別子（identifier）として扱います。
	user, err := models.GetUserByEmailOrName(r.Context(), env.DB, r.PostFormValue("identifier"))
	if err != nil {
		log.Printf("User not found: %s", r.PostFormValue("identifier"))
		http.Redirect(w, r, "/login", MovedTemporarily)
		return
	}

	// 保存されているハッシュ化パスワードと、入力された平文パスワードを照合します。
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.PostFormValue("password"))); err == nil {
		session, err := user.CreateSession(r.Context(), env.DB)
		if err != nil {
			log.Printf("Login session creation failed: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     "_cookie",
			Value:    session.UUID,
			Path:     "/",
			HttpOnly: true,
			Secure:   env.Config.Env == "production",
		}
		http.SetCookie(w, &cookie)
		// トップページへリダイレクト。トップのハンドラーがさらに /todos へ誘導します。
		http.Redirect(w, r, "/", MovedTemporarily)
	} else {
		log.Printf("Password mismatch for user: %s", user.Email)
		http.Redirect(w, r, "/login", MovedTemporarily)
	}
}

// logout は現在のセッションをサーバー側で破棄し、ブラウザのクッキーも無効化します。
func logout(env *Env, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_cookie")
	if err != nil {
		// クッキーがない場合はすでにログアウト状態とみなします。
		http.Redirect(w, r, "/login", MovedTemporarily)
		return
	}

	session := models.Session{UUID: cookie.Value}
	// サーバー側のセッションデータを削除します。
	if err := session.DeleteSessionByUUID(r.Context(), env.DB); err != nil {
		log.Printf("Failed to delete session during logout: %v", err)
	}

	// ブラウザ側のクッキーを即座に無効化するため、有効期限を過去（MaxAge = -1）にセットして再送します。
	cookie.MaxAge = -1
	cookie.Path = "/"
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/login", MovedTemporarily)
}
