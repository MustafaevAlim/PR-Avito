package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	Postgre PostgreConfig
}

type ServerConfig struct {
	Port string
}

type PostgreConfig struct {
	Password string
	User     string
	DBName   string
	Host     string
}

func NewConfig() (*Config, error) {

	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	c := viper.New()
	c.AutomaticEnv()

	return &Config{
		Server: ServerConfig{
			Port: c.GetString("PORT"),
		},
		Postgre: PostgreConfig{
			Password: c.GetString("POSTGRES_PASSWORD"),
			User:     c.GetString("POSTGRES_USER"),
			Host:     c.GetString("POSTGRES_HOST"),
			DBName:   c.GetString("POSTGRES_DB"),
		},
	}, nil
}
