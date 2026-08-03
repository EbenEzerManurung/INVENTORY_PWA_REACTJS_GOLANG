package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==================== RESPONSE STRUCTS ====================

// Response represents standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// ==================== SUCCESS RESPONSES ====================

// Success sends a success response
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a created response
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Accepted sends an accepted response
func Accepted(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusAccepted, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent sends a no content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ==================== ERROR RESPONSES ====================

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// BadRequest sends a bad request response
func BadRequest(c *gin.Context, message string, err string) {
	Error(c, http.StatusBadRequest, message, err)
}

// Unauthorized sends an unauthorized response
func Unauthorized(c *gin.Context, message string, err string) {
	Error(c, http.StatusUnauthorized, message, err)
}

// Forbidden sends a forbidden response
func Forbidden(c *gin.Context, message string, err string) {
	Error(c, http.StatusForbidden, message, err)
}

// NotFound sends a not found response
func NotFound(c *gin.Context, message string, err string) {
	Error(c, http.StatusNotFound, message, err)
}

// InternalServerError sends an internal server error response
func InternalServerError(c *gin.Context, message string, err string) {
	Error(c, http.StatusInternalServerError, message, err)
}

// Conflict sends a conflict response
func Conflict(c *gin.Context, message string, err string) {
	Error(c, http.StatusConflict, message, err)
}

// UnprocessableEntity sends an unprocessable entity response
func UnprocessableEntity(c *gin.Context, message string, err string) {
	Error(c, http.StatusUnprocessableEntity, message, err)
}

// ==================== PAGINATED RESPONSE ====================

// PaginatedResponse sends a paginated response
func PaginatedResponse(c *gin.Context, data interface{}, total int, page int, limit int) {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"pagination": PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// ==================== VALIDATION RESPONSE ====================

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}

// ==================== FILE RESPONSE ====================

// FileResponse sends a file response
func FileResponse(c *gin.Context, filePath string, fileName string) {
	c.FileAttachment(filePath, fileName)
}

// ==================== STREAM RESPONSE ====================

// StreamResponse sends a stream response
func StreamResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}