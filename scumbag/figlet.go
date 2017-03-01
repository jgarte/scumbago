package scumbag

import (
	"os/exec"
	"strings"
)

const (
	FIGLET = "/usr/bin/figlet"
)

type FigletCommand struct {
	bot     *Scumbag
	channel string
}

func (cmd *FigletCommand) Run(args ...string) {
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
			cmd.bot.Msg(cmd.channel, line)
		}
	}
}
