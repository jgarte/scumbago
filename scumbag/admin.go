package scumbag

import (
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	CMD_NICK = "nick"
)

type AdminCommand struct {
	bot     *Scumbag
	channel string
	conn    *irc.Conn
	line    *irc.Line
}

func (cmd *AdminCommand) Run(args ...string) {
	if !cmd.bot.Admin(cmd.line.Nick) {
		cmd.bot.Msg(cmd.conn, cmd.channel, "Fuck off.")
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
			cmd.bot.ircClient.Nick(commandArgs)
		}
	} else {
		cmd.bot.Log.WithField("args", args).Error("AdminCommand.Run(): Could not get command args")
	}
}
