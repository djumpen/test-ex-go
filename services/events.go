package services

import (
	"context"

	"github.com/djumpen/test-ex-go/models"
	"github.com/pkg/errors"
)

const defaultEventStatus = models.StatusProcessed

type eventsStorage interface {
	Create(context.Context, models.Event) error
}

type events struct {
	st eventsStorage
}

// NewEvents creates new balance service
func NewEvents(st eventsStorage) *events {
	return &events{
		st: st,
	}
}

func (s *events) Create(ctx context.Context, e models.Event) (err error) {
	e.Status = defaultEventStatus
	err = s.st.Create(ctx, e)
	if err != nil {
		return errors.Wrap(err, "Events service can`t create event")
	}
	return err
}
