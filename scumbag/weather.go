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

func (bot *Scumbag) HandleWeatherCommand(channel string, query string) {
	args := strings.Fields(query)

	if len(args) == 1 {
		currentConditions(bot, channel, args)
	} else {
		if len(args) == 0 {
			return
		}

		switch args[0] {
		case "-forecast":
			currentForecast(bot, channel, args)
		case "-hourly":
			hourlyForecast(bot, channel, args)
		default:
			bot.Msg(channel, "?weather <location/zip>")
			bot.Msg(channel, "?weather -forecast <location/zip>")
			bot.Msg(channel, "?weather -hourly <location/zip>")
		}
	}
}

func currentConditions(bot *Scumbag, channel string, args []string) {
	apiKey := bot.Config.WeatherUnderground.Key
	requestUrl := fmt.Sprintf(WEATHER_API_URL, apiKey, "conditions", args[0])
	bot.Log.WithField("requestUrl", requestUrl).Debug("currentConditions()")

	content, err := getContent(requestUrl)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleWeatherCommand()")
		return
	}

	var result ConditionsResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleWeatherCommand()")
		return
	}

	if result.Observation.Temperature != "" {
		msg := fmt.Sprintf("%s / %s humidity", result.Observation.Temperature, result.Observation.Humidity)
		bot.Msg(channel, msg)
	} else {
		bot.Msg(channel, "WTF zip code is that?")
	}
}

func currentForecast(bot *Scumbag, channel string, args []string) {
	apiKey := bot.Config.WeatherUnderground.Key
	requestUrl := fmt.Sprintf(WEATHER_API_URL, apiKey, "forecast", args[1])
	bot.Log.WithField("requestUrl", requestUrl).Debug("currentForecast()")

	content, err := getContent(requestUrl)
	if err != nil {
		bot.Log.WithField("error", err).Error("currentForecast()")
		return
	}

	var result ForecastResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		bot.Log.WithField("error", err).Error("currentForecast()")
		return
	}

	var forecast []string
	for _, day := range result.Forecast.SimpleForecast.Day {
		forecast = append(forecast, fmt.Sprintf("%s:%s'/%s'", day.Date.WeekdayShort, day.High.Fahrenheit, day.Low.Fahrenheit))
	}

	msg := strings.Join(forecast, "  ")
	bot.Msg(channel, msg)
}

func hourlyForecast(bot *Scumbag, channel string, args []string) {
	apiKey := bot.Config.WeatherUnderground.Key
	requestUrl := fmt.Sprintf(WEATHER_API_URL, apiKey, "hourly", args[1])
	bot.Log.WithField("requestUrl", requestUrl).Debug("hourlyForecast()")

	content, err := getContent(requestUrl)
	if err != nil {
		bot.Log.WithField("error", err).Error("hourlyForecast()")
		return
	}

	var result HourlyResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		bot.Log.WithField("error", err).Error("hourlyForecast()")
		return
	}

	var forecast []string
	for _, time := range result.Forecast {
		forecast = append(forecast, fmt.Sprintf("%s:%s (%s')", time.Time.Hour, time.Time.Minute, time.Temp.Fahrenheit))
	}

	// Just use the next 3 hours.
	forecast = forecast[:3]

	msg := strings.Join(forecast, "  ")
	bot.Msg(channel, msg)
}
