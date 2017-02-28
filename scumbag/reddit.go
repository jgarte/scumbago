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

type RedditCommand struct {
	bot     *Scumbag
	channel string
	query   string
}

func (cmd *RedditCommand) Run() {
	args := strings.Fields(cmd.query)

	if len(args) == 1 {
		cmd.randomSubredditSubmission(cmd.query)
	} else {
		if len(args) == 0 {
			return
		}

		switch args[0] {
		case "-t":
			cmd.subredditSubmission(args[1])
		case "-top":
			cmd.subredditSubmission(args[1])
		default:
			cmd.randomSubredditSubmission(cmd.query)
		}
	}
}

func (cmd *RedditCommand) getLatestSubmission(submissions []*geddit.Submission) (*geddit.Submission, error) {
	for _, submission := range submissions {
		if selfPostRegexp.Find([]byte(submission.Domain)) == nil {
			return submission, nil
		}
	}

	return nil, errors.New("No real submission found")
}

func (cmd *RedditCommand) subredditSubmission(subreddit string) {
	cmd.bot.Log.WithField("subreddit", subreddit).Debug("RedditCommand.subredditSubmission()")

	opts := geddit.ListingOptions{
		Limit: 3,
	}

	submissions, err := cmd.bot.Reddit.SubredditSubmissions(subreddit, geddit.HotSubmissions, opts)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("RedditCommand.subredditSubmission()")
		return
	}

	submission, err := cmd.getLatestSubmission(submissions)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("RedditCommand.subredditSubmission()")
		return
	}

	cmd.msg(submission)
}

func (cmd *RedditCommand) randomSubredditSubmission(subreddit string) {
	cmd.bot.Log.WithField("subreddit", subreddit).Debug("RedditCommand.randomSubredditSubmission()")

	opts := geddit.ListingOptions{
		Limit: 20,
	}

	submissions, err := cmd.bot.Reddit.SubredditSubmissions(subreddit, geddit.HotSubmissions, opts)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("RedditCommand.randomSubredditSubmission()")
		return
	}

	rand.Seed(time.Now().Unix())

	if len(submissions) > 0 {
		submission := submissions[rand.Intn(len(submissions))]
		cmd.msg(submission)
	}
}

func (cmd *RedditCommand) msg(submission *geddit.Submission) {
	// This is needed because the URL returned has HTML escaped params for some dumbass reason.
	url := strings.Replace(submission.URL, "&amp;", "&", -1)
	cmd.bot.Msg(cmd.channel, url)
}
