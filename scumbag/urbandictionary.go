package scumbag

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	URBAN_DICT_API_URL        = "http://api.urbandictionary.com/v0/define?term=%s&page=1"
	URBAN_DICT_RANDOM_API_URL = "http://api.urbandictionary.com/v0/random?page=1"
)

type UrbanDictResult struct {
	Definitions []Definition `json:"list"`
	ResultType  string       `json:"result_type"`
	Sounds      []string     `json:"sounds"`
	Tags        []string     `json:"tags"`
}

type Definition struct {
	Author       string `json:"author"`
	CurrentVote  string `json:"current_vote"`
	DefinitionId int64  `json:"defid"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
	Permalink    string `json:"permalink"`
	ThumbsDown   int64  `json:"thumbs_down"`
	ThumbsUp     int64  `json:"thumbs_up"`
	Word         string `json:"word"`
}

// Used to sort definitions by 'ThumbsUp' value.
type ByThumbsUp []Definition

func (a ByThumbsUp) Len() int           { return len(a) }
func (a ByThumbsUp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByThumbsUp) Less(i, j int) bool { return a[i].ThumbsUp > a[j].ThumbsUp }

type UrbanDictionaryCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

func NewUrbanDictionaryCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *UrbanDictionaryCommand {
	return &UrbanDictionaryCommand{bot: bot, conn: conn, line: line}
}

func (cmd *UrbanDictionaryCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("UrbanDictionaryCommand.Run()")
		return
	}

	var requestUrl string

	query := args[0]

	if query == "" || query == "-random" {
		requestUrl = URBAN_DICT_RANDOM_API_URL
	} else {
		encodedQuery := strings.Replace(query, " ", "%20", -1)
		requestUrl = fmt.Sprintf(URBAN_DICT_API_URL, encodedQuery)
	}

	content, err := getContent(requestUrl)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("UrbanDictionaryCommand.Run()")
		return
	}

	var result UrbanDictResult
	err = json.Unmarshal(content, &result)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("UrbanDictionaryCommand.Run()")
		return
	}

	sort.Sort(ByThumbsUp(result.Definitions))

	if len(result.Definitions) > 0 {
		definition := result.Definitions[0]

		var message string
		if query == "-random" {
			message = fmt.Sprintf("%s: %s", definition.Word, definition.Definition)
		} else {
			message = definition.Definition
		}

		cmd.bot.Msg(cmd.conn, channel, message)
		cmd.bot.Msg(cmd.conn, channel, definition.Permalink)
	}
}
