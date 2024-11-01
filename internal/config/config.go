package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-required:"true"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Postgres   Postgres   `yaml:"database"`
}

type Postgres struct {
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"app"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"secret"`
	Db       string `yaml:"db" env:"POSTGRES_DB" env-default:"app"`
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
	ReadTimeout time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-required:"true"`
}

// Functions that start with the Must prefix require that the config is loaded, otherwise panic will be thrown
func MustLoad() *Config {
	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")

	if configPath != "" {
		// log.Fatal("CONFIG_PATH is not set")

		// check if file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("config file %s does not exist", configPath)
		}

		// Read from file
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("failed to load config from %s: %v", configPath, err)
		}
	}

	// Read from environment
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to load config from env: %v", err)
	}

	return &cfg
}
