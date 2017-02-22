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
		user, protected := bot.screennameStatus(query)

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
		bot.Msg(channel, msg)
	default:
		status := bot.searchTwitter(query)
		if status != nil {
			msg := fmt.Sprintf("@%s %s", status.User.ScreenName, status.Text)
			bot.Msg(channel, msg)
		}
	}
}

func (bot *Scumbag) screennameStatus(screenname string) (*twitter.User, bool) {
	user, _, err := bot.Twitter.Users.Show(&twitter.UserShowParams{
		ScreenName:      strings.Replace(screenname, "@", "", 1),
		IncludeEntities: twitter.Bool(true),
	})
	if err != nil {
		bot.Log.WithField("error", err).Error("screennameStatus()")
		return nil, true
	}

	// Only return the User if their timeline is public.
	if user.Protected {
		bot.Log.WithField("user", user).Debug("screennameStatus(): User account is protected.")
		return user, true
	} else {
		return user, false
	}
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
