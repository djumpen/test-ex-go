package main

import (
	"log"
	"time"

	"github.com/djumpen/test-ex-go/config"
	"github.com/djumpen/test-ex-go/services"
	"github.com/djumpen/test-ex-go/storage"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.GetConfig()

	gormDB, err := gorm.Open("postgres", config.GetPostgresConnection())
	if err != nil {
		log.Fatal(err)
	}

	eventsStorage := storage.NewEvents(gormDB)
	eventsSvc := services.NewEvents(eventsStorage)

	numCancelRecords := 10
	if cfg.CancellationSelfRepeat {
		eventsSvc.RepeatCancellationTask(time.Duration(cfg.RepeatCancellationEvery)*time.Minute, numCancelRecords)
		exit := make(chan struct{})
		<-exit
	} else {
		err := eventsSvc.ExecCancellation(numCancelRecords)
		if err != nil {
			log.Print(err) // TODO: error logging
		}
	}
}
