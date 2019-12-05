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

// RunCancellationTask run cancellation task with self-repeat
func (s *events) RepeatCancellationTask(repeat time.Duration, number int) {
	once.Do(func() {
		go func() {
			for range time.Tick(repeat) {
				err := s.st.CancelLastOddEvents(context.TODO(), number)
				if err != nil {
					log.Print(err) // TODO: error logging
				}
			}
		}()
	})
}

func (s *events) ExecCancellation(number int) error {
	err := s.st.CancelLastOddEvents(context.TODO(), number)
	return errors.Wrap(err, "Events service cancellation error")
}
