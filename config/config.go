package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Environment string     `yaml:"environment"`
	DiscordBot  DiscordBot `yaml:"discord_bot"`
}

type DiscordBot struct {
	Token        string `yaml:"token"`
	Prefix       string `yaml:"prefix"`
	DefaultState bool   `yaml:"default_state"`
}

func MustSetupConfig() Config {
	env := os.Getenv("CONFIG_PATH")

	if env == "" {
		fmt.Println("CONFIG_PATH is not set")
		os.Exit(1)
	}

	if _, err := os.Stat(env); err != nil {
		fmt.Println("file in CONFIG_PATH is not exist")
		os.Exit(1)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(env, &cfg); err != nil {
		fmt.Println("failed to read config file in CONFIG_PATH")
		os.Exit(1)
	}

	return cfg
}
