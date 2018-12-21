package scumbag

import (
	"time"

	"github.com/davidscholberg/go-durationfmt"
	irc "github.com/fluffle/goirc/client"
)

// UptimeCommand displays bot uptime.
type UptimeCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewUptimeCommand returns a new UptimeCommand instance.
func NewUptimeCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *UptimeCommand {
	return &UptimeCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *UptimeCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("UptimeCommand.Run()")
		return
	}

	uptime, err := durationfmt.Format(time.Since(cmd.bot.startTime), "%d days, %0h:%0m:%0s")
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("UptimeCommand.Run()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, uptime)
}

// Help shows the command help.
func (cmd *UptimeCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("UptimeCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, "Show bot uptime")
}
