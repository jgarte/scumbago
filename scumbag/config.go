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
	BreweryDB          *BreweryDBConfig
	Twitter            *TwitterConfig
	WeatherUnderground *WeatherUndergroundConfig
	WolframAlpha       *WolframAlphaConfig
}

type ServerConfig struct {
	Name     string
	Server   string
	SSL      bool
	Channels []string
}

type DatabaseConfig struct {
	Host     string
	SSL      string
	Name     string
	User     string
	Password string
}

type BreweryDBConfig struct {
	Key string
}

type TwitterConfig struct {
	AccessToken string
}

type WeatherUndergroundConfig struct {
	Key string
}

type WolframAlphaConfig struct {
	AppID string
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
