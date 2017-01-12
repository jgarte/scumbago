package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Oshuma/scumbago/scumbag"
)

func main() {
	configFile := flag.String("config", scumbag.CONFIG_FILE, "Bot config JSON file")
	logFilename := flag.String("log", scumbag.LOG_FILE, "Bot log file")
	flag.Parse()

	bot, err := scumbag.NewBot(configFile, logFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		<-signalChannel
		bot.Shutdown()
	}()

	if err := bot.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	bot.Wait()
}
