package api

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type commonResource struct {
	resp *Responder
}

func NewCommonResource(resp *Responder) *commonResource {
	return &commonResource{
		resp: resp,
	}
}

func (r *commonResource) Health(c *gin.Context) {
	r.resp.OK(c, nil)
}

func (r *commonResource) NotFound(c *gin.Context) {
	r.resp.NotFound(c, errors.New("Resource not found"))
}
