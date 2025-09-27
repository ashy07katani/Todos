package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Status int

const (
	Pending Status = iota
	InProgess
	Completed
)

type Todo struct {
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	TaskStatus  Status    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (s *Todo) FuncToImplement() {

}

type GetTodoResponse struct {
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

type CreateResponse struct {
	Message  string `json:"message"`
	UserName string `json:"username"`
}

type SignupResponse struct {
	Id        string    `json:"id"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"joined_at"`
}

type SignupRequest struct {
	UserName string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (s *SignupRequest) FuncToImplement() {

}

type User struct {
	Id             string    `json:"id"`
	UserName       string    `json:"username"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	CreatedAt      time.Time `json:"joined_at"`
}

type LoginUser struct {
	UserName string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (s *LoginUser) FuncToImplement() {

}

type SaveRefresh struct {
	UserId    string
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
}

type JWTClaim struct {
	UserName   string
	UserId     string
	Expires_At time.Time
	jwt.RegisteredClaims
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (s *ForgotPasswordRequest) FuncToImplement() {

}

type UpdatePasswordRequest struct {
	NewPassword string `json:"newpassword" validate:"required,min=8"`
}

func (s *UpdatePasswordRequest) FuncToImplement() {

}
