package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"todos/models"
	"todos/repository"
	"todos/utilities"
)

func (th *TodoHandler) CreateUser(rw http.ResponseWriter, r *http.Request) {
	// you will take fields like username, password, email from the request
	body := r.Body
	user := new(models.SignupRequest)
	err := json.NewDecoder(body).Decode(user)
	if err != nil {
		utilities.WriteError("unable to read user signup request", rw, http.StatusInternalServerError)
		return
	}
	// create user object that you could send to database, for that you would need to encrypt the password
	userDBobj := new(models.User)
	hashPassword, err := utilities.HashPassword(user.Password)
	if err != nil {
		utilities.WriteError("error occured while processing the user for Signup. ", rw, http.StatusInternalServerError)
		return
	}
	userDBobj.HashedPassword = hashPassword
	userDBobj.Email = user.Email
	userDBobj.UserName = user.UserName
	err = repository.CreateUser(r.Context(), th.DB, userDBobj)
	if err != nil {
		utilities.WriteError("error occured while creating the user.", rw, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
	response := models.CreateResponse{
		Message: "Todo created successfully",
		Id:      user.UserName,
	}
	utilities.WriteResponse(rw, response)
}

func (th *TodoHandler) Login(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	loginRequest := new(models.LoginUser)
	body := r.Body
	err := json.NewDecoder(body).Decode(loginRequest)
	if err != nil {
		utilities.WriteError("error occured while processing the request for Login. ", rw, http.StatusInternalServerError)
		return
	}

	user, err := repository.FetchUserWithUserID(r.Context(), th.DB, loginRequest.UserName)
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	isEqual, err := utilities.CompareHash(loginRequest.Password, user.HashedPassword)
	if err != nil {
		utilities.WriteError("error occured while comparing password for Login. ", rw, http.StatusInternalServerError)
		return
	}
	if !isEqual {
		utilities.WriteError("incorrect username or password", rw, http.StatusInternalServerError)
		return
	}

	tokenString, err := utilities.GenerateJWT(user, th.TokenConfig)
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	refreshTokenString, err := utilities.GenerateRefresh(user, th.TokenConfig)
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	ResponseMap := make(map[string]string)
	ResponseMap["token"] = tokenString

	cookie := http.Cookie{
		Name:  "refresh-token",
		Value: refreshTokenString,
	}
	http.SetCookie(rw, &cookie)
	sum := sha256.Sum256([]byte(refreshTokenString))
	refreshTokenHash := hex.EncodeToString(sum[:])
	err = repository.SaveRefreshToken(r.Context(), th.DB, &models.SaveRefresh{UserId: user.Id, TokenHash: refreshTokenHash, ExpiresAt: time.Now().Add(th.TokenConfig.RefreshTTL)})
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	utilities.WriteResponse(rw, ResponseMap)
}

func (th *TodoHandler) Refresh(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	refreshCookie, err := r.Cookie("refresh-token")
	if err != nil {
		utilities.WriteError(fmt.Sprintf("cannot read cookie. %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	log.Println(refreshCookie.Value)
	hashedArray := sha256.Sum256([]byte(refreshCookie.Value))
	hashedRefresh := hex.EncodeToString(hashedArray[:])
	savedRefresh := new(models.SaveRefresh)
	savedRefresh, err = repository.FetchRefreshToken(r.Context(), th.DB, hashedRefresh)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("Error with refresh token %s", err.Error), rw, http.StatusInternalServerError)
		return
	}

	//fetch the userID using that refresh token
	//create an access token and refresh token
	//invalidate the previous refresh token, use hash value to query the db and turn the value to false
	secret := th.TokenConfig.JWTSecret
	claim, err := utilities.GetClaimFromJWT(refreshCookie.Value, secret)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("Error fetching claim from token %s", err.Error), rw, http.StatusInternalServerError)
		return
	}
	if time.Now().After(savedRefresh.ExpiresAt) || savedRefresh.Revoked {
		utilities.WriteError("Refresh token has expired/invalid, need to relogin again", rw, http.StatusUnauthorized)
		return
	}

	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	user := &models.User{
		Id:       claim.UserId,
		UserName: claim.UserName,
	}
	tokenString, err := utilities.GenerateJWT(user, th.TokenConfig)
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	refreshTokenString, err := utilities.GenerateRefresh(user, th.TokenConfig)
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	// invalidate token
	err = repository.InvalidateRefreshToken(r.Context(), th.DB, hashedRefresh)
	if err != nil {
		utilities.WriteError("error invalidating refresh token", rw, http.StatusUnauthorized)
		return
	}

	ResponseMap := make(map[string]string)
	ResponseMap["token"] = tokenString

	cookie := http.Cookie{
		Name:  "refresh-token",
		Value: refreshTokenString,
	}
	http.SetCookie(rw, &cookie)
	sum := sha256.Sum256([]byte(refreshTokenString))
	refreshTokenHash := hex.EncodeToString(sum[:])
	err = repository.SaveRefreshToken(r.Context(), th.DB, &models.SaveRefresh{UserId: user.Id, TokenHash: refreshTokenHash, ExpiresAt: time.Now().Add(th.TokenConfig.RefreshTTL)})
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	utilities.WriteResponse(rw, ResponseMap)
}
