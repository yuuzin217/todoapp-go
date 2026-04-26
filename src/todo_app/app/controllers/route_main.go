package controllers

import (
	"log"
	"net/http"
	"todo_app/app/models"
)

func top(w http.ResponseWriter, r *http.Request) {
	_, err := checkSession(w, r)
	if err != nil {
		// top画面を生成
		generateHTML(w, nil, "layout", "public_navbar", "top")
	} else {
		// todos画面にリダイレクト
		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	session, err := checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/", MovedPermanently)
	} else {
		user, err := session.GetUserBySession()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todos, err := user.GetTodosByUser()
		if err != nil {
			log.Println(err)
		}
		user.Todos = todos
		generateHTML(w, user, "layout", "private_navbar", "index")
	}
}

func todoNew(w http.ResponseWriter, r *http.Request) {
	_, err := checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		generateHTML(w, nil, "layout", "private_navbar", "todo_new")
	}
}

func todoSave(w http.ResponseWriter, r *http.Request) {
	session, err := checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		err = r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := session.GetUserBySession()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		content := r.PostFormValue("content")
		if err := user.CreateTodo(content); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}

func todoEdit(w http.ResponseWriter, r *http.Request, id int) {
	session, err := checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		_, err := session.GetUserBySession()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todo, err := models.GetTodo(id)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		generateHTML(w, todo, "layout", "private_navbar", "todo_edit")
	}

}

func todoUpdate(w http.ResponseWriter, r *http.Request, id int) {
	session, err := checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := session.GetUserBySession()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		content := r.PostFormValue("content")
		todo := &models.Todo{ID: id, Content: content, UserID: user.ID}
		if err := todo.UpdateTodo(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}

func todoDelete(w http.ResponseWriter, r *http.Request, id int) {
	session, err := checkSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", MovedPermanently)
	} else {
		_, err := session.GetUserBySession()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		todo, err := models.GetTodo(id)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := todo.DeleteTodo(); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/todos", MovedPermanently)
	}
}
