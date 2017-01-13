package scumbag

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
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

func (bot *Scumbag) HandleUrbanDictCommand(channel string, query string) {
	var requestUrl string

	if query == "-random" {
		requestUrl = URBAN_DICT_RANDOM_API_URL
	} else {
		encodedQuery := strings.Replace(query, " ", "%20", -1)
		requestUrl = fmt.Sprintf(URBAN_DICT_API_URL, encodedQuery)
	}

	response, err := http.Get(requestUrl)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleUrbanDictCommand()")
		return
	}

	content, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleUrbanDictCommand()")
		return
	}

	var result UrbanDictResult
	err = json.Unmarshal(content, &result)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleUrbanDictCommand()")
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

		bot.Msg(channel, message)
		bot.Msg(channel, definition.Permalink)
	}
}
