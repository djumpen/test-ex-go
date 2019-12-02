package middleware

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/djumpen/test-ex-go/apperrors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	validator "gopkg.in/go-playground/validator.v9"
)

type Responder interface {
	BadRequest(c *gin.Context, description string, err error)
	NotFound(c *gin.Context, err error)
	ResponseErrWithFields(c *gin.Context, fields []string)
	InternalError(c *gin.Context, err error)
}

func ErrorHandler(r Responder) gin.HandlerFunc {
	return func(c *gin.Context) {
		// handle errors from previous middleware
		if len(c.Errors) > 0 {
			processError(c, c.Errors[0].Err, r)
			return
		}
		c.Next()
		// handle errors from main handler
		if len(c.Errors) > 0 {
			processError(c, c.Errors[0].Err, r)
			return
		}
	}
}

func processError(c *gin.Context, err error, r Responder) {
	if err == nil {
		return
	}

	log.Printf("ERROR: %v", err)

	switch ve := errors.Cause(err).(type) {
	case validator.ValidationErrors:
		fields := make([]string, 0, len(ve))
		for _, v := range ve {
			fields = append(fields, validationErrorToText(v))
		}
		r.ResponseErrWithFields(c, fields)
	case *apperrors.Validation:
		r.ResponseErrWithFields(c, []string{ve.Error()})
	case *json.UnmarshalTypeError:
		err := err.(*json.UnmarshalTypeError)
		validationError := unmarshalTypeErrorToValidation(err)
		r.ResponseErrWithFields(c, []string{validationError})
	case *apperrors.BadRequest:
		r.BadRequest(c, ve.Error(), ve)
	default:
		r.InternalError(c, err)
	}
}

func validationErrorToText(e validator.FieldError) string {
	word := split(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", word)
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s", word, e.Param())
	case "min":
		return fmt.Sprintf("%s must be longer than %s", word, e.Param())
	case "numeric":
		return fmt.Sprintf("%s must be numeric", word)
	case "email":
		return fmt.Sprintf("Invalid email format")
	case "len":
		return fmt.Sprintf("%s must be %s characters long", word, e.Param())
	case "url":
		return fmt.Sprintf("Invalid url format")
	}
	return fmt.Sprintf("%s is not valid", word)
}

func unmarshalTypeErrorToValidation(err *json.UnmarshalTypeError) string {
	kind := err.Type.Kind().String()
	if stringInSlice(kind, []string{"int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64"}) {
		kind = "int"
	}
	return fmt.Sprintf("%s has type '%s', but '%s' required", err.Field, err.Value, kind)
}
