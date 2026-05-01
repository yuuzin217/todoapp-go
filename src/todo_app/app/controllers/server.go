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

// Env はコントローラー層全体で共有される依存関係（データベース接続、設定、キャッシュ等）を保持する構造体です。
// グローバル変数に頼らず、この構造体を介して各ハンドラーへ必要なリソースを注入（Dependency Injection）します。
type Env struct {
	DB            *sql.DB                       // データベース接続プール
	Config        *config.ConfigList            // アプリケーション全体の共通設定
	TemplateCache map[string]*template.Template // パフォーマンス向上のためのテンプレートキャッシュ（本番環境用）
	Translations  map[string]map[string]string  // 多言語対応（i18n）のための翻訳マッピングデータ
	Mu            sync.RWMutex                  // キャッシュおよび翻訳データの並行アクセスを安全に制御するためのミューテックス
}

// LoadTranslations は app/views/i18n ディレクトリ配下の JSON ファイルから翻訳データをメモリに読み込みます。
// アプリケーションの起動時に呼び出されることを想定しています。
func (env *Env) LoadTranslations() error {
	env.Mu.Lock()
	defer env.Mu.Unlock()

	env.Translations = make(map[string]map[string]string)
	i18nDir := "app/views/i18n"
	files, err := os.ReadDir(i18nDir)
	if err != nil {
		log.Printf("Failed to read i18n directory: %v", err)
		return err
	}

	for _, file := range files {
		// .json 拡張子のファイルのみを処理対象とします。
		if filepath.Ext(file.Name()) == ".json" {
			// ファイル名から拡張子を除いたものを言語コード（例: "ja", "en"）として扱います。
			lang := strings.TrimSuffix(file.Name(), ".json")
			data, err := os.ReadFile(filepath.Join(i18nDir, file.Name()))
			if err != nil {
				log.Printf("Failed to read translation file %s: %v", file.Name(), err)
				return err
			}

			var trans map[string]string
			if err := json.Unmarshal(data, &trans); err != nil {
				log.Printf("Failed to parse JSON in %s: %v", file.Name(), err)
				return err
			}
			env.Translations[lang] = trans
		}
	}
	return nil
}

// getLang は現在のリクエストに対して適用すべき言語を決定します。
// 優先順位: 1. Cookie ("lang"), 2. Accept-Language ヘッダー, 3. デフォルト ("en")
func (env *Env) getLang(r *http.Request) string {
	// 1. ユーザーが明示的に設定した Cookie を優先します。
	if cookie, err := r.Cookie("lang"); err == nil {
		if _, ok := env.Translations[cookie.Value]; ok {
			return cookie.Value
		}
	}

	// 2. ブラウザの設定（Accept-Language）から最適な言語を推測します。
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

	// TODO: サポート言語が増えた場合に備え、Translations マップのキーから動的に判定する仕組みを検討。
	return "en"
}

