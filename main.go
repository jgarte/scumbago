package main

import "github.com/Oshuma/scumbago/scumbag"

func main() {
	bot := scumbag.NewBot()
	bot.Start()
	defer bot.Shutdown()
}
