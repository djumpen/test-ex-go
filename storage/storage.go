package storage

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"

	"github.com/pkg/errors"
)

// TLevel is Transaction level
type TLevel int

const (
	TLRepeatbleRead TLevel = iota + 1
	TLSerializable
)

func withTransaction(db *gorm.DB, f func(tx *gorm.DB) error) error {
	tx := db.Begin()
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}

func setTransactionLevel(tx *gorm.DB, lvl TLevel) error {
	switch lvl {
	case TLRepeatbleRead:
		return tx.Exec(`set transaction isolation level repeatable read`).Error
	case TLSerializable:
		return tx.Exec(`set transaction isolation level serializable`).Error
	}
	return errors.New("Unknown transaction level")
}

type errorChecker func(err error) bool

func onSerializationFailures(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		switch err.Get('C') {
		case "40001", "40P01":
			return true
		}
	}
	return false
}

func retryWithStrategy(check errorChecker, attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}
		if !check(err) {
			return err
		}
		if i >= (attempts - 1) {
			break
		}
		time.Sleep(sleep)
		log.Printf("retrying after error[%d]: %v", i, err)
	}
	return errors.Wrapf(err, "after %d attempts", attempts)
}
