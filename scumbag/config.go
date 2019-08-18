package scumbag

import (
	"encoding/json"
	"fmt"
	"os"
)

// BotConfig contains the overall bot configuration.
type BotConfig struct {
	Servers      []*ServerConfig
	Admins       []string
	LogLevel     string
	APIXU        *APIXUConfig
	Database     *DatabaseConfig
	BreweryDB    *BreweryDBConfig
	IGDB         *IGDBConfig
	News         *NewsConfig
	OMDb         *OMDbConfig
	Rollbar      *RollbarConfig
	Twitter      *TwitterConfig
	WolframAlpha *WolframAlphaConfig
}

// ServerConfig stores IRC connection information.
type ServerConfig struct {
	Name     string
	Server   string
	SSL      bool
	Channels map[string]*ChannelConfig
}

// DatabaseConfig stores database connection information.
type DatabaseConfig struct {
	Host     string
	SSL      string
	Name     string
	User     string
	Password string
}

// ChannelConfig stores configuration information for a single channel.
type ChannelConfig struct {
	SaveURLs bool
}

// APIXUConfig stores the APIXU API information.
type APIXUConfig struct {
	Key string
}

// BreweryDBConfig stores BreweryDB API information.
type BreweryDBConfig struct {
	Key string
}

// IGDBConfig stores IGDB.com API information.
type IGDBConfig struct {
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

// RollbarConfig stores Rollbar config information.
type RollbarConfig struct {
	Token string
}

// TwitterConfig stores Twitter API information.
type TwitterConfig struct {
	AccessToken string
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
