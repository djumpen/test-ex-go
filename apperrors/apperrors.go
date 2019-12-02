package apperrors

type SimpleError struct {
	err error
}

func (e *SimpleError) Error() string {
	return e.err.Error()
}

type NotFound struct{ SimpleError }

type BadRequest struct{ SimpleError }

func NewNotFound(err error) *NotFound {
	return &NotFound{SimpleError{err}}
}

func NewBadRequest(err error) *BadRequest {
	return &BadRequest{SimpleError{err}}
}
