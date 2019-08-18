package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Oshuma/scumbago/scumbag"
)

func main() {
	configFile := flag.String("config", scumbag.ConfigFile, "Bot config JSON file")
	environment := flag.String("env", "development", "App environment (development, production, etc)")
	logFilename := flag.String("log", scumbag.LogFile, "Bot log file")
	versionFlag := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", scumbag.VersionString())
		os.Exit(0)
	}

	bot, err := scumbag.NewBot(configFile, logFilename, environment)
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
