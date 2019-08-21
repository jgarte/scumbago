package scumbag

import (
	"testing"
)

func loadTestConfig() (*BotConfig, error) {
	configFile := "../config/bot.json.test"
	return LoadConfig(&configFile)
}

func TestLoadConfig(t *testing.T) {
	config, err := loadTestConfig()
	if err != nil {
		t.Errorf("Error loading config: %s", err)
	}

	if len(config.Servers) != 1 {
		t.Error("BotConfig.Servers not loaded properly")
	}

	if len(config.Admins) != 2 {
		t.Error("BotConfig.Admins not loaded properly")
	}

	if config.LogLevel != "Info" {
		t.Error("BotConfig.LogLevel not set properly")
	}
}

func TestServer(t *testing.T) {
	config, _ := loadTestConfig()

	server, err := config.Server("irc.example.com:6667")
	if err != nil {
		t.Errorf("Error getting server config: %s", err)
	}

	if server.Name != "scumbag_bot" {
		t.Error("ServerConfig.Name not set")
	}

	if server.SSL != true {
		t.Error("ServerConfig.SSL not set")
	}

	if len(server.Channels) <= 0 {
		t.Error("ServerConfig.Channels not set")
	}
}

func TestChannelConfigs(t *testing.T) {
	config, _ := loadTestConfig()
	server, _ := config.Server("irc.example.com:6667")

	channel := server.Channels["#scumbag"]
	if channel.SaveURLs != true {
		t.Error("ChannelConfig.SaveURLs not set properly")
	}

	channel = server.Channels["#scumbag_two"]
	if channel.SaveURLs != false {
		t.Error("ChannelConfig.SaveURLs not set properly")
	}
}

func TestDatabaseConfig(t *testing.T) {
	config, _ := loadTestConfig()
	db := config.Database

	if db.Host != "db.example.com" {
		t.Error("DatabaseConfig.Host not set")
	}

	if db.SSL != "disable" {
		t.Error("DatabaseConfig.SSL not set")
	}

	if db.Name != "scumbag" {
		t.Error("DatabaseConfig.Name not set")
	}

	if db.User != "database_user" {
		t.Error("DatabaseConfig.User not set")
	}

	if db.Password != "database_password" {
		t.Error("DatabaseConfig.Password not set")
	}
}
