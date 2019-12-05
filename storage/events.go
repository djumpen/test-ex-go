package storage

import (
	"context"

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

type balance struct {
	Total float64
}

type orderedEvent struct {
	models.Event
	RowNumber int
}

// Create is for adding new event
func (s *events) Create(ctx context.Context, e models.Event) error {
	if err := validateEventAmount(e); err != nil {
		return errors.WithStack(err)
	}
	err := withTransaction(s.db, func(tx *gorm.DB) error {
		bal, err := getBalanceWithLock(ctx, tx)
		if err != nil {
			return errors.WithStack(err)
		}
		totalBal := bal + e.Amount
		if totalBal < 0 {
			return errors.WithStack(errNegativeBalance)
		}
		if err := tx.Create(&e).Error; err != nil {
			return errors.WithStack(err)
		}
		if err := setBalance(ctx, tx, totalBal); err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	return errors.Wrap(err, "Storage error while creating event")
}

// CancelLastOddEvents cancel last odd given events and recalculate balance
func (s *events) CancelLastOddEvents(ctx context.Context, num int) error {
	err := withTransaction(s.db, func(tx *gorm.DB) error {
		bal, err := getBalanceWithLock(ctx, tx)
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
			canBal -= e.Amount
			cancelIDs = append(cancelIDs, e.ID)
		}
		totalBal := bal + canBal
		if totalBal < 0 {
			return errors.WithStack(errNegativeBalance)
		}
		if err = cancelEventsByIDs(ctx, tx, cancelIDs); err != nil {
			return errors.WithStack(errCancellation)
		}
		if err := setBalance(ctx, tx, totalBal); err != nil {
			return errors.WithStack(err)
		}
		return nil
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
	if len(ids) == 0 {
		return nil
	}
	err := tx.Table("events").Where("id IN (?)", ids).Updates(map[string]interface{}{"status": models.StatusCanceled}).Error
	return errors.WithStack(err)
}

func getBalanceWithLock(_ context.Context, tx *gorm.DB) (float64, error) {
	var res balance
	err := tx.Raw("SELECT total FROM balance WHERE id = ? FOR UPDATE", 1).
		Scan(&res).Error
	if err != nil {
		return 0, errors.Wrap(err, "Can't get balance")
	}
	return res.Total, nil
}

func setBalance(_ context.Context, tx *gorm.DB, total float64) error {
	err := tx.Table("balance").Where("id = ?", 1).
		Updates(map[string]interface{}{"total": total}).Error
	if err != nil {
		return errors.Wrap(err, "Can't update balance")
	}
	return nil
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
