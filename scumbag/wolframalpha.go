package scumbag

import (
	"fmt"
	"net/url"

	irc "github.com/fluffle/goirc/client"
)

const (
	WOLFRAM_ALPHA_URL = "http://api.wolframalpha.com/v1/result?appid=%s&i=%s"

	WOLFRAM_HELP = CMD_PREFIX + "wolfram <query>"
)

type WolframAlphaCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

func NewWolframAlphaCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *WolframAlphaCommand {
	return &WolframAlphaCommand{bot: bot, conn: conn, line: line}
}

func (cmd *WolframAlphaCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WeatherCommand.currentConditions()")
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("WolframAlphaCommand.Run(): No query")
		cmd.Help()
		return
	}
	query = url.QueryEscape(query)

	requestUrl := fmt.Sprintf(WOLFRAM_ALPHA_URL, cmd.bot.Config.WolframAlpha.AppID, query)

	content, err := getContent(requestUrl)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("WolframAlphaCommand.Run()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, string(content[:]))
}

func (cmd *WolframAlphaCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("WolframAlphaCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, WOLFRAM_HELP)
}
