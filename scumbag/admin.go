package scumbag

import (
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	CMD_NICK = "nick"
)

type AdminCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

func NewAdminCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *AdminCommand {
	return &AdminCommand{bot: bot, conn: conn, line: line}
}

func (cmd *AdminCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("AdminCommand.Run()")
		return
	}

	if !cmd.bot.Admin(cmd.line.Nick) {
		cmd.bot.Msg(cmd.conn, channel, "Fuck off.")
		return
	}

	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("AdminCommand.Run(): No args")
		return
	}

	fields := strings.Fields(args[0])

	if len(fields) > 1 {
		command := fields[0]
		commandArgs := strings.Join(fields[1:], " ")

		switch command {
		case CMD_NICK:
			client := cmd.bot.ircClients[cmd.conn.Config().Server]
			client.Nick(commandArgs)
		}
	} else {
		cmd.bot.Log.WithField("args", args).Error("AdminCommand.Run(): Could not get command args")
	}
}
