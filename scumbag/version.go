package scumbag

import (
	irc "github.com/fluffle/goirc/client"
)

// VersionCommand displays the bot version.
type VersionCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewVersionCommand returns a new VersionCommand instance.
func NewVersionCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *VersionCommand {
	return &VersionCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *VersionCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("VersionCommand.Run()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, VersionString())
}
