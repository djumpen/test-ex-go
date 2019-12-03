package storage

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"database/sql"
	"log"

	"github.com/djumpen/test-ex-go/models"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestPsqlConfig struct {
	Username string
	Password string
	Database string
}

func setupPostgresContainer(ctx context.Context) (testcontainers.Container, *gorm.DB, error) {
	cfg := TestPsqlConfig{
		Username: "user",
		Password: "password",
		Database: "integration_db",
	}
	req := testcontainers.ContainerRequest{
		Image:        "postgres:9.6",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     cfg.Username,
			"POSTGRES_DB":       cfg.Database,
			"POSTGRES_PASSWORD": cfg.Password,
		},
		WaitingFor: wait.ForLog("LOG:  autovacuum launcher started"),
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, nil, err
	}

	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Database,
		port.Port(),
	)
	log.Println(connStr)
	time.Sleep(10 * time.Second) // Wait for postrgres launch
	gormDB, err := gorm.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	err = applyMigrations(gormDB.DB())
	if err != nil {
		return nil, nil, err
	}

	return postgresC, gormDB, nil
}

func applyMigrations(db *sql.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "../migrations",
	}
	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}
	if n > 0 {
		log.Printf("[Test] Applied %d migrations!\n", n)
	}
	return nil
}

// Integration test for checking non-negative balance
func TestEventCreate(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	postgresC, db, err := setupPostgresContainer(ctx)
	if postgresC != nil {
		defer postgresC.Terminate(ctx)
	}
	if err != nil {
		t.Error(err)
		return
	}

	eventsStorage := NewEvents(db)
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		go func(i int) {
			wg.Add(1)
			defer wg.Done()
			amount := rand.Float64() * 100
			var e models.Event
			if i%5 == 0 {
				e = genTestEvent(models.StateWin, amount)
			} else {
				e = genTestEvent(models.StateLoss, amount)
			}
			err = eventsStorage.Create(ctx, e)
			if err != nil && errors.Cause(err) != errNegativeBalance {
				t.Error(err)
			}
		}(i)
		t := time.Duration(rand.Int63n(5))
		time.Sleep(t * time.Millisecond)
	}

	wg.Wait()

	bal, err := calculateBalance(ctx, db)
	a.NoError(err)

	t.Logf("Total balance: %f", bal)

	if bal < 0 {
		t.Error("Negative balance")
	}
}

func TestCancelLastOddEvents(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	postgresC, db, err := setupPostgresContainer(ctx)
	if postgresC != nil {
		defer postgresC.Terminate(ctx)
	}
	if err != nil {
		t.Error(err)
		return
	}

	eventsStorage := NewEvents(db)

	assumeBalance := 520.

	for i := 0; i < 40; i++ {
		e := genTestEvent(models.StateWin, float64(i)+1)
		err := eventsStorage.Create(ctx, e)
		a.NoError(err)
	}

	err = eventsStorage.CancelLastOddEvents(ctx, 10)
	a.NoError(err)

	bal, err := calculateBalance(ctx, db)
	a.NoError(err)

	a.Equal(assumeBalance, bal)
}

func genTestEvent(state models.EventState, amount float64) models.Event {
	u := uuid.New()
	if state == models.StateLoss && amount > 0 {
		amount = amount * -1
	}
	return models.Event{
		State:         state,
		Amount:        amount,
		Status:        models.StatusProcessed,
		TransactionID: u.String(),
	}
}
