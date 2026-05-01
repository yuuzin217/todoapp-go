package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"todo_app/app/models"
	"todo_app/config"
)

// Env はコントローラー層全体で共有される依存関係 (データベース接続や設定など) を保持する構造体です。
// これにより、グローバル変数を使わずに各ハンドラーへ注入(Dependency Injection)できます。
type Env struct {
	DB            *sql.DB                       // データベースコネクション
	Config        *config.ConfigList            // アプリケーション設定
	TemplateCache map[string]*template.Template // テンプレートキャッシュ (本番用)
	Translations  map[string]map[string]string  // 言語ごとの翻訳データ
	Mu            sync.RWMutex                  // キャッシュ/翻訳操作用ミューテックス
}

/*
LoadTranslations は i18n ディレクトリから翻訳ファイルを読み込みます。
*/
func (env *Env) LoadTranslations() error {
	env.Mu.Lock()
	defer env.Mu.Unlock()

	env.Translations = make(map[string]map[string]string)
	i18nDir := "app/views/i18n"
	files, err := os.ReadDir(i18nDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			lang := strings.TrimSuffix(file.Name(), ".json")
			data, err := os.ReadFile(filepath.Join(i18nDir, file.Name()))
			if err != nil {
				return err
			}

			var trans map[string]string
			if err := json.Unmarshal(data, &trans); err != nil {
				return err
			}
			env.Translations[lang] = trans
		}
	}
	return nil
}

/*
getLang は リクエストから言語設定を取得します (Cookie > Accept-Language > Default)。
*/
func (env *Env) getLang(r *http.Request) string {
	// 1. Cookieを確認
	if cookie, err := r.Cookie("lang"); err == nil {
		if _, ok := env.Translations[cookie.Value]; ok {
			return cookie.Value
		}
	}

	// 2. Accept-Language ヘッダーを確認
	accept := r.Header.Get("Accept-Language")
	if accept != "" {
		langs := strings.Split(accept, ",")
		for _, l := range langs {
			tag := strings.Split(strings.TrimSpace(l), ";")[0]
			if strings.HasPrefix(tag, "ja") {
				return "ja"
			}
			if strings.HasPrefix(tag, "en") {
				return "en"
			}
		}
	}

	return "en" // デフォルト
}

/*
generateHTML は 指定されたテンプレートファイル群をパースしてHTMLを生成し、レスポンスとして書き込みます。
*/
func (env *Env) generateHTML(w http.ResponseWriter, r *http.Request, data interface{}, fileNames ...string) {
	var files []string
	for _, file := range fileNames {
		files = append(files, fmt.Sprintf("app/views/templates/%s.html", file))
	}

	lang := env.getLang(r)
	key := strings.Join(fileNames, ",") + ":" + lang

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

	// 翻訳関数の定義
	funcMap := template.FuncMap{
		"TR": func(key string) string {
			env.Mu.RLock()
			defer env.Mu.RUnlock()
			if trans, ok := env.Translations[lang]; ok {
				if val, ok := trans[key]; ok {
					return val
				}
			}
			return key
		},
		"CurrentLang": func() string {
			return lang
		},
	}

	// テンプレートのパース
	templates := template.Must(template.New("").Funcs(funcMap).ParseFiles(files...))

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
setLang は 言語設定をCookieに保存するハンドラーです。
*/
func setLang(env *Env, w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("l")
	if lang != "ja" && lang != "en" {
		lang = "en"
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "lang",
		Value: lang,
		Path:  "/",
	})

	// 元のページに戻るか、トップへ
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, http.StatusFound)
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
	http.HandleFunc("/set-lang", makeHandler(env, setLang))
	return http.ListenAndServe(":"+env.Config.Port, nil)
}
