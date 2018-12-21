package scumbag

import (
	"fmt"
	"strings"
	"time"

	"github.com/apixu/apixu-go"
	irc "github.com/fluffle/goirc/client"
	log "github.com/sirupsen/logrus"
)

const forecastDays = 3

var weatherHelp = [...]string{
	cmdPrefix + "weather <location/zip>",
	cmdPrefix + "weather -forecast <location/zip>",
}

// WeatherCommand interacts with the Weather Underground API.
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
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.Run()")
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("WeatherCommand.Run(): No query")
		return
	}

	cmdArgs := strings.Fields(query)

	if len(cmdArgs) == 1 {
		cmd.currentConditions(channel, cmdArgs)
	} else {
		if len(cmdArgs) == 0 {
			return
		}

		switch cmdArgs[0] {
		case "-forecast":
			cmd.currentForecast(channel, cmdArgs)
		default:
			cmd.Help()
		}
	}
}

// Help displays the command help.
func (cmd *WeatherCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.Help()")
		return
	}

	for _, helpText := range weatherHelp {
		cmd.bot.Msg(cmd.conn, channel, helpText)
	}
}

func (cmd *WeatherCommand) currentConditions(channel string, args []string) {
	client, err := cmd.getClient()
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.currentConditions()")
		return
	}

	current, err := client.Current(args[0])
	if err != nil {
		cmd.err(channel, err)
		return
	}

	msg := fmt.Sprintf("%.01f F / %.01f\" Precipitation / %d%% humidity",
		current.Current.TempFahrenheit,
		current.Current.PrecipIN,
		current.Current.Humidity,
	)
	cmd.bot.Msg(cmd.conn, channel, msg)
}

func (cmd *WeatherCommand) currentForecast(channel string, args []string) {
	client, err := cmd.getClient()
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.currentForecast()")
		return
	}

	forecast, err := client.Forecast(args[1], forecastDays)
	if err != nil {
		cmd.err(channel, err)
		return
	}

	var fcOut []string
	for _, fc := range forecast.Forecast.ForecastDay {
		fcOut = append(fcOut, fmt.Sprintf("%s:%.0f'/%.0f'", time.Time(fc.Date).Weekday(), fc.Day.MinTempFahrenheit, fc.Day.MaxTempFahrenheit))
	}

	msg := strings.Join(fcOut, "  ")
	cmd.bot.Msg(cmd.conn, channel, msg)
}

func (cmd *WeatherCommand) getClient() (apixu.Apixu, error) {
	return apixu.New(apixu.Config{APIKey: cmd.bot.Config.APIXU.Key})
}

func (cmd *WeatherCommand) err(channel string, err error) {
	cmd.bot.Log.WithField("error", err).Error("WeatherCommand.err()")
	if e, ok := err.(*apixu.Error); ok {
		cmd.bot.Log.WithFields(log.Fields{"code": e.Response().Code, "message": e.Response().Message}).Error("WeatherCommand.currentConditions()")
		cmd.bot.Msg(cmd.conn, channel, e.Response().Message)
	}
}
