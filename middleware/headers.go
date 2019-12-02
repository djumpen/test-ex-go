package middleware

import (
	"fmt"
	"strings"

	"github.com/djumpen/test-ex-go/apperrors"
	"github.com/gin-gonic/gin"
)

var sourceTypes = []string{"game", "server", "payment"}

const SourceTypeHeader = "Source-Type"

// ValidateSourceType ensures that request contains valid Source-Type header
func ValidateSourceType(r Responder) gin.HandlerFunc {
	return func(c *gin.Context) {
		st := c.GetHeader(SourceTypeHeader)
		if len(st) == 0 {
			err := apperrors.NewBadRequest(fmt.Errorf("%s header required", SourceTypeHeader))
			processError(c, err, r)
			return
		}
		if !stringInSlice(strings.ToLower(st), sourceTypes) {
			err := apperrors.NewBadRequest(fmt.Errorf("Unsupported %s header", SourceTypeHeader))
			processError(c, err, r)
			return
		}
	}
}
