package scumbag

import (
	"errors"
	"math/rand"
	"regexp"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/jzelinskie/geddit"
)

const (
	redditHelp = cmdPrefix + "reddit <subreddit>"
)

var (
	selfPostRegexp = regexp.MustCompile(`\Aself\.`)
)

// RedditCommand interacts with the Reddit API.
type RedditCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewRedditCommand returns a new RedditCommand instance.
func NewRedditCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *RedditCommand {
	return &RedditCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *RedditCommand) Run(args ...string) {
	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("RedditCommand.Run(): No args")
		return
	}

	cmdArgs := strings.Fields(args[0])

	if len(cmdArgs) == 1 {
		cmd.randomSubredditSubmission(cmdArgs[0])
	} else {
		if len(cmdArgs) == 0 {
			return
		}

		switch cmdArgs[0] {
		case "-t":
			cmd.subredditSubmission(cmdArgs[1])
		case "-top":
			cmd.subredditSubmission(cmdArgs[1])
		default:
			cmd.randomSubredditSubmission(cmdArgs[0])
		}
	}
}

// Help shows the command help.
func (cmd *RedditCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("RedditCommand.Help()", err)
		return
	}

	cmd.bot.Msg(cmd.conn, channel, redditHelp)
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
		cmd.bot.LogError("RedditCommand.subredditSubmission()", err)
		return
	}

	submission, err := cmd.getLatestSubmission(submissions)
	if err != nil {
		cmd.bot.LogError("RedditCommand.subredditSubmission()", err)
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
		cmd.bot.LogError("RedditCommand.randomSubredditSubmission()", err)
		return
	}

	rand.Seed(time.Now().Unix())

	if len(submissions) > 0 {
		submission := submissions[rand.Intn(len(submissions))]
		cmd.msg(submission)
	}
}

func (cmd *RedditCommand) msg(submission *geddit.Submission) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("RedditCommand.msg()", err)
		return
	}

	// This is needed because the URL returned has HTML escaped params for some dumbass reason.
	url := strings.Replace(submission.URL, "&amp;", "&", -1)
	cmd.bot.Msg(cmd.conn, channel, url)
}
