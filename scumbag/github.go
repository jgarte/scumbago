package scumbag

import (
	"encoding/json"
	"fmt"
)

const (
	GITHUB_USER_EVENTS_URL = "https://api.github.com/users/%s/events"
)

type GithubEvent struct {
	Repo    GithubRepo    `json:"repo"`
	Payload GithubPayload `json:"payload"`
}

type GithubRepo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type GithubPayload struct {
	Commits []GithubEventCommit `json:"commits"`
}

type GithubEventCommit struct {
	Message string `json:"message"`
	Url     string `json:"url"`
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

		if len(event.Payload.Commits) > 0 {
			eventCommit := event.Payload.Commits[0]

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
}
