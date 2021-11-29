package main

import (
	"SB/catb406.github.io/service/config"
	"SB/catb406.github.io/service/repository/db"
	"SB/catb406.github.io/service/repository/messenger"
	"SB/catb406.github.io/service/repository/persistence"
	"SB/catb406.github.io/service/repository/token"
	"SB/catb406.github.io/service/repository/training"
	"SB/catb406.github.io/service/service"
	"flag"
	"fmt"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func main() {
	address := flag.String("address", "", "address to listen")
	port := flag.String("port", "3000", "port to listen")

	flag.Parse()

	log.Info(fmt.Sprintf(
		"address: %s, port: %s", *address, *port,
	))

	dsn := fmt.Sprintf("user=%s password='%s' dbname=%s sslmode=disable", config.PostgresUser, config.PostgresPassword, config.PostgresDbName)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		for i := 1; i < 10; i++ {
			if err == nil {
				break
			}
			log.Error(err)
			time.Sleep(config.DbConnectTimeout)
			log.Info("Trying to reconnect...")
			database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		}
		log.Fatal(err)
	}

	persistent := persistence.NewPersistent(database)
	usrMgr := db.NewDbManager(persistent)
	tknMgr := token.NewTokenManager(persistent)
	trainingMgr := training.NewTrainingManager(persistent)
	messenger := messenger.NewMessenger(persistent)
	server := service.NewServer(*address, *port, usrMgr, tknMgr, trainingMgr, messenger)
	server.Start()
}
