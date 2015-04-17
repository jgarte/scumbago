package scumbag

import (
	"encoding/json"
	"os"
)

type BotConfig struct {
	Name     string
	Server   string
	Channel  string
	Database *DatabaseConfig
}

type DatabaseConfig struct {
	Name            string
	Host            string
	LinksCollection string
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
