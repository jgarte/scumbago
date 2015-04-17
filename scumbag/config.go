package scumbag

type BotConfig struct {
	Name    string
	Server  string
	Channel string
	DB      *DatabaseConfig
}

type DatabaseConfig struct {
	Name            string
	Host            string
	LinksCollection string
}
