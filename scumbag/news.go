package scumbag

import (
	"errors"
	"fmt"
	"strings"

	irc "github.com/fluffle/goirc/client"
	newsapi "github.com/kaelanb/newsapi-go"
)

var (
	newsHelp = []string{
		"Get top headline:   " + cmdPrefix + "news",
		"Get topic headline: " + cmdPrefix + "news <topic>",
		"Search headlines:   " + cmdPrefix + "news /query/",
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

	switch {
	case args[0] == "":
		cmd.getTopHeadline()
	case args[0] == "-topics":
		cmd.msg(strings.Join(topics, ", "))
	case strings.HasPrefix(args[0], "/") && strings.HasSuffix(args[0], "/"):
		cmd.searchNews(args[0])
	default:
		cmd.getTopicHeadline(args[0])
	}
}

// Help shows the command help.
func (cmd *NewsCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("NewsCommand.Help()", err)
		return
	}

	for _, helpLine := range newsHelp {
		cmd.bot.Msg(cmd.conn, channel, helpLine)
	}
}

func (cmd *NewsCommand) getTopHeadline() {
	newsResponse, err := cmd.getNewsResponse()
	if err != nil {
		cmd.bot.LogError("NewsCommand.getTopHeadline()", err)
		return
	}

	cmd.msgArticle(newsResponse.Articles[0])
}

func (cmd *NewsCommand) getTopicHeadline(topic string) {
	if unknownTopic(topic) {
		cmd.msg("Unknown topic: " + topic)
		cmd.msg("Topics: " + strings.Join(topics, ", "))
		return
	}

	newsResponse, err := cmd.getNewsResponse("category=" + topic)
	if err != nil {
		cmd.bot.LogError("NewsCommand.getTopHeadline()", err)
		return
	}

	cmd.msgArticle(newsResponse.Articles[0])
}

func (cmd *NewsCommand) searchNews(arg string) {
	query := strings.Replace(arg, "/", "", 2)

	// The params are already url.PathEscape'd in the newsapi library, so we don't need to do it here.
	newsResponse, err := cmd.getNewsResponse("q=" + query)
	if err != nil {
		cmd.bot.LogError("NewsCommand.getTopHeadline()", err)
		return
	}

	cmd.msgArticle(newsResponse.Articles[0])
}

func (cmd *NewsCommand) getNewsResponse(params ...string) (*newsapi.NewsResponse, error) {
	query := []string{"country=us", "pageSize=1"}
	for _, param := range params {
		query = append(query, param)
	}

	newsResponse, err := cmd.bot.News.GetTopHeadlines(query)
	if err != nil {
		return nil, err
	}

	if len(newsResponse.Articles) <= 0 {
		return nil, errors.New("no articles found")
	}

	return newsResponse, nil
}

func (cmd *NewsCommand) msgArticle(article newsapi.Article) {
	headline := fmt.Sprintf("[%s] %s", article.Source.Name, article.Title)
	cmd.msg(headline)
	cmd.msg(article.URL)
}

func (cmd *NewsCommand) msg(message string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("NewsCommand.getTopHeadline()", err)
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
