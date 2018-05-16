package scumbag

import (
	"encoding/json"
	"fmt"
	"os"
)

// BotConfig contains the overall bot configuration.
type BotConfig struct {
	Servers            []*ServerConfig
	Admins             []string
	LogLevel           string
	Database           *DatabaseConfig
	BreweryDB          *BreweryDBConfig
	News               *NewsConfig
	OMDb               *OMDbConfig
	Twitter            *TwitterConfig
	WeatherUnderground *WeatherUndergroundConfig
	WolframAlpha       *WolframAlphaConfig
}

// ServerConfig stores IRC connection information.
type ServerConfig struct {
	Name     string
	Server   string
	SSL      bool
	Channels []string
}

// DatabaseConfig stores database connection information.
type DatabaseConfig struct {
	Host     string
	SSL      string
	Name     string
	User     string
	Password string
}

// BreweryDBConfig stores BreweryDB API information.
type BreweryDBConfig struct {
	Key string
}

// NewsConfig stores News API information.
type NewsConfig struct {
	Key string
}

// OMDbConfig stores the OMDb API information.
type OMDbConfig struct {
	Key string
}

// TwitterConfig stores Twitter API information.
type TwitterConfig struct {
	AccessToken string
}

// WeatherUndergroundConfig stores Weather Underground API information.
type WeatherUndergroundConfig struct {
	Key string
}

// WolframAlphaConfig stores Wolfram Alpha API information.
type WolframAlphaConfig struct {
	AppID string
}

// LoadConfig loads the configFile.
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

// Server returns the server config.
func (config *BotConfig) Server(server string) (*ServerConfig, error) {
	for _, serverConfig := range config.Servers {
		if serverConfig.Server == server {
			return serverConfig, nil
		}
	}

	return nil, fmt.Errorf("Unknown server: %s", server)
}
