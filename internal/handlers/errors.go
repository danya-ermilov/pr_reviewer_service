package handlers

import (
	"encoding/json"
	"net/http"
)

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func sendAPIError(w http.ResponseWriter, httpStatus int, code, message string) {
	var e apiError
	e.Error.Code = code
	e.Error.Message = message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(e)
}
