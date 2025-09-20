package utilities

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"todos/config"
	"todos/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func CreateErrorResponse(errorMessage string, statusCode int) *models.ErrorResponse {
	errorResponse := new(models.ErrorResponse)
	errorResponse.Message = errorMessage
	errorResponse.Status = statusCode
	return errorResponse
}

func WriteResponse(rw http.ResponseWriter, v any) error {
	return json.NewEncoder(rw).Encode(v)
}

func WriteError(errorMessage string, rw http.ResponseWriter, httpErrorCode int) {
	errorResponse := CreateErrorResponse((errorMessage), httpErrorCode)
	rw.WriteHeader(httpErrorCode)
	WriteResponse(rw, errorResponse)
}

func HashPassword(rawPassword string) (hashPassword string, err error) {
	bytePassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), 10)
	computedHash := string(bytePassword)
	return computedHash, err
}

func CompareHash(rawPassword string, hashPassowrd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassowrd), []byte(rawPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GenerateJWT(user *models.User, tokenConfig *config.AuthConfig) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"username":   user.UserName,
		"userId":     user.Id,
		"expires_At": time.Now().Add(tokenConfig.AccessTTL),
		"issued_At":  time.Now(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(tokenConfig.JWTSecret))
	return tokenString, err
}

func GenerateRefresh(user *models.User, tokenConfig *config.AuthConfig) (tokenString string, err error) {
	claims := jwt.MapClaims{
		"username":   user.UserName,
		"userId":     user.Id,
		"expires_At": time.Now().Add(tokenConfig.RefreshTTL),
		"issued_At":  time.Now(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(tokenConfig.JWTSecret))
	return tokenString, err
}

func GetClaimFromJWT(tokenString string, topSecret string) (*models.JWTClaim, error) {
	claim := new(models.JWTClaim)
	_, err := jwt.ParseWithClaims(tokenString, claim, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return []byte(topSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claim, err
}

func GetMailBody(to string, subject string, mailBody string) []byte {

	msg := []byte(fmt.Sprintf("From: yourname@gmail.com\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+ // blank line headers aur body ke beech
		"%s\r\n", to, subject, mailBody))
	return msg
}
