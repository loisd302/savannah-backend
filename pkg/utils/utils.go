package utils

import (
	"encoding/json"
	"net/http"

	"backend/pkg/models"
	"github.com/gin-gonic/gin"
)

// SuccessResponse sends a successful JSON response
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	response := models.Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, response)
}

// ErrorResponse sends an error JSON response
func ErrorResponse(c *gin.Context, statusCode int, message string, err interface{}) {
	response := models.Response{
		Success: false,
		Message: message,
		Error:   err,
	}
	c.JSON(statusCode, response)
}

// BadRequestResponse sends a bad request error response
func BadRequestResponse(c *gin.Context, message string, err interface{}) {
	ErrorResponse(c, http.StatusBadRequest, message, err)
}

// NotFoundResponse sends a not found error response
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message, nil)
}

// InternalServerErrorResponse sends an internal server error response
func InternalServerErrorResponse(c *gin.Context, message string, err interface{}) {
	ErrorResponse(c, http.StatusInternalServerError, message, err)
}

// UnauthorizedResponse sends an unauthorized error response
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// ParseJSON parses JSON from request body
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	return nil
}

// ToJSON converts an object to JSON string
func ToJSON(obj interface{}) (string, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// FromJSON converts JSON string to object
func FromJSON(jsonStr string, obj interface{}) error {
	return json.Unmarshal([]byte(jsonStr), obj)
}

// Contains checks if a string slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveIndex removes an element at a specific index from a slice
func RemoveIndex(slice []string, index int) []string {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}