package storage

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"github.com/djumpen/test-ex-go/models"
	"github.com/jinzhu/gorm"
)

type events struct {
	db *gorm.DB
}

// NewEvents returns Events storage
func NewEvents(db *gorm.DB) *events {
	return &events{
		db: db,
	}
}

var (
	errNegativeBalance = errors.New("Balance cannot be negative")
	errCancellation    = errors.New("Cannot cancel last events due to low balance")
)

// Create is for adding new event
func (s *events) Create(ctx context.Context, e models.Event) error {
	t := time.Duration(rand.Int63n(400) + 100)
	if err := validateEventAmount(e); err != nil {
		return err
	}
	err := retryWithStrategy(onSerializationFailures, 15, t*time.Millisecond, func() error {
		return withTransaction(s.db, func(tx *gorm.DB) error {
			if err := setTransactionLevel(tx, TLSerializable); err != nil {
				return errors.WithStack(err)
			}
			if e.Amount < 0 {
				bal, err := calculateBalance(ctx, tx)
				if err != nil {
					return errors.WithStack(err)
				}
				if bal+e.Amount < 0 {
					return errNegativeBalance // TODO: apperror
				}
			}
			return tx.Create(&e).Error
		})
	})
	return errors.Wrap(err, "Storage error while creating event")
}

type orderedEvent struct {
	models.Event
	RowNumber int
}

func (s *events) CancelLastOddEvents(ctx context.Context, num int) error {
	err := retryWithStrategy(onSerializationFailures, 15, 200*time.Millisecond, func() error {
		return withTransaction(s.db, func(tx *gorm.DB) error {
			bal, err := calculateBalance(ctx, tx)
			if err != nil {
				return errors.WithStack(err)
			}
			events, err := getLastOrderedEvents(ctx, tx, num*2)
			if err != nil {
				return errors.Wrap(err, "Cannot get last events")
			}
			var canBal float64
			cancelIDs := make([]int, 0, num)
			for _, e := range events {
				// skip already canceled and EVEN records
				if e.Status == models.StatusCanceled || e.RowNumber%2 == 0 {
					continue
				}
				canBal += e.Amount
				cancelIDs = append(cancelIDs, e.ID)
			}
			totalBal := bal + canBal
			if totalBal < 0 {
				return errors.WithStack(errNegativeBalance)
			}
			err = cancelEventsByIDs(ctx, tx, cancelIDs)
			if err != nil {
				return errors.WithStack(errCancellation)
			}
			return nil
		})
	})
	return errors.Wrap(err, "Canceling events error")
}

func getLastOrderedEvents(_ context.Context, tx *gorm.DB, num int) ([]orderedEvent, error) {
	var events []orderedEvent
	err := tx.Raw(`
			SELECT *, ROW_NUMBER () OVER (ORDER BY id)
			FROM events ORDER BY id DESC LIMIT ?`, num).
		Find(&events).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return events, nil
}

func cancelEventsByIDs(_ context.Context, tx *gorm.DB, ids []int) error {
	err := tx.Table("events").Where("id IN (?)", ids).Updates(map[string]interface{}{"status": models.StatusCanceled}).Error
	return errors.WithStack(err)
}

func calculateBalance(_ context.Context, tx *gorm.DB) (float64, error) {
	var res struct {
		Amount float64
	}
	err := tx.Raw("SELECT SUM(amount) as Amount FROM events WHERE status = ?", models.StatusProcessed).
		Scan(&res).Error
	if err != nil {
		return 0, errors.Wrap(err, "Can't calculate balance")
	}
	return res.Amount, nil
}

func validateEventAmount(e models.Event) error {
	if e.State == models.StateLoss && e.Amount > 0 {
		return errors.New("Invlid amount")
	}
	if e.State == models.StateWin && e.Amount < 0 {
		return errors.New("Invlid amount")
	}
	return nil
}
