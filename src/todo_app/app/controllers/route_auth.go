package controllers

import (
	"log"
	"net/http"
	"todo_app/app/models"
)

/*
signup は ユーザー登録を行います。
*/
func signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		_, err := checkSession(w, r)
		if err != nil {
			generateHTML(w, nil, "layout", "public_navbar", "signup")
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
		if err := user.CreateUser(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", MovedPermanently)
	}
}

/*
login は ログイン処理を行うハンドラーです。
*/
func login(w http.ResponseWriter, r *http.Request) {
	_, err := checkSession(w, r)
	if err != nil {
		generateHTML(w, nil, "layout", "public_navbar", "login")
	} else {
		http.Redirect(w, r, "/todos", MovedTemporarily)
	}
}

/*
authenticate は パスワード認証を行うハンドラーです。
*/
func authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := models.GetUserByEmail(r.PostFormValue("email"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", MovedPermanently)
		return
	}
	// パスワード整合チェック
	if user.Password == models.Encrypt(r.PostFormValue("password")) {
		// セッション作成
		session, err := user.CreateSession()
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
logout は ログアウト処理を行うハンドラーです。
*/
func logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("_cookie")
	if err != nil {
		log.Println(err)
	}
	if err != http.ErrNoCookie {
		session := models.Session{UUID: cookie.Value}
		session.DeleteSessionByUUID()
	}
	http.Redirect(w, r, "/login", MovedPermanently)
}
