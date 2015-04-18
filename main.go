package main

import (
	"flag"

	"github.com/Oshuma/scumbago/scumbag"
)

func main() {
	configFile := flag.String("config", scumbag.CONFIG_FILE, "Bot config JSON file")
	logFilename := flag.String("log", scumbag.LOG_FILE, "Bot log file")
	flag.Parse()

	bot := scumbag.NewBot(configFile, logFilename)
	bot.Start()
	defer bot.Shutdown()
}
