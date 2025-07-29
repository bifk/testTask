package main

import (
	"context"
	"errors"
	"github.com/bifk/testTask/internal/config"
	"github.com/bifk/testTask/internal/dataBase/postgres"
	"github.com/bifk/testTask/internal/logger"
	"github.com/bifk/testTask/internal/server/handlers/transaction"
	"github.com/bifk/testTask/internal/server/handlers/wallet"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger.Init()
	l := logger.GetLogger()

	db, err := postgres.New()
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
	err = db.Init(&l)
	if err != nil {
		l.Error(err)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)

	router.Post("/api/send", wallet.Send(db, &l))
	router.Get("/api/wallet/{address}/balance", wallet.GetBalance(db, &l))
	router.Get("/api/transactions", transaction.GetLast(db, &l))

	l.Info("Запуск сервера")
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("Не удалось запустить сервер")
		}
	}()

	l.Info("Сервер стартовал по адрессу: " + cfg.Address)

	//Graceful shutdown
	<-stop
	l.Info("Остановка сервера")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Error("Не удается остановить сервер: " + err.Error())

		return
	}

	l.Info("Закрытие базы данных ")
	err = db.Stop()
	if err != nil {
		l.Error("Не удается закрыть базу данных: " + err.Error())

		return
	}

	l.Info("Сервер остановлен")

}