// generateHTML は指定された複数のテンプレートファイルを結合・パースし、HTMLレスポンスを生成します。
// 本番環境（production）では、結合済みのテンプレートオブジェクトをキャッシュして再利用します。
func (env *Env) generateHTML(w http.ResponseWriter, r *http.Request, data interface{}, fileNames ...string) {
	var files []string
	for _, file := range fileNames {
		files = append(files, fmt.Sprintf("app/views/templates/%s.html", file))
	}

	lang := env.getLang(r)
	// テンプレートの組み合わせと言語をキーとしてキャッシュを管理します。
	key := strings.Join(fileNames, ",") + ":" + lang

	// 本番環境かつキャッシュが存在する場合は、即座に実行して戻ります。
	if env.Config.Env == "production" {
		env.Mu.RLock()
		t, ok := env.TemplateCache[key]
		env.Mu.RUnlock()
		if ok {
			if err := t.ExecuteTemplate(w, "layout", data); err != nil {
				log.Printf("Failed to execute cached template: %v", err)
			}
			return
		}
	}

	// テンプレート内で利用可能なカスタム関数を定義します。
	funcMap := template.FuncMap{
		// TR はキーに基づいた翻訳文字列を返します。
		"TR": func(key string) string {
			env.Mu.RLock()
			defer env.Mu.RUnlock()
			if trans, ok := env.Translations[lang]; ok {
				if val, ok := trans[key]; ok {
					return val
				}
			}
			return key // 翻訳が見つからない場合はキーをそのまま返します。
		},
		// CurrentLang は現在適用されている言語コードを返します。
		"CurrentLang": func() string {
			return lang
		},
	}

	// テンプレートの新規作成と関数の登録、およびファイルのパース。
	// template.Must はパースエラー時にパニックを発生させるため、開発中にミスに気づきやすくなります。
	templates := template.Must(template.New("").Funcs(funcMap).ParseFiles(files...))

	// 本番環境の場合は、次回以降のためにパース結果をキャッシュします。
	if env.Config.Env == "production" {
		env.Mu.Lock()
		if env.TemplateCache == nil {
			env.TemplateCache = make(map[string]*template.Template)
		}
		env.TemplateCache[key] = templates
		env.Mu.Unlock()
	}

	if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// setLang はユーザーの言語設定を Cookie に保存し、元のページへリダイレクトします。
func setLang(env *Env, w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("l")
	// 不正な言語コードが指定された場合はデフォルトをセットします。
	if lang != "ja" && lang != "en" {
		lang = "en"
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "lang",
		Value: lang,
		Path:  "/", // サイト全体で有効にするためにルートを指定。
	})

	// リファラー（遷移元URL）があればそこへ、なければトップに戻ります。
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, MovedTemporarily)
}

// checkSession はリクエストに含まれるクッキーからセッションの有効性を確認します。
// ログイン必須のページにおけるガードとして機能します。
func (env *Env) checkSession(w http.ResponseWriter, r *http.Request) (session models.Session, err error) {
	cookie, err := r.Cookie("_cookie")
	if err == nil {
		session = models.Session{UUID: cookie.Value}
		// DBを参照してセッションの存在と期限を確認します。
		if ok, _ := session.CheckSession(r.Context(), env.DB); !ok {
			err = fmt.Errorf("invalid session")
		}
	}
	return session, err
}

// IDを含む特定のパス（編集・更新・削除）を検証するための正規表現。
var validPath = regexp.MustCompile("^/todos/(edit|update|delete)/([0-9]+)")

// parseURL は URL パスからTODOのIDを抽出し、指定されたハンドラー関数に ID を渡すためのミドルウェアです。
// これにより、個別のハンドラー内でパスのパース処理を重複して書く必要がなくなります。
func (env *Env) parseURL(fn func(*Env, http.ResponseWriter, *http.Request, int)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := validPath.FindStringSubmatch(r.URL.Path)
		if q == nil {
			http.NotFound(w, r)
			return
		}
		// 正規表現のキャプチャグループから数値を取得して int に変換します。
		qi, err := strconv.Atoi(q[2])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		fn(env, w, r, qi)
	}
}

// makeHandler は Env 構造体を受け取る独自形式のハンドラーを、標準の http.HandlerFunc に適合させるためのラッパーです。
func makeHandler(env *Env, fn func(*Env, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(env, w, r)
	}
}

// StartMainServer はすべてのルーティングを登録し、指定されたポートで HTTP サーバーを起動します。
func StartMainServer(env *Env) error {
	// 静的ファイル（CSS, JS, 画像など）の配信設定。
	files := http.FileServer(http.Dir(env.Config.Static))
	http.Handle("/static/", http.StripPrefix("/static/", files))

	// ルーティングの登録。各ハンドラーは makeHandler 等でラップして Env を注入します。
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

	log.Printf("Starting server on port %s", env.Config.Port)
	return http.ListenAndServe(":"+env.Config.Port, nil)
}
