package scumbag

import (
	"fmt"
	"net/url"

	irc "github.com/fluffle/goirc/client"
)

const (
	wolframAPIURL = "http://api.wolframalpha.com/v1/result?appid=%s&i=%s"

	wolframHelp = cmdPrefix + "wolfram <query>"
)

// WolframAlphaCommand interacts with the Wolfram Alpha API.
type WolframAlphaCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewWolframAlphaCommand returns a new WolframAlphaCommand instance.
func NewWolframAlphaCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *WolframAlphaCommand {
	return &WolframAlphaCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *WolframAlphaCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("WolframAlphaCommand.Run()", err)
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("WolframAlphaCommand.Run(): No query")
		cmd.Help()
		return
	}
	query = url.QueryEscape(query)

	requestURL := fmt.Sprintf(wolframAPIURL, cmd.bot.Config.WolframAlpha.AppID, query)

	content, err := getContent(requestURL)
	if err != nil {
		cmd.bot.LogError("WolframAlphaCommand.Run()", err)
		return
	}

	cmd.bot.Msg(cmd.conn, channel, string(content[:]))
}

// Help displays the command help.
func (cmd *WolframAlphaCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("WolframAlphaCommand.Help()", err)
		return
	}

	cmd.bot.Msg(cmd.conn, channel, wolframHelp)
}
