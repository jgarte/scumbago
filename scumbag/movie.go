package scumbag

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

// MovieCommand interacts with the OMDb API.
type MovieCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// MovieSearchResult stores MovieSearchData.
type MovieSearchResult struct {
	Search []MovieSearchData `json:"Search"`
}

// MovieSearchData stores search result data.
type MovieSearchData struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"Movie"`
	Poster string `json:"Poster"`
}

// MovieData stores data about a single movie.
type MovieData struct {
	Title    string        `json:"Title"`
	Year     string        `json:"Year"`
	Rated    string        `json:"Rated"`
	Released string        `json:"Released"`
	Runtime  string        `json:"Runtime"`
	Genre    string        `json:"Genre"`
	Director string        `json:"Director"`
	Writer   string        `json:"Writer"`
	Actors   string        `json:"Actors"`
	Plot     string        `json:"Plot"`
	Ratings  []RatingsData `json:"Ratings"`
}

// RatingsData stores rating data about a single movie from a single source.
type RatingsData struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

const (
	omdbSearchURL = "http://www.omdbapi.com/?apikey=%s&s=%s"
	omdbImdbURL   = "http://www.omdbapi.com/?apikey=%s&i=%s"

	movieHelp = cmdPrefix + "movie <query>"
)

// NewMovieCommand returns a new MovieCommand instance.
func NewMovieCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *MovieCommand {
	return &MovieCommand{bot: bot, conn: conn, line: line}
}

// Run handles the movie searches.
func (cmd *MovieCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("MovieCommand.Run()")
		return
	}

	movieQuery := args[0]
	if movieQuery == "" {
		cmd.bot.Log.Debug("MovieCommand.Run(): No query")
		cmd.Help()
		return
	}
	movieQuery = url.QueryEscape(movieQuery)

	// First, we need to search to get the IMDB ID.
	searchRequestURL := fmt.Sprintf(omdbSearchURL, cmd.bot.Config.OMDb.Key, movieQuery)
	content, err := getContent(searchRequestURL)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("MovieCommand.Run()")
		return
	}

	var movieSearchResult MovieSearchResult
	err = json.Unmarshal(content, &movieSearchResult)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("MovieCommand.Run()")
		return
	}

	// Now we get the actual movie details.
	if len(movieSearchResult.Search) <= 0 {
		cmd.bot.Log.WithField("movieSearchResult", movieSearchResult).Debug("MovieCommand.Run(): No results")
		cmd.bot.Msg(cmd.conn, channel, "Beats me...")
		return
	}
	firstResult := movieSearchResult.Search[0]

	imdbRequestURL := fmt.Sprintf(omdbImdbURL, cmd.bot.Config.OMDb.Key, firstResult.ImdbID)
	content, err = getContent(imdbRequestURL)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("MovieCommand.Run()")
		return
	}

	var movieData MovieData
	err = json.Unmarshal(content, &movieData)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("MovieCommand.Run()")
		return
	}

	ratings := []string{}
	for _, rating := range movieData.Ratings {
		ratings = append(ratings, fmt.Sprintf("%s %s", rating.Source, rating.Value))
	}
	summary := fmt.Sprintf("%s (%s) (%s) (%s)", movieData.Title, movieData.Year, movieData.Genre, strings.Join(ratings, ", "))

	cmd.bot.Msg(cmd.conn, channel, summary)
	cmd.bot.Msg(cmd.conn, channel, movieData.Plot)
}

// Help shows the command help.
func (cmd *MovieCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("MovieCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, movieHelp)
}
