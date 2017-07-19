package scumbag

import (
	"fmt"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	irc "github.com/fluffle/goirc/client"
)

const (
	twitterHelp = cmdPrefix + "twitter <@username> or <search_term>"
)

// TwitterCommand interacts with the Twitter API.
type TwitterCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewTwitterCommand returns a new TwitterCommand instance.
func NewTwitterCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *TwitterCommand {
	return &TwitterCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *TwitterCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("TwitterCommand.Run()")
		return
	}

	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("TwitterCommand.Run(): No args")
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("TwitterCommand.Run(): No query")
		return
	}

	switch {
	case strings.HasPrefix(query, "@"):
		user, protected := cmd.screennameStatus(query)

		var msg string
		if protected {
			msg = fmt.Sprintf("Account is not public: %s", query)
		} else {
			if user.ScreenName != "" && user.Status != nil {
				msg = fmt.Sprintf("@%s %s", user.ScreenName, user.Status.Text)
			} else {
				msg = fmt.Sprintf("Account has no tweets: %s", query)
			}
		}
		cmd.bot.Msg(cmd.conn, channel, msg)
	default:
		status := cmd.searchTwitter(query)
		if status != nil {
			msg := fmt.Sprintf("@%s %s", status.User.ScreenName, status.Text)
			cmd.bot.Msg(cmd.conn, channel, msg)
		}
	}
}

// Help displays the command help.
func (cmd *TwitterCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("TwitterCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, twitterHelp)
}

func (cmd *TwitterCommand) screennameStatus(query string) (*twitter.User, bool) {
	user, _, err := cmd.bot.Twitter.Users.Show(&twitter.UserShowParams{
		ScreenName:      strings.Replace(query, "@", "", 1),
		IncludeEntities: twitter.Bool(true),
	})
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("TwitterCommand.screennameStatus()")
		return nil, true
	}

	// Only return the User if their timeline is public.
	if user.Protected {
		cmd.bot.Log.WithField("user", user).Debug("TwitterCommand.screennameStatus(): User account is protected.")
		return user, true
	}

	return user, false
}

func (cmd *TwitterCommand) searchTwitter(query string) *twitter.Tweet {
	search, _, err := cmd.bot.Twitter.Search.Tweets(&twitter.SearchTweetParams{
		Query: query,
	})
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("TwitterCommand.searchTwitter()")
		return nil
	}

	if len(search.Statuses) > 0 {
		return &search.Statuses[0]
	}

	return nil
}
