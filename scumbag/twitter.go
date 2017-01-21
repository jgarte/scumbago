package scumbag

import (
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
)

func (bot *Scumbag) HandleTwitterCommand(channel string, query string) {
	bot.Log.WithField("query", query).Debug("HandleTwitterCommand()")

	search, _, err := bot.Twitter.Search.Tweets(&twitter.SearchTweetParams{
		Query: query,
	})
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleTwitterCommand()")
	}

	if len(search.Statuses) > 0 {
		status := search.Statuses[0]

		msg := fmt.Sprintf("@%s %s", status.User.ScreenName, status.Text)
		bot.Msg(channel, msg)
	}
}
