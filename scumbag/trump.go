package scumbag

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"

	irc "github.com/fluffle/goirc/client"
	"github.com/kennygrant/sanitize"
	"golang.org/x/tools/blog/atom"
)

const (
	WTF_TRUMP_URL = "https://whatthefuckjusthappenedtoday.com/atom.xml"
)

type TrumpCommand struct {
	bot     *Scumbag
	channel string
	conn    *irc.Conn
}

func (cmd *TrumpCommand) Run(args ...string) {
	response, err := http.Get(WTF_TRUMP_URL)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("TrumpCommand.Run()")
		return
	}

	content, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("TrumpCommand.Run()")
		return
	}

	var feed atom.Feed
	err = xml.Unmarshal(content, &feed)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("TrumpCommand.Run()")
		return
	}

	if len(feed.Entry) > 0 {
		entry := feed.Entry[0]
		content := sanitize.HTML(entry.Content.Body[0:300])

		msg := strings.Join([]string{content, "..."}, "")
		cmd.bot.Msg(cmd.conn, cmd.channel, msg)

		if len(entry.Link) > 0 {
			cmd.bot.Msg(cmd.conn, cmd.channel, entry.Link[0].Href)
		}
	}
}
