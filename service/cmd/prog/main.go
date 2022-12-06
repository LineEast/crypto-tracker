package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LineEast/crypto-tracker/service/internal/collector"
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

	// Подключение к базе
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		panic(err)
	}

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
