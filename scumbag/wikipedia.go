package scumbag

import (
	"encoding/json"
	"fmt"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	wikiAPIURL = "https://en.wikipedia.org/w/api.php?action=opensearch&search=%s&format=json&limit=1&redirects=resolve"
	wikiHelp   = cmdPrefix + "wp <phrase>"
)

// WikiResult stores data returned from the API.
type WikiResult struct {
	Query   string
	Title   []string
	Content []string
	URL     []string
}

// WikiCommand interacts with the Wikipedia API.
type WikiCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewWikiCommand returns a new WikiCommand instance.
func NewWikiCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *WikiCommand {
	return &WikiCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *WikiCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WikiCommand.Run()")
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("WikiCommand.Run(): No query")
		return
	}

	encodedQuery := strings.Replace(query, " ", "%20", -1)
	requestURL := fmt.Sprintf(wikiAPIURL, encodedQuery)

	content, err := getContent(requestURL)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WikiCommand.Run()")
		return
	}

	var result WikiResult
	resultArray := []interface{}{&result.Query, &result.Title, &result.Content, &result.URL}

	err = json.Unmarshal(content, &resultArray)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WikiCommand.Run()")
		return
	}

	if len(result.Content) > 0 {
		cmd.bot.Msg(cmd.conn, channel, result.Content[0])
	}

	if len(result.URL) > 0 {
		cmd.bot.Msg(cmd.conn, channel, result.URL[0])
	}
}

// Help displays the command help.
func (cmd *WikiCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WikiCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, wikiHelp)
}
