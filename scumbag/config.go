package scumbag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type BotConfig struct {
	Servers            []*ServerConfig
	Admins             []string
	LogLevel           string
	Database           *DatabaseConfig
	Twitter            *TwitterConfig
	WeatherUnderground *WeatherUndergroundConfig
}

type ServerConfig struct {
	Name     string
	Server   string
	SSL      bool
	Channels []string
}

type DatabaseConfig struct {
	Name     string
	User     string
	Password string
}

type TwitterConfig struct {
	AccessToken string
}

type WeatherUndergroundConfig struct {
	Key string
}

func LoadConfig(configFile *string) (*BotConfig, error) {
	var botConfig BotConfig

	file, err := os.Open(*configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&botConfig); err != nil {
		return nil, err
	}

	return &botConfig, nil
}

func (config *BotConfig) Server(server string) (*ServerConfig, error) {
	for _, serverConfig := range config.Servers {
		if serverConfig.Server == server {
			return serverConfig, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Unknown server: %s", server))
}
