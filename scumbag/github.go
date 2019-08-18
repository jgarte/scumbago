package scumbag

import (
	"encoding/json"
	"fmt"

	irc "github.com/fluffle/goirc/client"
)

const (
	githubUserEventsURL = "https://api.github.com/users/%s/events"
	githubHelp          = cmdPrefix + "gh <username>"
)

// GithubEvent stores a Github API response.
type GithubEvent struct {
	Payload GithubPayload `json:"payload"`
	Repo    GithubRepo    `json:"repo"`
	Type    string        `json:"type"`
}

// GithubRepo stores repo information.
type GithubRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// GithubPayload stores various information from the API.
type GithubPayload struct {
	Commits     []GithubEventCommit    `json:"commits"`
	Comment     GithubEventComment     `json:"comment"`
	PullRequest GithubEventPullRequest `json:"pull_request"`
}

// GithubEventCommit stores information on a commit event.
type GithubEventCommit struct {
	Message string `json:"message"`
	URL     string `json:"url"`
}

// GithubEventComment stores information on a comment event.
type GithubEventComment struct {
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

// GithubEventPullRequest stores information on a pull request event.
type GithubEventPullRequest struct {
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

// GithubCommit stores information on a commit.
type GithubCommit struct {
	HTMLURL string            `json:"html_url"`
	Stats   GithubCommitStats `json:"stats"`
}

// GithubCommitStats stores statistics on a single commit.
type GithubCommitStats struct {
	Total     int `json:"total"`
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

// GithubCommand interacts with the Github API.
type GithubCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewGithubCommand returns a new GithubCommand instance.
func NewGithubCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *GithubCommand {
	return &GithubCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *GithubCommand) Run(args ...string) {
	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("GithubCommand.Run(): No args")
		return
	}

	username := args[0]
	if username == "" {
		cmd.bot.Log.Debug("GithubCommand.Run(): No username")
		return
	}

	requestURL := fmt.Sprintf(githubUserEventsURL, username)

	content, err := getContent(requestURL)
	if err != nil {
		cmd.bot.LogError("GithubCommand.Run()", err)
		return
	}

	events := make([]GithubEvent, 0)
	err = json.Unmarshal(content, &events)
	if err != nil {
		cmd.bot.LogError("GithubCommand.Run()", err)
		return
	}

	if len(events) > 0 {
		event := events[0]

		switch event.Type {
		case "PushEvent":
			cmd.pushEvent(event)
		case "IssueCommentEvent":
			cmd.issueCommentEvent(event)
		case "PullRequestEvent":
			cmd.pullRequestEvent(event)
		default:
			cmd.bot.Log.WithField("event", event).Warn("GithubCommand.Run(): Unhandled event")
		}
	}
}

// Help shows the command help.
func (cmd *GithubCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("GithubCommand.Help()", err)
		return
	}

	cmd.bot.Msg(cmd.conn, channel, githubHelp)
}

func (cmd *GithubCommand) pushEvent(event GithubEvent) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("GithubCommand.pushEvent()", err)
		return
	}

	if len(event.Payload.Commits) > 0 {
		eventCommit := event.Payload.Commits[len(event.Payload.Commits)-1]

		content, err := getContent(eventCommit.URL)
		if err != nil {
			cmd.bot.LogError("GithubCommand.pushEvent()", err)
			return
		}

		var commit GithubCommit
		err = json.Unmarshal(content, &commit)
		if err != nil {
			cmd.bot.LogError("GithubCommand.pushEvent()", err)
			return
		}

		eventMsg := fmt.Sprintf("%s: %s", event.Repo.Name, eventCommit.Message)

		cmd.bot.Msg(cmd.conn, channel, eventMsg)
		cmd.bot.Msg(cmd.conn, channel, commit.HTMLURL)
	}
}

func (cmd *GithubCommand) issueCommentEvent(event GithubEvent) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("GithubCommand.issueCommentEvent()", err)
		return
	}

	eventMsg := fmt.Sprintf("%s: %s", event.Repo.Name, event.Payload.Comment.Body)
	cmd.bot.Msg(cmd.conn, channel, eventMsg)
	cmd.bot.Msg(cmd.conn, channel, event.Payload.Comment.HTMLURL)
}

func (cmd *GithubCommand) pullRequestEvent(event GithubEvent) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("GithubCommand.pullRequestEvent()", err)
		return
	}

	eventMsg := fmt.Sprintf("%s: PR: %s", event.Repo.Name, event.Payload.PullRequest.Title)
	cmd.bot.Msg(cmd.conn, channel, eventMsg)
	cmd.bot.Msg(cmd.conn, channel, event.Payload.PullRequest.HTMLURL)
}
