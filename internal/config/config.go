package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Logging logging
	Storage storage
}

type logging struct {
	Level  int    `yaml:"level" env:"LEVEL"`
	Output string `yaml:"output" env:"OUTPUT"`
}

type storage struct {
	Username string `yaml:"username" env:"DB_USERNAME"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	Port     string `yaml:"port" env:"DB_PORT"`
	Database string `yaml:"database" env:"DB_DATABASE"`
	Host     string `yaml:"host" env:"DB_HOST"`
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
