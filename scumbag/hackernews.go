package scumbag

import (
	"encoding/json"
	"fmt"

	irc "github.com/fluffle/goirc/client"
)

const (
	hackerNewsAPIURL      = "https://hacker-news.firebaseio.com/v0"
	hackerNewsTopStories  = hackerNewsAPIURL + "/topstories.json"
	hackerNewsNewStories  = hackerNewsAPIURL + "/newstories.json"
	hackerNewsBestStories = hackerNewsAPIURL + "/beststories.json"
	hackerNewsItem        = hackerNewsAPIURL + "/item/%d.json"
)

var hackerNewsHelp = []string{
	"Get top story:    " + cmdPrefix + "hn",
	"Get newest story: " + cmdPrefix + "hn -new",
	"Get best story:   " + cmdPrefix + "hn -best",
}

// HackerNewsItem represents a single item (story) from Hacker News.
type HackerNewsItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Score int    `json:"score"`
	URL   string `json:"url"`

	// Type        string  `json:"type"`
	// By          string  `json:"by"`
	// Time        int64   `json:"time"`
	// Text        string  `json:"text"`
	// Deleted     bool    `json:"deleted"`
	// Parent      int64   `json:"parent"`
	// Kids        []int64 `json:"kids"`
	// Descendants int     `json:"descendants"`
}

// HackerNewsCommand interacts with the Hacker News API.
type HackerNewsCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewHackerNewsCommand returns a new HackerNewsCommand instance.
func NewHackerNewsCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *HackerNewsCommand {
	return &HackerNewsCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *HackerNewsCommand) Run(args ...string) {
	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("HackerNewsCommand.Run(): No args.")
		return
	}

	switch {
	case args[0] == "":
		cmd.getTopStory()
	case args[0] == "-new":
		cmd.getNewStory()
	case args[0] == "-best":
		cmd.getBestStory()
	default:
		cmd.getTopStory()
	}
}

// Help shows the command help.
func (cmd *HackerNewsCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.Help()")
		return
	}

	for _, helpLine := range hackerNewsHelp {
		cmd.bot.Msg(cmd.conn, channel, helpLine)
	}
}

func (cmd *HackerNewsCommand) getTopStory() {
	stories, err := cmd.getStories(hackerNewsTopStories)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getTopStory()")
		return
	}

	if len(stories) > 0 {
		item, err := cmd.getItem(stories[0])
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getTopStory()")
			return
		}
		cmd.msgItem(item)
	}
}

func (cmd *HackerNewsCommand) getNewStory() {
	stories, err := cmd.getStories(hackerNewsNewStories)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getNewStory()")
		return
	}

	if len(stories) > 0 {
		item, err := cmd.getItem(stories[0])
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getNewStory()")
			return
		}
		cmd.msgItem(item)
	}
}

func (cmd *HackerNewsCommand) getBestStory() {
	stories, err := cmd.getStories(hackerNewsBestStories)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getBestStory()")
		return
	}

	if len(stories) > 0 {
		item, err := cmd.getItem(stories[0])
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getBestStory()")
			return
		}
		cmd.msgItem(item)
	}
}

func (cmd *HackerNewsCommand) getStories(url string) ([]int64, error) {
	stories := make([]int64, 1)

	content, err := getContent(url)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getStories()")
		return nil, err
	}

	err = json.Unmarshal(content, &stories)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getStories()")
		return nil, err
	}

	return stories, nil
}

func (cmd *HackerNewsCommand) getItem(storyID int64) (HackerNewsItem, error) {
	storyURL := fmt.Sprintf(hackerNewsItem, storyID)

	content, err := getContent(storyURL)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getItem()")
		return HackerNewsItem{}, err
	}

	var item HackerNewsItem
	err = json.Unmarshal(content, &item)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.getItem()")
		return HackerNewsItem{}, err
	}

	return item, nil
}

func (cmd *HackerNewsCommand) msgItem(item HackerNewsItem) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HackerNewsCommand.msgItem()")
		return
	}

	cmd.bot.Log.WithField("item", item).Debug("HackerNewsCommand.msgItem()")

	message := fmt.Sprintf("[%d] %s", item.Score, item.Title)
	cmd.bot.Msg(cmd.conn, channel, message)
	cmd.bot.Msg(cmd.conn, channel, item.URL)
}
