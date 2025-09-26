package router

import (
	"database/sql"
	"log"
	"net/http"
	"todos/config"
	"todos/handlers"
	"todos/mail"
	"todos/middleware"

	"github.com/gorilla/mux"
)

func NewRouter(db *sql.DB, authConfig *config.AuthConfig, mailConfig *mail.Mail, frontEndConfig *config.FrontEndConfig) *mux.Router {
	todoHandler := handlers.NewTodoHandler(db, authConfig, mailConfig, frontEndConfig)
	log.Println(todoHandler.MailConfig.From)
	rl := new(middleware.RateLimiter)
	r := mux.NewRouter()
	authMiddleWare := middleware.AuthMiddleWare(todoHandler.TokenConfig.JWTSecret, todoHandler.DB)
	todoSubrouter := r.PathPrefix("/todos").Subrouter()
	userSubrouter := r.PathPrefix("/users").Subrouter()
	todoSubrouter.Use(rl.RateLimiterMiddleWare)
	todoSubrouter.Use(authMiddleWare)
	todoSubrouter.HandleFunc("/", todoHandler.ListAllTodos).Methods(http.MethodGet)
	todoSubrouter.HandleFunc("/search", todoHandler.SearchTask).Methods(http.MethodGet)
	todoSubrouter.HandleFunc("/{id}", todoHandler.FetchTodoByID).Methods(http.MethodGet)
	todoSubrouter.HandleFunc("/", todoHandler.CreateTask).Methods(http.MethodPost)
	todoSubrouter.HandleFunc("/{id}", todoHandler.DeleteTask).Methods(http.MethodDelete)
	todoSubrouter.HandleFunc("/{id}", todoHandler.UpdateTask).Methods(http.MethodPatch)

	userSubrouter.HandleFunc("/signup", todoHandler.CreateUser).Methods(http.MethodPost)
	userSubrouter.HandleFunc("/login", todoHandler.Login).Methods(http.MethodPost)
	userSubrouter.HandleFunc("/refresh", todoHandler.Refresh).Methods(http.MethodPost)
	userSubrouter.HandleFunc("/forgot-password", todoHandler.ForgotPassword).Methods(http.MethodPost)
	userSubrouter.HandleFunc("/update-password", todoHandler.UpdatePassword).Methods(http.MethodPatch)
	return r

}
