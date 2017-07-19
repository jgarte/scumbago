package scumbag

import (
	"encoding/json"
	"fmt"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	weatherAPIURL = "http://api.wunderground.com/api/%s/%s/q/%s.json"
)

var weatherHelp = [...]string{
	cmdPrefix + "weather <location/zip>",
	cmdPrefix + "weather -forecast <location/zip>",
	cmdPrefix + "weather -hourly <location/zip>",
}

// ConditionsResponse stores the weather condition.
type ConditionsResponse struct {
	Observation `json:"current_observation"`
}

// Observation stores the condition's observation data.
type Observation struct {
	Temperature string `json:"temperature_string"`
	Humidity    string `json:"relative_humidity"`
}

// ForecastResponse stores the forecast data.
type ForecastResponse struct {
	Forecast Forecast `json:"forecast"`
}

// Forecast stores the simple forecase data.
type Forecast struct {
	SimpleForecast SimpleForecast `json:"simpleforecast"`
}

// SimpleForecast stores the day's forecast.
type SimpleForecast struct {
	Day []ForecastDay `json:"forecastday"`
}

// ForecastDay stores the day's temp data.
type ForecastDay struct {
	Date Date     `json:"date"`
	High HighTemp `json:"high"`
	Low  LowTemp  `json:"low"`
}

// Date stores weather date data.
type Date struct {
	WeekdayShort string `json:"weekday_short"`
}

// HighTemp stores data for the high temp.
type HighTemp struct {
	Fahrenheit string `json:"fahrenheit"`
}

// LowTemp stores data for the low temp.
type LowTemp struct {
	Fahrenheit string `json:"fahrenheit"`
}

// HourlyResponse stores data from the hourly forecast API.
type HourlyResponse struct {
	Forecast []HourlyForecast `json:"hourly_forecast"`
}

// HourlyForecast stores data for the hourly forecast.
type HourlyForecast struct {
	Time HourlyTime `json:"FCTTIME"`
	Temp HourlyTemp `json:"temp"`
}

// HourlyTime stores time data.
type HourlyTime struct {
	Hour   string `json:"hour_padded"`
	Minute string `json:"min"`
}

// HourlyTemp stores data for an hourly forecast.
type HourlyTemp struct {
	Fahrenheit string `json:"english"`
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
		case "-hourly":
			cmd.hourlyForecast(channel, cmdArgs)
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
	apiKey := cmd.bot.Config.WeatherUnderground.Key
	requestURL := fmt.Sprintf(weatherAPIURL, apiKey, "conditions", args[0])
	cmd.bot.Log.WithField("requestUrl", requestURL).Debug("WeatherCommand.currentConditions()")

	content, err := getContent(requestURL)
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

func (cmd *WeatherCommand) currentForecast(channel string, args []string) {
	apiKey := cmd.bot.Config.WeatherUnderground.Key
	requestURL := fmt.Sprintf(weatherAPIURL, apiKey, "forecast", args[1])
	cmd.bot.Log.WithField("requestUrl", requestURL).Debug("WeatherCommand.currentForecast()")

	content, err := getContent(requestURL)
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

func (cmd *WeatherCommand) hourlyForecast(channel string, args []string) {
	apiKey := cmd.bot.Config.WeatherUnderground.Key
	requestURL := fmt.Sprintf(weatherAPIURL, apiKey, "hourly", args[1])
	cmd.bot.Log.WithField("requestUrl", requestURL).Debug("WeatherCommand.hourlyForecast()")

	content, err := getContent(requestURL)
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
