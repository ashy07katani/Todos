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
	Message string `json:"message"`
	Id      string `json:"id"`
}

type SignupResponse struct {
	Id        string    `json:"id"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"joined_at"`
}

type SignupRequest struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	Id             string    `json:"id"`
	UserName       string    `json:"username"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	CreatedAt      time.Time `json:"joined_at"`
}

type LoginUser struct {
	UserName string `json:"username"`
	Password string `json:"password"`
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
