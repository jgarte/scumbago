package main

import (
	"flag"

	"github.com/Oshuma/scumbago/scumbag"
)

func main() {
	configFile := flag.String("config", "config/bot.json", "Bot config JSON file")
	flag.Parse()

	bot := scumbag.NewBot(configFile)
	bot.Start()
	defer bot.Shutdown()
}
