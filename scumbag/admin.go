package scumbag

import (
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	CMD_NICK = "nick"
)

func (bot *Scumbag) HandleAdminCommand(channel string, command_and_args string, line *irc.Line) {
	if !bot.Admin(line.Nick) {
		msg := line.Nick + ": Fuck off."
		bot.Msg(channel, msg)
		return
	}

	fields := strings.Fields(command_and_args)

	if len(fields) > 1 {
		command := fields[0]
		args := strings.Join(fields[1:], " ")

		switch command {
		case CMD_NICK:
			bot.ircClient.Nick(args)
		}
	} else {
		bot.Log.WithField("command_and_args", command_and_args).Error("HandleAdminCommand(): Could not get command args")
	}
}
