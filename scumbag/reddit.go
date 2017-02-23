package scumbag

import (
	"errors"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/jzelinskie/geddit"
)

var (
	selfPostRegexp = regexp.MustCompile(`\Aself\.`)
)

func (bot *Scumbag) HandleRedditCommand(channel string, query string) {
	args := strings.Fields(query)

	if len(args) == 1 {
		randomSubredditSubmission(bot, channel, query)
	} else {
		if len(args) == 0 {
			return
		}

		switch args[0] {
		case "-t":
			subredditSubmission(bot, channel, args[1])
		case "-top":
			subredditSubmission(bot, channel, args[1])
		default:
			randomSubredditSubmission(bot, channel, query)
		}
	}
}

func getLatestSubmission(submissions []*geddit.Submission) (*geddit.Submission, error) {
	for _, submission := range submissions {
		if selfPostRegexp.Find([]byte(submission.Domain)) == nil {
			return submission, nil
		}
	}

	return nil, errors.New("No real submission found")
}

func subredditSubmission(bot *Scumbag, channel string, subreddit string) {
	bot.Log.WithField("subreddit", subreddit).Debug("subredditSubmission()")

	opts := geddit.ListingOptions{
		Limit: 3,
	}

	submissions, err := bot.Reddit.SubredditSubmissions(subreddit, geddit.HotSubmissions, opts)
	if err != nil {
		bot.Log.WithField("error", err).Error("subredditSubmission()")
		return
	}

	submission, err := getLatestSubmission(submissions)
	if err != nil {
		bot.Log.WithField("error", err).Error("subredditSubmission()")
		return
	}

	msg(bot, channel, submission)
}

func randomSubredditSubmission(bot *Scumbag, channel string, subreddit string) {
	bot.Log.WithField("subreddit", subreddit).Debug("randomSubredditSubmission()")

	opts := geddit.ListingOptions{
		Limit: 20,
	}

	submissions, err := bot.Reddit.SubredditSubmissions(subreddit, geddit.HotSubmissions, opts)
	if err != nil {
		bot.Log.WithField("error", err).Error("randomSubredditSubmission()")
		return
	}

	rand.Seed(time.Now().Unix())

	if len(submissions) > 0 {
		submission := submissions[rand.Intn(len(submissions))]
		msg(bot, channel, submission)
	}
}

func msg(bot *Scumbag, channel string, submission *geddit.Submission) {
	// This is needed because the URL returned has HTML escaped params for some dumbass reason.
	url := strings.Replace(submission.URL, "&amp;", "&", -1)
	bot.Msg(channel, url)
}
