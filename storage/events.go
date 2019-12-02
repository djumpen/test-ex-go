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

// Create is for adding new event
func (s *events) Create(ctx context.Context, e models.Event) error {
	t := time.Duration(rand.Int63n(500) + 5)
	if err := validateEventAmount(e); err != nil {
		return err
	}
	err := retryWithStrategy(onSerializationFailures, 10, t*time.Millisecond, func() error {
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
					return errors.New("Balance cannot be negative") // TODO: apperror
				}
			}
			return tx.Create(&e).Error
		})
	})
	return errors.WithStack(err)
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

func (s *events) CancelLastEvents(ctx context.Context, num int) error {
	// TODO
	return nil
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
