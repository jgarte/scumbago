package scumbag

import (
	"encoding/json"
	"fmt"
)

const (
	GITHUB_USER_EVENTS_URL = "https://api.github.com/users/%s/events"
)

type GithubEvent struct {
	Payload GithubPayload `json:"payload"`
	Repo    GithubRepo    `json:"repo"`
	Type    string        `json:"type"`
}

type GithubRepo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type GithubPayload struct {
	Commits     []GithubEventCommit    `json:"commits"`
	Comment     GithubEventComment     `json:"comment"`
	PullRequest GithubEventPullRequest `json:"pull_request"`
}

type GithubEventCommit struct {
	Message string `json:"message"`
	Url     string `json:"url"`
}

type GithubEventComment struct {
	Body    string `json:"body"`
	HtmlUrl string `json:"html_url"`
}

type GithubEventPullRequest struct {
	HtmlUrl string `json:"html_url"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

type GithubCommit struct {
	HtmlUrl string            `json:"html_url"`
	Stats   GithubCommitStats `json:"stats"`
}

type GithubCommitStats struct {
	Total     int `json:"total"`
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

type GithubCommand struct {
	bot     *Scumbag
	channel string
}

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

	requestUrl := fmt.Sprintf(GITHUB_USER_EVENTS_URL, username)

	content, err := getContent(requestUrl)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("GithubCommand.Run()")
		return
	}

	events := make([]GithubEvent, 0)
	err = json.Unmarshal(content, &events)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("GithubCommand.Run()")
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

func (cmd *GithubCommand) pushEvent(event GithubEvent) {
	if len(event.Payload.Commits) > 0 {
		eventCommit := event.Payload.Commits[len(event.Payload.Commits)-1]

		content, err := getContent(eventCommit.Url)
		if err != nil {
			cmd.bot.Log.WithField("error", err).Error("GithubCommand.pushEvent()")
			return
		}

		var commit GithubCommit
		err = json.Unmarshal(content, &commit)
		if err != nil {
			cmd.bot.Log.WithField("error", err).Error("GithubCommand.pushEvent()")
			return
		}

		eventMsg := fmt.Sprintf("%s: %s", event.Repo.Name, eventCommit.Message)

		cmd.bot.Msg(cmd.channel, eventMsg)
		cmd.bot.Msg(cmd.channel, commit.HtmlUrl)
	}
}

func (cmd *GithubCommand) issueCommentEvent(event GithubEvent) {
	eventMsg := fmt.Sprintf("%s: %s", event.Repo.Name, event.Payload.Comment.Body)
	cmd.bot.Msg(cmd.channel, eventMsg)
	cmd.bot.Msg(cmd.channel, event.Payload.Comment.HtmlUrl)
}

func (cmd *GithubCommand) pullRequestEvent(event GithubEvent) {
	eventMsg := fmt.Sprintf("%s: PR: %s", event.Repo.Name, event.Payload.PullRequest.Title)
	cmd.bot.Msg(cmd.channel, eventMsg)
	cmd.bot.Msg(cmd.channel, event.Payload.PullRequest.HtmlUrl)
}
