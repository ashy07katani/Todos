package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"todos/models"
	"todos/repository"
	"todos/utilities"
	validateapp "todos/validator"
)

func (th *TodoHandler) CreateUser(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		rw.WriteHeader(http.StatusOK)
		return
	}
	body := r.Body
	user := new(models.SignupRequest)
	err := json.NewDecoder(body).Decode(user)
	if err != nil {
		utilities.WriteError("unable to read user signup request", rw, http.StatusInternalServerError)
		return
	}
	err = validateapp.ValidateStruct(user)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while validating the input: %s", err.Error()), rw, http.StatusBadRequest)
		return
	}
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
		Message:  "Signup done successfully",
		UserName: user.UserName,
	}
	utilities.WriteResponse(rw, response)
}

func (th *TodoHandler) Login(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		rw.WriteHeader(http.StatusOK)
		return
	}
	loginRequest := new(models.LoginUser)
	body := r.Body
	err := json.NewDecoder(body).Decode(loginRequest)
	if err != nil {
		utilities.WriteError("error occured while processing the request for Login. ", rw, http.StatusInternalServerError)
		return
	}
	err = validateapp.ValidateStruct(loginRequest)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while validating the input: %s", err.Error()), rw, http.StatusBadRequest)
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
	if r.Method == http.MethodOptions {
		rw.WriteHeader(http.StatusOK)
		return
	}
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
		utilities.WriteError(fmt.Sprintf("Error with refresh token %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	secret := th.TokenConfig.JWTSecret
	claim, err := utilities.GetClaimFromJWT(refreshCookie.Value, secret)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("Error fetching claim from token %s", err.Error()), rw, http.StatusInternalServerError)
		return
	}
	if time.Now().After(savedRefresh.ExpiresAt) || savedRefresh.Revoked {
		utilities.WriteError("Refresh token has expired/invalid, need to relogin again", rw, http.StatusUnauthorized)
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

func (th *TodoHandler) ForgotPassword(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		rw.WriteHeader(http.StatusOK)
		return
	}
	body := r.Body
	forgotRequest := new(models.ForgotPasswordRequest)
	err := json.NewDecoder(body).Decode(forgotRequest)
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}
	err = validateapp.ValidateStruct(forgotRequest)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while validating the input: %s", err.Error()), rw, http.StatusBadRequest)
		return
	}

	userId, err := repository.IsEmailExists(context.Background(), th.DB, forgotRequest)
	if err != nil {
		utilities.WriteError(err.Error(), rw, http.StatusInternalServerError)
		return
	}

	forgotToken := rand.Text()
	shaToken := sha256.Sum256([]byte(forgotToken))
	shaTokenString := hex.EncodeToString(shaToken[:])
	forgotPassswordRequestMap := make(map[string]string)
	forgotPassswordRequestMap["email"] = forgotRequest.Email
	forgotPassswordRequestMap["token"] = shaTokenString
	forgotPassswordRequestMap["userid"] = userId
	errResponse := repository.StoreForgotPasswordToken(r.Context(), th.DB, forgotPassswordRequestMap)
	if errResponse != nil {
		utilities.WriteError(errResponse.Message, rw, errResponse.Status)
		return
	}
	go func() {
		forgotPasswordLink := fmt.Sprintf("%s%s?token=%s", th.FrontEndConfig.FrontEndDomain, th.FrontEndConfig.ResetPath, forgotToken)
		msg := utilities.GetMailBody("displayTodo@example.com", "Password Reset - Todo", fmt.Sprintf("Your password reset link is ready : %s ,and will be valid for 15 mins", forgotPasswordLink))
		a := th.MailConfig.GetAuth()

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*15)
		defer cancel()
		if err := th.MailConfig.SendMail(ctx, a, []string{"tripathi.ashish29@gmail.com", "cu.16bcs1336@gmail.com"}, msg); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Printf("Mail sending took too long")
				return
			}
			log.Printf("error sending mail : %s", err.Error())
		}
	}()
	responseMap := make(map[string]string)
	responseMap["message"] = "A mail has been sent with the reset password link to your registered Email. Kindly check."
	utilities.WriteResponse(rw, responseMap)

}

func (th *TodoHandler) UpdatePassword(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		rw.WriteHeader(http.StatusOK)
		return
	}
	values := r.URL.Query()
	fetchedToken := values.Get("token")
	tokenbyte := sha256.Sum256([]byte(fetchedToken))
	token := hex.EncodeToString(tokenbyte[:])
	if token == "" {
		utilities.WriteError("no token has been passed to the request", rw, http.StatusInternalServerError)
		return
	}

	body := r.Body
	newPassword := new(models.UpdatePasswordRequest)
	if err := json.NewDecoder(body).Decode(newPassword); err != nil {
		utilities.WriteError("error processing new password from request", rw, http.StatusInternalServerError)
		return
	}
	err := validateapp.ValidateStruct(newPassword)
	if err != nil {
		utilities.WriteError(fmt.Sprintf("error while validating the input: %s", err.Error()), rw, http.StatusBadRequest)
		return
	}
	hashPassword, err := utilities.HashPassword(newPassword.NewPassword)
	if err != nil {
		utilities.WriteError("error occured while processing the password for update request. ", rw, http.StatusInternalServerError)
		return
	}
	newPassword.NewPassword = hashPassword

	errResponse := repository.UpdatePassword(r.Context(), th.DB, newPassword, token)
	if errResponse != nil {
		utilities.WriteError(errResponse.Message, rw, errResponse.Status)
		return
	}
}
