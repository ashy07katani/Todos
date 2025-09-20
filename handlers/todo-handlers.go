package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"todos/config"
	"todos/mail"
	"todos/models"
	"todos/repository"
	"todos/utilities"

	"github.com/gorilla/mux"
)

type TodoHandler struct {
	DB          *sql.DB
	TokenConfig *config.AuthConfig
	MailConfig  *mail.Mail
}

func NewTodoHandler(DB *sql.DB, authConfig *config.AuthConfig, mailConfig *mail.Mail) *TodoHandler {
	todoHandler := new(TodoHandler)
	todoHandler.DB = DB
	todoHandler.TokenConfig = authConfig
	todoHandler.MailConfig = mailConfig
	return todoHandler
}

func (th *TodoHandler) ListAllTodos(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	queryMap := r.URL.Query()
	page := queryMap.Get("page")
	limit := queryMap.Get("limit")
	limitInt := 10
	pageInt := 0
	offset := 0
	if len(page) != 0 && len(limit) != 0 {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			utilities.WriteError(fmt.Sprintf("Invalid limit passed %s", err.Error()), rw, http.StatusInternalServerError)
			return
		}
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			utilities.WriteError(fmt.Sprintf("Invalid lpage passed %s", err.Error()), rw, http.StatusInternalServerError)
			return
		}
		offset = (pageInt - 1) * limitInt
	}

	ctx := r.Context()
	todos, err := repository.GetAllTodos(ctx, th.DB, offset, limitInt)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("Error fetching the todos %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	err = utilities.WriteResponse(rw, todos)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("Error writing the todos in response %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}

}

func (th *TodoHandler) FetchTodoByID(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	todo, _ := repository.GetTodoByID(r.Context(), th.DB, id)
	if todo != nil {
		json.NewEncoder(rw).Encode(todo)
	} else {
		errorMessage := fmt.Sprintf("There is no todo with Id: %s", id)
		utilities.WriteError(errorMessage, rw, http.StatusNotFound)
		return
	}

}

func (th *TodoHandler) CreateTask(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	body := r.Body
	v := new(models.Todo)
	err := json.NewDecoder(body).Decode(v)
	if err != nil {
		utilities.WriteError("error while creating task.", rw, http.StatusInternalServerError)
		return
	}
	err = repository.CreateTodo(r.Context(), th.DB, v)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while creating task, at Database layer: %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	response := models.CreateResponse{
		Message: "Todo created successfully",
		Id:      v.Id,
	}
	utilities.WriteResponse(rw, response)

}

func (th *TodoHandler) DeleteTask(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	rw.Header().Set("Content-Type", "application/json")
	err := repository.DeleteTodo(r.Context(), th.DB, id)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while deleting task, at Database layer: %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}

func (th *TodoHandler) UpdateTask(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	params := make(map[string]interface{})
	vars := mux.Vars(r)
	id := vars["id"]
	body := r.Body
	err := json.NewDecoder(body).Decode(&params)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while decoding request: %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	err = repository.UpdateTodo(r.Context(), th.DB, params, id)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while updating database: %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	utilities.WriteResponse(rw, params)
}

func (th *TodoHandler) SearchTask(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("entering SearchTask")
	rw.Header().Set("Content-Type", "application/json")
	queryMap := r.URL.Query()
	searchParam := queryMap.Get("query")
	page := queryMap.Get("page")
	limit := queryMap.Get("limit")
	limitInt := 10
	pageInt := 0
	offset := 0
	if len(page) != 0 && len(limit) != 0 {
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			utilities.WriteError(fmt.Sprintf("Invalid limit passed %s", err.Error()), rw, http.StatusInternalServerError)
			return
		}
		pageInt, err = strconv.Atoi(page)
		if err != nil {
			utilities.WriteError(fmt.Sprintf("Invalid lpage passed %s", err.Error()), rw, http.StatusInternalServerError)
			return
		}
		offset = (pageInt - 1) * limitInt
	}
	todo, err := repository.SearchTodo(r.Context(), th.DB, searchParam, limitInt, offset)
	if todo != nil {
		json.NewEncoder(rw).Encode(todo)
	} else {
		errorMessage := fmt.Sprintf("There is no todo with text: %s : %s", searchParam, err.Error())
		utilities.WriteError(errorMessage, rw, http.StatusNotFound)
		return
	}
}
