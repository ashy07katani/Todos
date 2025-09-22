package router

import (
	"database/sql"
	"log"
	"net/http"
	"todos/config"
	"todos/handlers"
	"todos/mail"

	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB, authConfig *config.AuthConfig, mailConfig *mail.Mail, frontEndConfig *config.FrontEndConfig) *mux.Router {
	todoHandler := handlers.NewTodoHandler(db, authConfig, mailConfig, frontEndConfig)
	log.Println(todoHandler.MailConfig.From)
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
	r.HandleFunc("/users/forgot-password", todoHandler.ForgotPassword).Methods(http.MethodPost)
	r.HandleFunc("/users/update-password", todoHandler.UpdatePassword).Methods(http.MethodPatch)
	return r

}
