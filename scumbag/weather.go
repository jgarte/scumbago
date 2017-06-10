package scumbag

import (
	"encoding/json"
	"fmt"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	WEATHER_API_URL = "http://api.wunderground.com/api/%s/%s/q/%s.json"
)

var WEATHER_HELP = [...]string{
	"?weather <location/zip>",
	"?weather -forecast <location/zip>",
	"?weather -hourly <location/zip>",
}

type ConditionsResponse struct {
	Observation `json:"current_observation"`
}

type Observation struct {
	Temperature string `json:"temperature_string"`
	Humidity    string `json:"relative_humidity"`
}

type ForecastResponse struct {
	Forecast Forecast `json:"forecast"`
}

type Forecast struct {
	SimpleForecast SimpleForecast `json:"simpleforecast"`
}

type SimpleForecast struct {
	Day []ForecastDay `json:"forecastday"`
}

type ForecastDay struct {
	Date Date     `json:"date"`
	High HighTemp `json:"high"`
	Low  LowTemp  `json:"low"`
}

type Date struct {
	WeekdayShort string `json:"weekday_short"`
}

type HighTemp struct {
	Fahrenheit string `json:"fahrenheit"`
}

type LowTemp struct {
	Fahrenheit string `json:"fahrenheit"`
}

type HourlyResponse struct {
	Forecast []HourlyForecast `json:"hourly_forecast"`
}

type HourlyForecast struct {
	Time HourlyTime `json:"FCTTIME"`
	Temp HourlyTemp `json:"temp"`
}

type HourlyTime struct {
	Hour   string `json:"hour_padded"`
	Minute string `json:"min"`
}

type HourlyTemp struct {
	Fahrenheit string `json:"english"`
}

type WeatherCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

func NewWeatherCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *WeatherCommand {
	return &WeatherCommand{bot: bot, conn: conn, line: line}
}

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
		cmd.currentConditions(cmdArgs)
	} else {
		if len(cmdArgs) == 0 {
			return
		}

		switch cmdArgs[0] {
		case "-forecast":
			cmd.currentForecast(cmdArgs)
		case "-hourly":
			cmd.hourlyForecast(cmdArgs)
		default:
			cmd.bot.Msg(cmd.conn, channel, "?weather <location/zip>")
			cmd.bot.Msg(cmd.conn, channel, "?weather -forecast <location/zip>")
			cmd.bot.Msg(cmd.conn, channel, "?weather -hourly <location/zip>")
		}
	}
}

func (cmd *WeatherCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.Help()")
		return
	}

	for _, helpText := range WEATHER_HELP {
		cmd.bot.Msg(cmd.conn, channel, helpText)
	}
}

func (cmd *WeatherCommand) currentConditions(args []string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.currentConditions()")
		return
	}

	apiKey := cmd.bot.Config.WeatherUnderground.Key
	requestUrl := fmt.Sprintf(WEATHER_API_URL, apiKey, "conditions", args[0])
	cmd.bot.Log.WithField("requestUrl", requestUrl).Debug("WeatherCommand.currentConditions()")

	content, err := getContent(requestUrl)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.HandleWeatherCommand()")
		return
	}

	var result ConditionsResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.HandleWeatherCommand()")
		return
	}

	if result.Observation.Temperature != "" {
		msg := fmt.Sprintf("%s / %s humidity", result.Observation.Temperature, result.Observation.Humidity)
		cmd.bot.Msg(cmd.conn, channel, msg)
	} else {
		cmd.bot.Msg(cmd.conn, channel, "WTF zip code is that?")
	}
}

func (cmd *WeatherCommand) currentForecast(args []string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.currentForecast()")
		return
	}

	apiKey := cmd.bot.Config.WeatherUnderground.Key
	requestUrl := fmt.Sprintf(WEATHER_API_URL, apiKey, "forecast", args[1])
	cmd.bot.Log.WithField("requestUrl", requestUrl).Debug("WeatherCommand.currentForecast()")

	content, err := getContent(requestUrl)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.currentForecast()")
		return
	}

	var result ForecastResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.currentForecast()")
		return
	}

	var forecast []string
	for _, day := range result.Forecast.SimpleForecast.Day {
		forecast = append(forecast, fmt.Sprintf("%s:%s'/%s'", day.Date.WeekdayShort, day.High.Fahrenheit, day.Low.Fahrenheit))
	}

	msg := strings.Join(forecast, "  ")
	cmd.bot.Msg(cmd.conn, channel, msg)
}

func (cmd *WeatherCommand) hourlyForecast(args []string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.hourlyForecast()")
		return
	}

	apiKey := cmd.bot.Config.WeatherUnderground.Key
	requestUrl := fmt.Sprintf(WEATHER_API_URL, apiKey, "hourly", args[1])
	cmd.bot.Log.WithField("requestUrl", requestUrl).Debug("WeatherCommand.hourlyForecast()")

	content, err := getContent(requestUrl)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.hourlyForecast()")
		return
	}

	var result HourlyResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WeatherCommand.hourlyForecast()")
		return
	}

	var forecast []string
	for _, time := range result.Forecast {
		forecast = append(forecast, fmt.Sprintf("%s:%s (%s')", time.Time.Hour, time.Time.Minute, time.Temp.Fahrenheit))
	}

	// Just use the next 3 hours.
	forecast = forecast[:3]

	msg := strings.Join(forecast, "  ")
	cmd.bot.Msg(cmd.conn, channel, msg)
}
