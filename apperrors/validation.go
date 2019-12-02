package apperrors

type Validation struct {
	namespace string
	err       error
}

func NewValidation(ns string, err error) *Validation {
	return &Validation{
		namespace: ns,
		err:       err,
	}
}

func (e *Validation) Error() string {
	return e.err.Error()
}

func (e *Validation) Namespace() string {
	return e.namespace
}
