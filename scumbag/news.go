package scumbag

import (
	"fmt"
	"strings"

	irc "github.com/fluffle/goirc/client"
	newsapi "github.com/kaelanb/newsapi-go"
)

var (
	newsHelp = []string{
		"Get top headline:   " + cmdPrefix + "news",
		"Get topic headline: " + cmdPrefix + "news <topic>",
		"List topics:        " + cmdPrefix + "news -topics",
	}

	topics = []string{
		"business",
		"entertainment",
		"general",
		"health",
		"science",
		"sports",
		"technology",
	}
)

// NewsCommand interacts with the News API.
type NewsCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewNewsCommand returns a new NewsCommand instance.
func NewNewsCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *NewsCommand {
	return &NewsCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *NewsCommand) Run(args ...string) {
	cmd.bot.Log.WithField("args", args).Debug("NewsCommand.Run()")

	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("NewsCommand.Run(): No args")
		return
	}

	topic := args[0]
	switch topic {
	case "":
		cmd.getTopHeadline()
	case "-topics":
		cmd.msg(strings.Join(topics, ", "))
	default:
		cmd.getHeadline(topic)
	}
}

// Help shows the command help.
func (cmd *NewsCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("NewsCommand.Help()")
		return
	}

	for _, helpLine := range newsHelp {
		cmd.bot.Msg(cmd.conn, channel, helpLine)
	}
}

func (cmd *NewsCommand) getTopHeadline() {
	query := []string{"country=us", "pageSize=1"}
	newsResponse, err := cmd.bot.News.GetTopHeadlines(query)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("NewsCommand.getTopHeadline")
		return
	}

	if len(newsResponse.Articles) <= 0 {
		cmd.bot.Log.WithField("newsResponse", newsResponse).Error("NewsCommand.getTopHeadline(): No articles.")
		return
	}

	cmd.msgArticle(newsResponse.Articles[0])
}

func (cmd *NewsCommand) getHeadline(topic string) {
	if unknownTopic(topic) {
		cmd.msg("Unknown topic: " + topic)
		cmd.msg("Topics: " + strings.Join(topics, ", "))
		return
	}

	query := []string{"country=us", "pageSize=1", "category=" + topic}
	newsResponse, err := cmd.bot.News.GetTopHeadlines(query)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("NewsCommand.getHeadline()")
		return
	}

	if len(newsResponse.Articles) <= 0 {
		cmd.bot.Log.WithField("newsResponse", newsResponse).Error("NewsCommand.getTopHeadline(): No articles.")
		return
	}

	cmd.msgArticle(newsResponse.Articles[0])
}

func (cmd *NewsCommand) msgArticle(article newsapi.Article) {
	headline := fmt.Sprintf("[%s] %s", article.Source.Name, article.Title)
	cmd.msg(headline)
	cmd.msg(article.URL)
}

func (cmd *NewsCommand) msg(message string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("NewsCommand.getTopHeadline()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, message)
}

func unknownTopic(topic string) bool {
	for _, t := range topics {
		if t == topic {
			return false
		}
	}
	return true
}
