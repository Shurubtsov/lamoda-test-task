package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Logging logging
	Storage storage
	Service service
}

type logging struct {
	Level  int    `env:"LEVEL"`
	Output string `env:"OUTPUT"`
}

type storage struct {
	Username string `env:"DB_USERNAME"`
	Password string `env:"DB_PASSWORD"`
	Port     string `env:"DB_PORT"`
	Database string `env:"DB_DATABASE"`
	Host     string `env:"DB_HOST"`
}

type service struct {
	Address          string `env:"ADDRESS"`
	MigrationVersion uint   `env:"MIGRATION_VERSION"`
	MigrationsPath   string `env:"MIGRATIONS_PATH"`
}

var (
	once     sync.Once
	instance *Config
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}

		if err := cleanenv.ReadEnv(instance); err != nil {
			log.Println("Can't read environment")
		}

	})
	return instance
}
