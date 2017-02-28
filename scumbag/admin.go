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
	args    string
	line    *irc.Line
}

func (cmd *AdminCommand) Run() {
	if !cmd.bot.Admin(cmd.line.Nick) {
		cmd.bot.Msg(cmd.channel, "Fuck off.")
		return
	}

	fields := strings.Fields(cmd.args)

	if len(fields) > 1 {
		command := fields[0]
		commandArgs := strings.Join(fields[1:], " ")

		switch command {
		case CMD_NICK:
			cmd.bot.ircClient.Nick(commandArgs)
		}
	} else {
		cmd.bot.Log.WithField("args", cmd.args).Error("AdminCommand.Run(): Could not get command args")
	}
}
