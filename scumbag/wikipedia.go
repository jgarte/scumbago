package scumbag

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	WIKI_API_URL = "https://en.wikipedia.org/w/api.php?action=opensearch&search=%s&format=json&limit=1&redirects=resolve"
)

type WikiResult struct {
	Query   string
	Title   []string
	Content []string
	URL     []string
}

type WikiCommand struct {
	bot     *Scumbag
	channel string
}

func (cmd *WikiCommand) Run(args ...string) {
	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("WikiCommand.Run(): No query")
		return
	}

	encodedQuery := strings.Replace(query, " ", "%20", -1)
	requestUrl := fmt.Sprintf(WIKI_API_URL, encodedQuery)

	content, err := getContent(requestUrl)
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
		cmd.bot.Msg(cmd.channel, result.Content[0])
	}

	if len(result.URL) > 0 {
		cmd.bot.Msg(cmd.channel, result.URL[0])
	}
}
