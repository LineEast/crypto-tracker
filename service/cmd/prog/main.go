package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/LineEast/crypto-tracker/service/internal/collector"
	"github.com/LineEast/crypto-tracker/service/internal/database"
	"github.com/LineEast/crypto-tracker/service/internal/server"
)

type (
	config struct {
		DSN       string
		Collector collector.Config
	}
)

func main() {
	// Чтение конфига
	config := config{
		DSN: os.Getenv("DSN"),
		Collector: collector.Config{
			FiatEndPoint:   os.Getenv("FIAT_END_POINT"),
			CryptoEndPoint: os.Getenv("CRYPTO_END_POINT"),
		},
	}

	db, err := database.Conn(config.DSN)

	errs := make(chan error)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// создание Data_collector
	collector := collector.New(db, &config.Collector)

	// запуск Data_collector
	go func() {
		errs <- collector.Run()
	}()

	// Создание сервера
	server := server.New(db)

	// Запуск сервера
	go func() {
		errs <- server.Run()
	}()

	select {
	case err = <-errs:
		if err != nil {
			panic(err)
		}
	case <-signals:
	}
}
