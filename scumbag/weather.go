package scumbag

import (
	"strconv"

	owm "github.com/briandowns/openweathermap"
	irc "github.com/fluffle/goirc/client"
)

const (
	weatherUnit        = "F"
	weatherLang        = "EN"
	weatherCountryCode = "US"
)

var weatherHelp = [...]string{
	cmdWeather + " <location>",
}

// WeatherCommand interacts with the OpenWeatherMap API.
type WeatherCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewWeatherCommand returns a new WeatherCommand instance.
func NewWeatherCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *WeatherCommand {
	return &WeatherCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *WeatherCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("WeatherCommand.Run()", err)
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("WeatherCommand.Run(): No query")
		cmd.Help()
		return
	}

	zip, err := strconv.Atoi(query)
	if err != nil {
		cmd.bot.LogError("WeatherCommand.Run()", err)
		cmd.Help()
		return
	}

	cmd.currentConditions(channel, zip)
}

// Help displays the command help.
func (cmd *WeatherCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("WeatherCommand.Help()", err)
		return
	}

	for _, helpText := range weatherHelp {
		cmd.bot.Msg(cmd.conn, channel, helpText)
	}
}

func (cmd *WeatherCommand) currentConditions(channel string, zip int) {
	apiKey := cmd.bot.Config.OWM.Key
	w, err := owm.NewCurrent(weatherUnit, weatherLang, apiKey)
	if err != nil {
		cmd.bot.LogError("WeatherCommand.currentConditions()", err)
	}

	err = w.CurrentByZip(zip, weatherCountryCode)
	if err != nil {
		cmd.bot.LogError("WeatherCommand.currentConditions()", err)
	}

	cmd.bot.Msg(cmd.conn, channel, "%.01f F / %.01f\" Precipitation / %d%% humidity",
		w.Main.Temp,
		w.Rain.OneH,
		w.Main.Humidity,
	)
}
