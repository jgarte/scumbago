package scumbag

import (
	"encoding/json"
	"os"
)

type BotConfig struct {
	Name     string
	Server   string
	Channel  string
	Admins   []string
	LogLevel string
	Database *DatabaseConfig
	Twitter  *TwitterConfig
}

type DatabaseConfig struct {
	Name     string
	User     string
	Password string
}

type TwitterConfig struct {
	AccessToken string
}

func LoadConfig(configFile *string) *BotConfig {
	var botConfig BotConfig

	file, err := os.Open(*configFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&botConfig); err != nil {
		panic(err)
	}

	return &botConfig
}
