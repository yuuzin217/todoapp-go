package controllers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"todo_app/app/models"
	"todo_app/config"
)

// Env はコントローラー層全体で共有される依存関係 (データベース接続や設定など) を保持する構造体です。
// これにより、グローバル変数を使わずに各ハンドラーへ依存を注入(Dependency Injection)できます。
type Env struct {
	DB            *sql.DB                       // データベースコネクション
	Config        *config.ConfigList            // アプリケーション設定
	TemplateCache map[string]*template.Template // テンプレートキャッシュ (本番用)
	Mu            sync.RWMutex                  // キャッシュ操作用ミューテックス
}

/*
generateHTML は 指定されたテンプレートファイル群をパースしてHTMLを生成し、レスポンスとして書き込みます。
本番環境 (production) ではキャッシュを使用し、開発環境 (development) では毎回パースを行います。
*/
func (env *Env) generateHTML(w http.ResponseWriter, data interface{}, fileNames ...string) {
	var files []string
	for _, file := range fileNames {
		files = append(files, fmt.Sprintf("app/views/templates/%s.html", file))
	}

	key := strings.Join(fileNames, ",")

	// 本番環境かつキャッシュがある場合はキャッシュを使用
	if env.Config.Env == "production" {
		env.Mu.RLock()
		t, ok := env.TemplateCache[key]
		env.Mu.RUnlock()
		if ok {
			t.ExecuteTemplate(w, "layout", data)
			return
		}
	}

	// テンプレートのパース
	templates := template.Must(template.ParseFiles(files...))

	// 本番環境の場合はキャッシュに保存
	if env.Config.Env == "production" {
		env.Mu.Lock()
		if env.TemplateCache == nil {
			env.TemplateCache = make(map[string]*template.Template)
		}
		env.TemplateCache[key] = templates
		env.Mu.Unlock()
	}

	templates.ExecuteTemplate(w, "layout", data)
}

/*
checkSession は セッションの確認をおこないます。
*/
func (env *Env) checkSession(w http.ResponseWriter, r *http.Request) (session models.Session, err error) {
	cookie, err := r.Cookie("_cookie")
	if err == nil {
		session = models.Session{UUID: cookie.Value}
		if ok, _ := session.CheckSession(r.Context(), env.DB); !ok {
			err = fmt.Errorf("invalid session")
		}
	}
	return session, err
}

var validPath = regexp.MustCompile("^/todos/(edit|update|delete)/([0-9]+)")

// parseURL は URLパスからTODOのIDを抽出し、指定されたハンドラー関数に渡すミドルウェアです。
func (env *Env) parseURL(fn func(*Env, http.ResponseWriter, *http.Request, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		q := validPath.FindStringSubmatch(r.URL.Path)
		for _, i := range q {
			log.Println(i)
		}
		if q == nil {
			http.NotFound(w, r)
			return
		}
		qi, err := strconv.Atoi(q[2])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		fn(env, w, r, qi)
	}

}

// makeHandler は Envを必要とするハンドラー関数を、標準の http.HandlerFunc インターフェースに適合させるクロージャです。
func makeHandler(env *Env, fn func(*Env, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(env, w, r)
	}
}

/*
StartMainServer は ルーティングを設定し、Webサーバーを起動します。
*/
func StartMainServer(env *Env) error {
	// FileServerでルートディレクトリを指定する。
	files := http.FileServer(http.Dir(env.Config.Static))
	/*
		指定されたパターンのハンドラーを DefaultServeMux に登録する。
		StripPrefix では URL のパスから prefix を削除している。
	*/
	http.Handle("/static/", http.StripPrefix("/static/", files))

	/*
		URLに対応するハンドラー関数を登録する
		func(ResponseWriter, *Request)のハンドラー関数を実装する必要がある
	*/
	http.HandleFunc("/", makeHandler(env, top))
	http.HandleFunc("/signup", makeHandler(env, signup))
	http.HandleFunc("/login", makeHandler(env, login))
	http.HandleFunc("/authenticate", makeHandler(env, authenticate))
	http.HandleFunc("/todos", makeHandler(env, index))
	http.HandleFunc("/logout", makeHandler(env, logout))
	http.HandleFunc("/todos/new", makeHandler(env, todoNew))
	http.HandleFunc("/todos/save", makeHandler(env, todoSave))
	http.HandleFunc("/todos/edit/", env.parseURL(todoEdit))
	http.HandleFunc("/todos/update/", env.parseURL(todoUpdate))
	http.HandleFunc("/todos/delete/", env.parseURL(todoDelete))
	return http.ListenAndServe(":"+env.Config.Port, nil)
}
