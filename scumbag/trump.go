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
	trumpHelp   = cmdPrefix + "trump - Returns the latest WTF Trump news."
	wtfTrumpURL = "https://whatthefuckjusthappenedtoday.com/atom.xml"
)

// TrumpCommand interacts with the (boring and I'm fucking tired of hearing about it) WTFTrump API.
type TrumpCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewTrumpCommand returns a new TrumpCommand instance.
func NewTrumpCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *TrumpCommand {
	return &TrumpCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *TrumpCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("TrumpCommand.Run()")
		return
	}

	response, err := http.Get(wtfTrumpURL)
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
		cmd.bot.Msg(cmd.conn, channel, msg)

		if len(entry.Link) > 0 {
			cmd.bot.Msg(cmd.conn, channel, entry.Link[0].Href)
		}
	}
}

// Help displays the command help.
func (cmd *TrumpCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("TrumpCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, trumpHelp)
}
