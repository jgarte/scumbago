package scumbag

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	WEATHER_API_URL = "http://api.wunderground.com/api/%s/%s/q/%s.json"
)

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
	bot     *Scumbag
	channel string
	query   string
}

func (cmd *WeatherCommand) Run() {
	args := strings.Fields(cmd.query)

	if len(args) == 1 {
		cmd.currentConditions(args)
	} else {
		if len(args) == 0 {
			return
		}

		switch args[0] {
		case "-forecast":
			cmd.currentForecast(args)
		case "-hourly":
			cmd.hourlyForecast(args)
		default:
			cmd.bot.Msg(cmd.channel, "?weather <location/zip>")
			cmd.bot.Msg(cmd.channel, "?weather -forecast <location/zip>")
			cmd.bot.Msg(cmd.channel, "?weather -hourly <location/zip>")
		}
	}
}

// func (bot *Scumbag) HandleWeatherCommand(channel string, query string) {
//   args := strings.Fields(query)

//   if len(args) == 1 {
//     currentConditions(bot, channel, args)
//   } else {
//     if len(args) == 0 {
//       return
//     }

//     switch args[0] {
//     case "-forecast":
//       currentForecast(bot, channel, args)
//     case "-hourly":
//       hourlyForecast(bot, channel, args)
//     default:
//       bot.Msg(channel, "?weather <location/zip>")
//       bot.Msg(channel, "?weather -forecast <location/zip>")
//       bot.Msg(channel, "?weather -hourly <location/zip>")
//     }
//   }
// }

func (cmd *WeatherCommand) currentConditions(args []string) {
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
		cmd.bot.Msg(cmd.channel, msg)
	} else {
		cmd.bot.Msg(cmd.channel, "WTF zip code is that?")
	}
}

func (cmd *WeatherCommand) currentForecast(args []string) {
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
	cmd.bot.Msg(cmd.channel, msg)
}

func (cmd *WeatherCommand) hourlyForecast(args []string) {
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
	cmd.bot.Msg(cmd.channel, msg)
}
