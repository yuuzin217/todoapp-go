package controllers

import (
	"log"
	"net/http"
	"todo_app/app/models"

	"golang.org/x/crypto/bcrypt"
)

/*
signup は 新規ユーザー登録フォームの表示(GET)と、フォームデータの処理(POST)を行います。
*/
func signup(env *Env, w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		_, err := env.checkSession(w, r)
		if err != nil {
			env.generateHTML(w, r, nil, "layout", "public_navbar", "signup")
		} else {
			http.Redirect(w, r, "/todos", MovedPermanently)
		}
	} else if r.Method == "POST" {
		// 入力フォームの解析
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user := models.User{
			// value属性から値を取得
			Name:     r.PostFormValue("name"),
			Email:    r.PostFormValue("email"),
			Password: r.PostFormValue("password"),
		}
		if err := user.CreateUser(r.Context(), env.DB); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", MovedPermanently)
	}
}

/*
login は ログインフォームを表示するハンドラーです。
すでにセッションが存在する場合はTODO一覧へリダイレクトします。
*/
func login(env *Env, w http.ResponseWriter, r *http.Request) {
	_, err := env.checkSession(w, r)
	if err != nil {
		env.generateHTML(w, r, nil, "layout", "public_navbar", "login")
	} else {
		http.Redirect(w, r, "/todos", MovedTemporarily)
	}
}

/*
authenticate は ログインフォームから送信されたメールアドレスとパスワードを検証し、
認証に成功した場合は新しいセッションを作成してクッキーに保存します。
*/
func authenticate(env *Env, w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := models.GetUserByEmailOrName(r.Context(), env.DB, r.PostFormValue("identifier"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", MovedPermanently)
		return
	}
	// パスワード整合チェック
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.PostFormValue("password")))
	if err == nil {
		// セッション作成
		session, err := user.CreateSession(r.Context(), env.DB)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cookie := http.Cookie{
			Name:     "_cookie",    // Key
			Value:    session.UUID, // Value
			Secure:   true,         // https通信のみcookie送信、インジェクション対策
			HttpOnly: true,         // 参照操作権限をhttpアクセスのみに限定、JavaScriptからの参照防止
		}
		// クッキーを設定
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", MovedPermanently)
	} else {
		http.Redirect(w, r, "/login", MovedPermanently)
	}
}

/*
logout は ユーザーのセッションを破棄し、ログアウト状態にします。
データベース上のセッションレコードも削除します。
*/
func logout(env *Env, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_cookie")
	if err != nil && err != http.ErrNoCookie {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err != http.ErrNoCookie {
		session := models.Session{UUID: cookie.Value}
		if err := session.DeleteSessionByUUID(r.Context(), env.DB); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/login", MovedPermanently)
}
