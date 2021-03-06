package scumbag

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	urbanDictAPIURL       = "http://api.urbandictionary.com/v0/define?term=%s&page=1"
	urbanDictRandomAPIURL = "http://api.urbandictionary.com/v0/random?page=1"

	urbanDictHelp = cmdPrefix + "ud <phrase>"
)

// UrbanDictResult stores a UrbanDictionary API response.
type UrbanDictResult struct {
	Definitions []Definition `json:"list"`
	ResultType  string       `json:"result_type"`
	Sounds      []string     `json:"sounds"`
	Tags        []string     `json:"tags"`
}

// Definition is a word's definition.
type Definition struct {
	Author       string `json:"author"`
	CurrentVote  string `json:"current_vote"`
	DefinitionID int64  `json:"defid"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
	Permalink    string `json:"permalink"`
	ThumbsDown   int64  `json:"thumbs_down"`
	ThumbsUp     int64  `json:"thumbs_up"`
	Word         string `json:"word"`
}

// ByThumbsUp is used to sort definitions by 'ThumbsUp' value.
type ByThumbsUp []Definition

func (a ByThumbsUp) Len() int           { return len(a) }
func (a ByThumbsUp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByThumbsUp) Less(i, j int) bool { return a[i].ThumbsUp > a[j].ThumbsUp }

// UrbanDictionaryCommand interacts with the UrbanDictionary API.
type UrbanDictionaryCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewUrbanDictionaryCommand returns a new UrbanDictionaryCommand instance.
func NewUrbanDictionaryCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *UrbanDictionaryCommand {
	return &UrbanDictionaryCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *UrbanDictionaryCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("UrbanDictionaryCommand.Run()", err)
		return
	}

	var requestURL string

	query := args[0]

	if query == "" || query == "-random" {
		requestURL = urbanDictRandomAPIURL
	} else {
		encodedQuery := strings.Replace(query, " ", "%20", -1)
		requestURL = fmt.Sprintf(urbanDictAPIURL, encodedQuery)
	}

	content, err := getContent(requestURL)
	if err != nil {
		cmd.bot.LogError("UrbanDictionaryCommand.Run()", err)
		return
	}

	var result UrbanDictResult
	err = json.Unmarshal(content, &result)
	if err != nil {
		cmd.bot.LogError("UrbanDictionaryCommand.Run()", err)
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

// Help displays the command help.
func (cmd *UrbanDictionaryCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("UrbanDictionaryCommand.Help()", err)
		return
	}

	cmd.bot.Msg(cmd.conn, channel, urbanDictHelp)
}
