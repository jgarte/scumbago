package scumbag

import (
	"fmt"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
)

func (bot *Scumbag) HandleTwitterCommand(channel string, query string) {
	bot.Log.WithField("query", query).Debug("HandleTwitterCommand()")

	switch {
	case strings.HasPrefix(query, "@"):
		user := bot.screennameStatus(query)
		if user != nil {
			msg := fmt.Sprintf("@%s %s", user.ScreenName, user.Status.Text)
			bot.Msg(channel, msg)
		}
	default:
		status := bot.searchTwitter(query)
		if status != nil {
			msg := fmt.Sprintf("@%s %s", status.User.ScreenName, status.Text)
			bot.Msg(channel, msg)
		}
	}
}

func (bot *Scumbag) screennameStatus(screenname string) *twitter.User {
	includeEntities := true

	user, _, err := bot.Twitter.Users.Show(&twitter.UserShowParams{
		ScreenName:      strings.Replace(screenname, "@", "", 1),
		IncludeEntities: &includeEntities,
	})
	if err != nil {
		bot.Log.WithField("error", err).Error("screennameStatus()")
		return nil
	}

	return user
}

func (bot *Scumbag) searchTwitter(query string) *twitter.Tweet {
	search, _, err := bot.Twitter.Search.Tweets(&twitter.SearchTweetParams{
		Query: query,
	})
	if err != nil {
		bot.Log.WithField("error", err).Error("searchTwitter()")
		return nil
	}

	if len(search.Statuses) > 0 {
		return &search.Statuses[0]
	}

	return nil
}

func (bot *Scumbag) sendStatus(channel string, status *twitter.Tweet) {
	msg := fmt.Sprintf("@%s %s", status.User.ScreenName, status.Text)
	bot.Msg(channel, msg)
}
