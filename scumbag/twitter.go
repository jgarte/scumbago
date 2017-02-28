package scumbag

import (
	"fmt"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
)

type TwitterCommand struct {
	bot     *Scumbag
	channel string
	query   string
}

func (cmd *TwitterCommand) Run() {
	cmd.bot.Log.WithField("query", cmd.query).Debug("TwitterCommand.Run()")

	switch {
	case strings.HasPrefix(cmd.query, "@"):
		user, protected := cmd.screennameStatus()

		var msg string
		if protected {
			msg = fmt.Sprintf("Account is not public: %s", cmd.query)
		} else {
			if user.ScreenName != "" && user.Status != nil {
				msg = fmt.Sprintf("@%s %s", user.ScreenName, user.Status.Text)
			} else {
				msg = fmt.Sprintf("Account has no tweets: %s", cmd.query)
			}
		}
		cmd.bot.Msg(cmd.channel, msg)
	default:
		status := cmd.searchTwitter()
		if status != nil {
			msg := fmt.Sprintf("@%s %s", status.User.ScreenName, status.Text)
			cmd.bot.Msg(cmd.channel, msg)
		}
	}
}

// func (bot *Scumbag) HandleTwitterCommand(channel string, query string) {
//   bot.Log.WithField("query", query).Debug("HandleTwitterCommand()")

//   switch {
//   case strings.HasPrefix(query, "@"):
//     user, protected := bot.screennameStatus(query)

//     var msg string
//     if protected {
//       msg = fmt.Sprintf("Account is not public: %s", query)
//     } else {
//       if user.ScreenName != "" && user.Status != nil {
//         msg = fmt.Sprintf("@%s %s", user.ScreenName, user.Status.Text)
//       } else {
//         msg = fmt.Sprintf("Account has no tweets: %s", query)
//       }
//     }
//     bot.Msg(channel, msg)
//   default:
//     status := bot.searchTwitter(query)
//     if status != nil {
//       msg := fmt.Sprintf("@%s %s", status.User.ScreenName, status.Text)
//       bot.Msg(channel, msg)
//     }
//   }
// }

func (cmd *TwitterCommand) screennameStatus() (*twitter.User, bool) {
	user, _, err := cmd.bot.Twitter.Users.Show(&twitter.UserShowParams{
		ScreenName:      strings.Replace(cmd.query, "@", "", 1),
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
	} else {
		return user, false
	}
}

func (cmd *TwitterCommand) searchTwitter() *twitter.Tweet {
	search, _, err := cmd.bot.Twitter.Search.Tweets(&twitter.SearchTweetParams{
		Query: cmd.query,
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
