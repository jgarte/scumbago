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

func (bot *Scumbag) HandleGithubCommand(channel string, username string) {
	requestUrl := fmt.Sprintf(GITHUB_USER_EVENTS_URL, username)

	content, err := getContent(requestUrl)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleGithubCommand()")
		return
	}

	events := make([]GithubEvent, 0)
	err = json.Unmarshal(content, &events)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleGithubCommand()")
		return
	}

	if len(events) > 0 {
		event := events[0]

		switch event.Type {
		case "PushEvent":
			pushEvent(bot, channel, event)
		case "IssueCommentEvent":
			issueCommentEvent(bot, channel, event)
		case "PullRequestEvent":
			pullRequestEvent(bot, channel, event)
		default:
			bot.Log.WithField("event", event).Warn("HandleGithubCommand(): Unhandled event")
		}
	}
}

func pushEvent(bot *Scumbag, channel string, event GithubEvent) {
	if len(event.Payload.Commits) > 0 {
		eventCommit := event.Payload.Commits[len(event.Payload.Commits)-1]

		content, err := getContent(eventCommit.Url)
		if err != nil {
			bot.Log.WithField("error", err).Error("HandleGithubCommand()")
			return
		}

		var commit GithubCommit
		err = json.Unmarshal(content, &commit)
		if err != nil {
			bot.Log.WithField("error", err).Error("HandleGithubCommand()")
			return
		}

		eventMsg := fmt.Sprintf("%s: %s", event.Repo.Name, eventCommit.Message)

		bot.Msg(channel, eventMsg)
		bot.Msg(channel, commit.HtmlUrl)
	}
}

func issueCommentEvent(bot *Scumbag, channel string, event GithubEvent) {
	eventMsg := fmt.Sprintf("%s: %s", event.Repo.Name, event.Payload.Comment.Body)
	bot.Msg(channel, eventMsg)
	bot.Msg(channel, event.Payload.Comment.HtmlUrl)
}

func pullRequestEvent(bot *Scumbag, channel string, event GithubEvent) {
	eventMsg := fmt.Sprintf("%s: PR: %s", event.Repo.Name, event.Payload.PullRequest.Title)
	bot.Msg(channel, eventMsg)
	bot.Msg(channel, event.Payload.PullRequest.HtmlUrl)
}
