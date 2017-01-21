package scumbag

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jzelinskie/geddit"
)

var (
	selfPostRegexp = regexp.MustCompile(`\Aself\.`)
)

func (bot *Scumbag) HandleRedditCommand(channel string, subreddit string) {
	opts := geddit.ListingOptions{
		Limit: 5,
	}

	submissions, err := bot.Reddit.SubredditSubmissions(subreddit, geddit.HotSubmissions, opts)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleRedditCommand()")
		return
	}

	submission, err := getLatestSubmission(submissions)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleRedditCommand()")
		return
	}

	// This is needed because the URL returned has HTML escaped params for some dumbass reason.
	url := strings.Replace(submission.URL, "&amp;", "&", -1)
	bot.Msg(channel, url)
}

func getLatestSubmission(submissions []*geddit.Submission) (*geddit.Submission, error) {
	for _, submission := range submissions {
		if selfPostRegexp.Find([]byte(submission.Domain)) == nil {
			return submission, nil
		}
	}

	return nil, errors.New("No real submission found")
}
