package scumbag

import (
	"strings"

	"github.com/jzelinskie/geddit"
)

func (bot *Scumbag) HandleRedditCommand(channel string, subreddit string) {
	opts := geddit.ListingOptions{
		Limit: 1,
	}

	submissions, err := bot.Reddit.SubredditSubmissions(subreddit, geddit.HotSubmissions, opts)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleRedditCommand()")
		return
	}

	for _, submission := range submissions {
		// This is needed because the URL returned has HTML escaped params for some dumbass reason.
		url := strings.Replace(submission.URL, "&amp;", "&", -1)
		bot.Msg(channel, url)
	}
}
