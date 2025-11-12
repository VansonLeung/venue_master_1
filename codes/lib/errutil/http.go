package errutil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the canonical shape for API errors.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Write writes a JSON error payload with the correct status code.
func Write(ctx *gin.Context, status int, code string, message string, details any) {
	ctx.AbortWithStatusJSON(status, ErrorResponse{Code: code, Message: message, Details: details})
}

// HandleInternal writes a standard 500 error.
func HandleInternal(ctx *gin.Context, err error) {
	Write(ctx, http.StatusInternalServerError, "internal_error", "Something went wrong", err.Error())
}
