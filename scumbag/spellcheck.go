package scumbag

import (
	"os/exec"
	"regexp"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	ASPELL        = "/usr/bin/aspell"
	ASPELL_REGEXP = `\A&\s\w+\s\d+\s\d+:\s(.+)\z`
)

var (
	aspellRegexp = regexp.MustCompile(ASPELL_REGEXP)
	wordRegexp   = regexp.MustCompile(CMD_ARG_REGEX)
)

type SpellcheckCommand struct {
	bot     *Scumbag
	channel string
	word    string
}

// Handler for "?sp <word>"
func (cmd *SpellcheckCommand) Run() {
	response, err := cmd.Spellcheck(cmd.word)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("SpellcheckCommand.Run()")
		return
	}

	cmd.bot.Msg(cmd.channel, response)
}

// Called from a goroutine to search for text like "some word (sp?) to spellcheck"
func (bot *Scumbag) SpellcheckLine(line *irc.Line) {
	if len(line.Args) <= 0 {
		return
	}

	channel := line.Args[0]

	cmd := &SpellcheckCommand{bot: bot, channel: channel}
	if word, ok := cmd.getWordFromLine(line); ok == true {
		response, err := cmd.Spellcheck(word)
		if err != nil {
			cmd.bot.Log.WithField("error", err).Error("Scumbag.SpellcheckLine()")
			return
		}

		cmd.bot.Msg(cmd.channel, response)
	}
}

func (cmd *SpellcheckCommand) Spellcheck(word string) (string, error) {
	echo := exec.Command("echo", word)
	aspell := exec.Command(ASPELL, "pipe")

	echoOut, err := echo.StdoutPipe()
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("SpellcheckCommand.Spellcheck()")
		return "", err
	}
	echo.Start()

	aspell.Stdin = echoOut
	output, err := aspell.Output()
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("SpellcheckCommand.Spellcheck()")
		return "", err
	}
	line := strings.Split(string(output[:]), "\n")[1]

	if strings.HasPrefix(line, "#") { // aspell's output starts with a '#' if no matches found.
		return "Beats me...", nil
	}

	spellMatch := aspellRegexp.FindStringSubmatch(line)
	if len(spellMatch) > 0 {
		return spellMatch[1], nil
	} else {
		return "GJ U CAN SPELL", nil
	}
}

func (cmd *SpellcheckCommand) getWordFromLine(line *irc.Line) (string, bool) {
	msg := line.Args[1]
	match := wordRegexp.FindStringSubmatch(msg)
	if len(match) > 0 {
		return match[1], true
	} else {
		return "", false
	}
}
