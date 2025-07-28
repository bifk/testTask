package main

import (
	"github.com/bifk/testTask/internal/config"
	"github.com/bifk/testTask/internal/dataBase/postgres"
	"github.com/bifk/testTask/internal/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger.Init()
	l := logger.GetLogger()
	l.Info("Запуск сервера")

	db, err := postgres.New(cfg.Connection)
	if err != nil {
		l.Error(err)
		os.Exit(1)
	}
	err = db.Init()
	if err != nil {
		l.Error(err)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	//router.Use(middleware.RealIP)

}
