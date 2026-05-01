package controllers

import (
	"log"
	"net/http"
	"todo_app/app/models"
)

// top はアプリケーションのランディングページ（ルート）を制御します。
// すでに認証済みのユーザーがアクセスした場合は、利便性のために自動的にタスク一覧（/todos）へ誘導します。
func top(env *Env, w http.ResponseWriter, r *http.Request) {
	_, err := env.checkSession(w, r)
	if err != nil {
		// 未ログイン時は紹介ページ（top.html）を表示します。
		env.generateHTML(w, r, nil, "layout", "public_navbar", "top")
	} else {
		// ログイン済みならタスク一覧へ。
		http.Redirect(w, r, "/todos", MovedTemporarily)
	}
}

// index はログインユーザー専用のメイン画面（タスク一覧）を表示します。
// セッションからユーザーを特定し、そのユーザーが所有するタスクのみをDBから抽出して渡します。
func index(env *Env, w http.ResponseWriter, r *http.Request) {
	session, err := env.checkSession(w, r)
	if err != nil {
		// セッションが無効（期限切れ等）な場合はトップページへ戻します。
		http.Redirect(w, r, "/", MovedTemporarily)
	} else {
		user, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Printf("Failed to get user from session: %v", err)
			http.Error(w, "User identification failed", http.StatusInternalServerError)
			return
		}

		todos, err := user.GetTodosByUser(r.Context(), env.DB)
		if err != nil {
			log.Printf("Failed to load todos for user %d: %v", user.ID, err)
			http.Error(w, "Task loading failed", http.StatusInternalServerError)
			return
		}

		user.Todos = todos
		env.generateHTML(w, r, user, "layout", "private_navbar", "index")
	}
}

// todoNew は新規タスク作成用の入力フォームを表示します。
func todoNew(env *Env, w http.ResponseWriter, r *http.Request) {
	_, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedTemporarily)
	} else {
		env.generateHTML(w, r, nil, "layout", "private_navbar", "todo_new")
	}
}

// todoSave は送信された新規タスクの内容をバリデーションし、データベースに永続化します。
func todoSave(env *Env, w http.ResponseWriter, r *http.Request) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedTemporarily)
	} else {
		if err := r.ParseForm(); err != nil {
			log.Printf("Failed to parse todo form: %v", err)
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		user, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Printf("Session-User sync error: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		content := r.PostFormValue("content")
		if err := user.CreateTodo(r.Context(), env.DB, content); err != nil {
			log.Printf("Failed to save todo for user %d: %v", user.ID, err)
			http.Error(w, "Storage failed", http.StatusInternalServerError)
			return
		}

		// 保存後は一覧に戻り、最新の状態を表示させます。
		http.Redirect(w, r, "/todos", MovedTemporarily)
	}
}

// todoEdit は既存タスクの編集画面を表示します。
// 対象のタスクが現在のログインユーザーのものであることを保証する必要があります。
func todoEdit(env *Env, w http.ResponseWriter, r *http.Request, id int) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedTemporarily)
	} else {
		// セッションのユーザーがタスクの所有者であることを確認するための前段階。
		_, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Printf("User sync error during edit: %v", err)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		todo, err := models.GetTodo(r.Context(), env.DB, id)
		if err != nil {
			log.Printf("Task %d not found: %v", id, err)
			http.NotFound(w, r)
			return
		}
		
		// TODO: ここで todo.UserID と user.ID の一致チェックを入れるとより安全です。
		env.generateHTML(w, r, todo, "layout", "private_navbar", "todo_edit")
	}
}

// todoUpdate は編集されたタスク内容でデータベースを更新します。
func todoUpdate(env *Env, w http.ResponseWriter, r *http.Request, id int) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedTemporarily)
	} else {
		if err := r.ParseForm(); err != nil {
			log.Printf("Form parse error in update: %v", err)
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		user, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Printf("User resolution failed in update: %v", err)
			http.Error(w, "Identification failed", http.StatusInternalServerError)
			return
		}

		content := r.PostFormValue("content")
		todo := &models.Todo{ID: id, Content: content, UserID: user.ID}
		// UpdateTodo は内部的に UserID をチェックし、他人のタスクを編集できないようにします。
		if err := todo.UpdateTodo(r.Context(), env.DB); err != nil {
			log.Printf("Update failed for task %d (user %d): %v", id, user.ID, err)
			http.Error(w, "Update failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/todos", MovedTemporarily)
	}
}

// todoDelete は指定された ID のタスクを削除します。
func todoDelete(env *Env, w http.ResponseWriter, r *http.Request, id int) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedTemporarily)
	} else {
		_, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Printf("User sync error during delete: %v", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// 削除対象の存在確認。
		todo, err := models.GetTodo(r.Context(), env.DB, id)
		if err != nil {
			log.Printf("Delete target %d not found: %v", id, err)
			http.Redirect(w, r, "/todos", MovedTemporarily)
			return
		}

		// DeleteTodo は所有者権限を内部でチェックすることを期待します。
		if err := todo.DeleteTodo(r.Context(), env.DB); err != nil {
			log.Printf("Failed to delete task %d: %v", id, err)
			http.Error(w, "Deletion failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/todos", MovedTemporarily)
	}
}
