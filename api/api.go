package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

type SimpleResponder interface {
	OK(c *gin.Context, res interface{})
	Created(c *gin.Context, res interface{})
	NotFound(c *gin.Context, err error)
	BadRequest(c *gin.Context, description string, err error)
}

type Response struct {
	Success bool        `json:"success"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Code    int         `json:"code,omitempty"`
}

type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
}

type Responder struct{}

// NewResponder returns Responder struct
func NewResponder() *Responder {
	return &Responder{}
}

func (r *Responder) OK(c *gin.Context, res interface{}) {
	response(c, http.StatusOK, 0, res, nil)
}

func (r *Responder) Created(c *gin.Context, res interface{}) {
	response(c, http.StatusCreated, 0, res, nil)
}

func (r *Responder) BadRequest(c *gin.Context, description string, err error) {
	responseErr(c, http.StatusBadRequest, description, err, nil)
}

func (r *Responder) NotFound(c *gin.Context, err error) {
	responseErr(c, http.StatusNotFound, "", err, nil)
}

func (r *Responder) NotAllowed(c *gin.Context, err error) {
	responseErr(c, http.StatusMethodNotAllowed, "", err, nil)
}

func (r *Responder) Unprocessable(c *gin.Context, description string, err error) {
	responseErr(c, http.StatusUnprocessableEntity, description, err, nil)
}

func (r *Responder) InternalError(c *gin.Context, err error) {
	responseErr(c, http.StatusInternalServerError, "", err, nil)
}

func (r *Responder) ResponseErrWithFields(c *gin.Context, fields []string) {
	responseErr(c, http.StatusUnprocessableEntity, "", nil, fields)
}

func response(c *gin.Context, httpCode, code int, res interface{}, pagination *Pagination) {
	respType := "item"
	data := gin.H{
		"item": res,
	}
	if pagination != nil {
		respType = "list"
		data = gin.H{
			"items":      res,
			"pagination": pagination,
		}
	}
	c.JSON(httpCode, Response{
		Success: true,
		Type:    respType,
		Code:    code,
		Data:    data,
	})
	c.Abort()
}

func responseErr(c *gin.Context, httpCode int, description string, err error, fields []string) {
	var errorText string
	if err != nil {
		errorText = err.Error()
	}
	if httpCode == http.StatusInternalServerError {
		errorText = "Server error"
	}
	if description != "" {
		errorText = description
	}
	var errData []string
	if err == nil {
		errData = fields
	} else {
		errData = []string{errorText}
	}
	c.JSON(httpCode, Response{
		Success: false,
		Type:    "request_error",
		Data: gin.H{
			"errors": errData,
		},
	})
	c.Abort()
}

// GetCorsConfig returns CORS configuration
func GetCorsConfig() cors.Config {
	return cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Length", "Content-Type", "Source-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}
