package models

import "time"

type Status int

const (
	Pending Status = iota
	InProgess
	Completed
)

type Todo struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TaskStatus  Status    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type CreateTodoResponse struct {
	Message string `json:"message"`
	Id      string `json:"id"`
}
