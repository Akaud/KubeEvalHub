package handler

import (
	"errors"
	"net/http"

	"backend/internal/service"
)

func writeAnalysisHandlerError(w http.ResponseWriter, err error, internalMessage string) {
	var badReq *badRequestError
	switch {
	case errors.As(err, &badReq):
		status := http.StatusBadRequest
		if badReq.message == "unauthorized" {
			status = http.StatusUnauthorized
		}
		writeError(w, status, badReq.message)
	case errors.Is(err, service.ErrInvalidAnalysisQuery):
		writeError(w, http.StatusBadRequest, "invalid analysis query")
	default:
		writeError(w, http.StatusInternalServerError, internalMessage)
	}
}
