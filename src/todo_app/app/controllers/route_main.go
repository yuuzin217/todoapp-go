package controllers

import (
	"log"
	"net/http"
	"todo_app/app/models"
)

func top(env *Env, w http.ResponseWriter, r *http.Request) {
	_, err := env.checkSession(w, r)
	if err != nil {
		// top画面を生成
		generateHTML(w, nil, "layout", "public_navbar", "top")
	} else {
		// todos画面にリダイレクト
		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}

func index(env *Env, w http.ResponseWriter, r *http.Request) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/", MovedPermanently)
	} else {
		user, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos, err := user.GetTodosByUser(r.Context(), env.DB)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user.Todos = todos
		generateHTML(w, user, "layout", "private_navbar", "index")
	}
}

func todoNew(env *Env, w http.ResponseWriter, r *http.Request) {
	_, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		generateHTML(w, nil, "layout", "private_navbar", "todo_new")
	}
}

func todoSave(env *Env, w http.ResponseWriter, r *http.Request) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		err = r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		content := r.PostFormValue("content")
		if err := user.CreateTodo(r.Context(), env.DB, content); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}

func todoEdit(env *Env, w http.ResponseWriter, r *http.Request, id int) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		_, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todo, err := models.GetTodo(r.Context(), env.DB, id)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		generateHTML(w, todo, "layout", "private_navbar", "todo_edit")
	}

}

func todoUpdate(env *Env, w http.ResponseWriter, r *http.Request, id int) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		content := r.PostFormValue("content")
		todo := &models.Todo{ID: id, Content: content, UserID: user.ID}
		if err := todo.UpdateTodo(r.Context(), env.DB); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}

func todoDelete(env *Env, w http.ResponseWriter, r *http.Request, id int) {
	session, err := env.checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		_, err := session.GetUserBySession(r.Context(), env.DB)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todo, err := models.GetTodo(r.Context(), env.DB, id)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := todo.DeleteTodo(r.Context(), env.DB); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}
