// Package config is a nice package
package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment string `yaml:"environment" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	Storage     `yaml:"storage"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	/* User string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"` */
}

type Storage struct {
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	DBName   string `yaml:"db_name" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-required:"true"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file: ", err)
	}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	var cfg Config

	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatal("error reading config file: ", err)
	}

	return &cfg
}
