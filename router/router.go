package router

import (
	"database/sql"
	"net/http"
	"todos/config"
	"todos/handlers"

	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB, authConfig *config.AuthConfig) *mux.Router {
	todoHandler := handlers.NewTodoHandler(db, authConfig)
	r := mux.NewRouter()
	r.HandleFunc("/todos", todoHandler.ListAllTodos).Methods(http.MethodGet)
	r.HandleFunc("/todos/search", todoHandler.SearchTask).Methods(http.MethodGet)
	r.HandleFunc("/todos/{id}", todoHandler.FetchTodoByID).Methods(http.MethodGet)
	r.HandleFunc("/todos", todoHandler.CreateTask).Methods(http.MethodPost)
	r.HandleFunc("/todos/{id}", todoHandler.DeleteTask).Methods(http.MethodDelete)
	r.HandleFunc("/todos/{id}", todoHandler.UpdateTask).Methods(http.MethodPatch)

	r.HandleFunc("/users/signup", todoHandler.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/users/login", todoHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/users/refresh", todoHandler.Refresh).Methods(http.MethodPost)
	return r
	//fmt.Println("starting server at port 6161")
}
