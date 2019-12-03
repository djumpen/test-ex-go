package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/djumpen/test-ex-go/api"
	"github.com/djumpen/test-ex-go/config"
	"github.com/djumpen/test-ex-go/middleware"
	"github.com/djumpen/test-ex-go/services"
	"github.com/djumpen/test-ex-go/storage"
	"github.com/djumpen/test-ex-go/validation"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

func main() {
	cfg := config.GetConfig()

	if cfg.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	gormDB, err := gorm.Open("postgres", config.GetPostgresConnection())
	if err != nil {
		log.Fatal(err)
	}

	err = applyMigrations(gormDB.DB())
	if err != nil {
		log.Fatal(err)
	}

	// Upgrade gin validator
	binding.Validator = new(validation.DefaultValidator)

	responder := api.NewResponder()

	r := gin.Default()
	r.RedirectTrailingSlash = true

	r.Use(
		cors.New(api.GetCorsConfig()),
		middleware.ErrorHandler(responder),
	)

	rValidHeader := r.Group("/",
		middleware.ValidateSourceType(responder),
	)

	eventsStorage := storage.NewEvents(gormDB)
	eventsSvc := services.NewEvents(eventsStorage)

	commonRes := api.NewCommonResource(responder)
	eventsRes := api.NewEventsResource(eventsSvc, responder)

	// Setup routes
	rValidHeader.POST("/event", eventsRes.ProcessNewEvent)
	r.GET("/health", commonRes.Health)
	r.NoRoute(commonRes.NotFound)

	// Run cancellation task
	eventsSvc.RunCancellationTask()

	useSSL := cfg.CertFile != "" && cfg.KeyFile != ""
	address := fmt.Sprintf(":%d", cfg.Port)

	if useSSL {
		if err := http.ListenAndServeTLS(address, cfg.CertFile, cfg.KeyFile, r); err != nil {
			log.Fatalf("error in ListenAndServe: %s", err)
		}
	} else {
		log.Printf("Listening on port %d\n", cfg.Port)
		if err := http.ListenAndServe(address, r); err != nil {
			log.Fatalf("error in ListenAndServe: %s", err)
		}
	}
}

func applyMigrations(db *sql.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}
	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}
	if n > 0 {
		log.Printf("Applied %d migrations!\n", n)
	}
	return nil
}
