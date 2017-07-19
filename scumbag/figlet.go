package scumbag

import (
	"os/exec"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	figletPath = "/usr/bin/figlet"
	figletHelp = cmdPrefix + "fig <phrase>"
)

// FigletCommand interacts with the figlet command.
type FigletCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewFigletCommand returns a new FigletCommand instance.
func NewFigletCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *FigletCommand {
	return &FigletCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
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

	if output, err := exec.Command(figletPath, phrase).Output(); err != nil {
		cmd.bot.Log.WithField("error", err).Error("FigletCommand.Run()")
	} else {
		for _, line := range strings.Split(string(output), "\n") {
			cmd.bot.Msg(cmd.conn, channel, line)
		}
	}
}

// Help shows the command help.
func (cmd *FigletCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("FigletCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, figletHelp)
}
