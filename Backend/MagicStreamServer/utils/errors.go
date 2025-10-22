package utils

import (
	"net/http"

	"github.com/afdhali/magic-stream/Backend/MagicStreamServer/repositories"
	"github.com/gin-gonic/gin"
)

// AppError represents a standardized application error
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code int, message string, details ...string) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// Common application errors
var (
	ErrBadRequest          = NewAppError(http.StatusBadRequest, "Bad request")
	ErrUnauthorized        = NewAppError(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden           = NewAppError(http.StatusForbidden, "Forbidden")
	ErrNotFound            = NewAppError(http.StatusNotFound, "Not found")
	ErrConflict            = NewAppError(http.StatusConflict, "Conflict")
	ErrInternalServerError = NewAppError(http.StatusInternalServerError, "Internal server error")
	ErrValidationFailed    = NewAppError(http.StatusBadRequest, "Validation failed")
)

// HandleError handles errors in a standardized way
func HandleError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		c.JSON(appErr.Code, gin.H{
			"error":   appErr.Message,
			"details": appErr.Details,
		})
		return
	}

	// Handle specific repository errors
	switch err {
	case repositories.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case repositories.ErrUserAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
	case repositories.ErrMovieNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
	case repositories.ErrMovieAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Movie already exists"})
	case repositories.ErrGenreNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Genre not found"})
	case repositories.ErrRefreshTokenNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Refresh token not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"details": err.Error(),
		})
	}
}

// ValidateRequest validates request data and returns standardized error
func ValidateRequest(c *gin.Context, data interface{}) bool {
	if err := c.ShouldBindJSON(data); err != nil {
		HandleError(c, NewAppError(http.StatusBadRequest, "Invalid request data", err.Error()))
		return false
	}
	return true
}