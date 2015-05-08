package scumbag

import (
	"strings"
)

const (
	URL_SEP = " | "
)

func HandleUrlCommand(bot *Scumbag, channel string, args string) {
	links, err := SearchLinks(bot, args)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleUrlCommand()")
		return
	}

	response := make([]string, len(links))
	for i, link := range links {
		response[i] = link.Url
	}

	bot.ircClient.Privmsg(channel, strings.Join(response, URL_SEP))
}
