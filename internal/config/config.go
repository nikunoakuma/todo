package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env                  string        `yaml:"env" env-required:"true"`
	ConnectionString     string        `yaml:"connection_string" env-required:"true"`
	AccessTokenTTl       time.Duration `yaml:"access_token_ttl" env-required:"true"`
	StandardQueryTimeout time.Duration `yaml:"standard_query_timeout" env-required:"true"`
	RequestTimeout       time.Duration `yaml:"request_timeout" env-required:"true"`
	HTTPServer           `yaml:"http-server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

func MustLoad() *Config {
	cfg := Config{}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("config path is not set")
	}

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("file doesn't exist: %s", configPath)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	return &cfg
}
