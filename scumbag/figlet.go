package scumbag

import (
	"os/exec"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	FIGLET      = "/usr/bin/figlet"
	FIGLET_HELP = CMD_PREFIX + "fig <phrase>"
)

type FigletCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

func NewFigletCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *FigletCommand {
	return &FigletCommand{bot: bot, conn: conn, line: line}
}

func (cmd *FigletCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("FigletCommand.Run()")
		return
	}

	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("FigletCommand.Run(): No args")
		return
	}

	phrase := args[0]
	if phrase == "" {
		cmd.bot.Log.Debug("FigletCommand.Run(): No phrase")
		return
	}

	if output, err := exec.Command(FIGLET, phrase).Output(); err != nil {
		cmd.bot.Log.WithField("error", err).Error("FigletCommand.Run()")
	} else {
		for _, line := range strings.Split(string(output), "\n") {
			cmd.bot.Msg(cmd.conn, channel, line)
		}
	}
}

func (cmd *FigletCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("FigletCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, FIGLET_HELP)
}
