package main

import (
	"flag"
	"notification/internal/pkg/config"
	"notification/internal/pkg/db"
	"notification/internal/pkg/log"
	"notification/internal/task"
	"time"
)

var Version = "1.0.0"

var flagConfig = flag.String("configs", "./configs/local.yml", "path to the configs file")

func main() {
	flag.Parse()
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		panic(err)
	}

	logger := log.NewLogger(cfg).Init()
	mysql, err := db.NewDB(cfg, logger).Open()
	defer mysql.Close()

	mongodb, err := db.NewMongoDB(cfg, logger).Open()
	defer mongodb.Logout()

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	doneChan := make(chan int)
	defer close(doneChan)

	task := task.NewTask(mysql, mongodb, logger, cfg)

	for {
		select {
		case <-ticker.C:
			task.SendNotification(doneChan)
		case id, ok := <-doneChan:
			if !ok {
				return
			}
			task.Done(id)
		}
	}
}
