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
	phrase  string
}

func (cmd *FigletCommand) Run() {
	if output, err := exec.Command(FIGLET, cmd.phrase).Output(); err != nil {
		cmd.bot.Log.WithField("error", err).Error("FigletCommand.Run()")
	} else {
		for _, line := range strings.Split(string(output), "\n") {
			cmd.bot.Msg(cmd.channel, line)
		}
	}
}
