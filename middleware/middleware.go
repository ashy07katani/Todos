package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"
	"todos/repository"
	"todos/utilities"
)

func AuthMiddleWare(secret string, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			arr := strings.Split(header, " ")
			if len(arr) != 2 || arr[0] != "Bearer" {
				http.Error(w, "invalid Authorization header in the request", http.StatusUnauthorized)
				return
			}
			claim, err := utilities.GetClaimFromJWT(arr[1], secret)
			if err != nil {
				http.Error(w, "error fetching claim from token", http.StatusInternalServerError)
				return
			}
			if time.Now().After(claim.Expires_At) {
				http.Error(w, "token is already expired", http.StatusUnauthorized)
				return
			}
			userName := claim.UserName
			user, err := repository.FetchUserWithUserID(r.Context(), db, userName)
			if err != nil {
				http.Error(w, "token is not linked to any real user", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "userId", user.Id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
