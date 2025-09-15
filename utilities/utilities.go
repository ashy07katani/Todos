package utilities

import (
	"encoding/json"
	"net/http"
	"todos/models"
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
