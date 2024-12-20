package config

import (
	"log"
	"os"
	"time"

	"github.com/chistyakoviv/converter/internal/model"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-required:"true"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Postgres   Postgres   `yaml:"database"`
	Task       Task       `yaml:"task"`
	Image      Image      `yaml:"image"`
	Video      Video      `yaml:"video"`
	Defaults   *Defaults  `env:"-"`
}

type Postgres struct {
	Dsn string `yaml:"dsn" env:"POSTGRES_DSN" env-required:"true"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env:"ADDRESS" env-default:"0.0.0.0:80"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-required:"true"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-required:"true"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-required:"true"`
}

type Task struct {
	CheckTimeout time.Duration `yaml:"check_timeout" env:"TASK_CHECK_TIMEOUT" env-default:"5m"`
}

type Image struct {
	Threads int `yaml:"threads" env:"IMAGE_THREADS" env-default:"1"`
}

type Video struct {
	Threads int `yaml:"threads" env:"VIDEO_THREADS" env-default:"1"`
}

type Defaults struct {
	Image ImageDefaults `yaml:"image"`
	Video VideoDefaults `yaml:"video"`
}

type ImageDefaults struct {
	Formats []model.ConvertTo `yaml:"formats"`
}

type VideoDefaults struct {
	Formats []model.ConvertTo `yaml:"formats"`
}

// Functions that start with the Must prefix require that the config is loaded, otherwise panic will be thrown
func MustLoad(args ...interface{}) *Config {
	var (
		cfg          Config
		dfs          Defaults
		configPath   string
		defaultsPath string
	)

	// First argument is config path
	// Second argument is defaults path
	switch len(args) {
	case 2:
		defaultsPath = args[1].(string)
		fallthrough
	case 1:
		configPath = args[0].(string)
	}

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

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

	if defaultsPath == "" {
		defaultsPath = os.Getenv("DEFAULTS_PATH")
	}

	// Defaults are empty by default and are left unchanged if no path is provided
	if defaultsPath != "" {
		if _, err := os.Stat(defaultsPath); os.IsNotExist(err) {
			log.Fatalf("file with defaults %s does not exist", defaultsPath)
		}

		if err := cleanenv.ReadConfig(defaultsPath, &dfs); err != nil {
			log.Fatalf("failed to load file with defaults from %s: %v", defaultsPath, err)
		}
	}

	cfg.Defaults = &dfs

	return &cfg
}
