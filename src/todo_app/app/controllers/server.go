package controllers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"todo_app/app/models"
	"todo_app/config"
)

type Env struct {
	DB     *sql.DB
	Config *config.ConfigList
}

/*
generateHTML は HTMLを生成します。
*/
func generateHTML(w http.ResponseWriter, data interface{}, fileNames ...string) {
	var files []string
	for _, file := range fileNames {
		files = append(files, fmt.Sprintf("app/views/templates/%s.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
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

func makeHandler(env *Env, fn func(*Env, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(env, w, r)
	}
}

/*
StartMainServer は サーバーを起動します。
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
	return http.ListenAndServe("localhost:"+env.Config.Port, nil)
}
