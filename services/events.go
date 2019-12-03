package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/djumpen/test-ex-go/models"
	"github.com/pkg/errors"
)

const defaultEventStatus = models.StatusProcessed

type eventsStorage interface {
	Create(context.Context, models.Event) error
	CancelLastOddEvents(context.Context, int) error
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

var once sync.Once

// TODO: add context to be able to cancel a task
func (s *events) RunCancellationTask() {
	once.Do(func() {
		go func() {
			for range time.Tick(10 * time.Minute) {
				err := s.st.CancelLastOddEvents(context.Background(), 10)
				if err != nil {
					log.Print(err) // TODO: error logging
				}
			}
		}()
	})
}
