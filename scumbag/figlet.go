package scumbag

import (
	"os/exec"
	"strings"
)

const (
	FIGLET = "/usr/bin/figlet"
)

func (bot *Scumbag) HandleFiglet(channel string, phrase string) {
	if output, err := exec.Command(FIGLET, phrase).Output(); err != nil {
		bot.Log.WithField("error", err).Error("HandleFiglet()")
		return
	} else {
		for _, line := range strings.Split(string(output), "\n") {
			bot.Msg(channel, line)
		}
	}
}
