package response

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string, err interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}

func BadRequest(c *gin.Context, message string, err interface{}) {
	Error(c, 400, message, err)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message, nil)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, 403, message, nil)
}

func NotFound(c *gin.Context, message string) {
	Error(c, 404, message, nil)
}

func InternalServerError(c *gin.Context, message string, err interface{}) {
	Error(c, 500, message, err)
}