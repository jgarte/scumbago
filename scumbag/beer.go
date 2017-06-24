package scumbag

import (
	"encoding/json"
	"fmt"
	"net/url"

	irc "github.com/fluffle/goirc/client"
)

const (
	BREWERYDB_URL = "http://api.brewerydb.com/v2/search?type=beer&withBreweries=Y&key=%s&q=%s"

	BEER_HELP = "?beer <query>"
)

type BeerCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

type BreweryDBResult struct {
	Status        string     `json:"status"`
	NumberOfPages int        `json:"numberOfPages"`
	Beers         []BeerData `json:"data"`
}

type BeerData struct {
	ABV         string        `json:"abv"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Breweries   []BreweryData `json:"breweries"`
	Style       StyleData     `json:"style"`
}

type BreweryData struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}

type StyleData struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

func NewBeerCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *BeerCommand {
	return &BeerCommand{bot: bot, conn: conn, line: line}
}

func (cmd *BeerCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.currentConditions()")
		return
	}

	beerQuery := args[0]
	if beerQuery == "" {
		cmd.bot.Log.Debug("BeerCommand.Run(): No query")
		cmd.Help()
		return
	}
	beerQuery = url.QueryEscape(beerQuery)

	requestUrl := fmt.Sprintf(BREWERYDB_URL, cmd.bot.Config.BreweryDB.Key, beerQuery)

	content, err := getContent(requestUrl)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("BeerCommand.Run()")
		return
	}

	var result BreweryDBResult
	err = json.Unmarshal(content, &result)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("BeerCommand.Run()")
		return
	}

	if len(result.Beers) <= 0 {
		cmd.bot.Msg(cmd.conn, channel, "No beers found.")
		return
	}
	beer := result.Beers[0]

	var message string
	if len(beer.Breweries) > 0 {
		brewery := beer.Breweries[0]
		message = fmt.Sprintf("%s [%s] (%s%% ABV) by %s (%s)", beer.Name, beer.Style.ShortName, beer.ABV, brewery.Name, brewery.Website)
	} else {
		message = fmt.Sprintf("%s [%s] (%s%% ABV)", beer.Name, beer.Style.ShortName, beer.ABV)
	}

	cmd.bot.Msg(cmd.conn, channel, message)
}

func (cmd *BeerCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("BeerCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, BEER_HELP)
}
