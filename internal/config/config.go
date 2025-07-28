package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Connection string `yaml:"connection" env-required:"true"`
	HTTPServer `yaml:"httpServer"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"127.0.0.1:8001"`
	Timeout     time.Duration `yaml:"timeout" env-default:"6s"`
	IdleTimeout time.Duration `yaml:"idleTimeout" env-default:"30s"`
}

func Load() (*Config, error) {
	configPath := "config/server.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, errors.New("Файл конфига не найден по следующему пути: " + configPath)
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return &Config{}, errors.New("Некоректный конфиг: " + configPath)
	}

	return &config, nil
}
