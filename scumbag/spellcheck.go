package scumbag

import (
	"os/exec"
	"regexp"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	ASPELL = "/usr/bin/aspell"
)

var (
	aspellRegexp = regexp.MustCompile(`\A&\s\w+\s\d+\s\d+:\s(.+)\z`)
	wordRegexp   = regexp.MustCompile(`(\w+)\s{1}\(sp\?\)`)
)

// Handler for "?sp <word>"
func (bot *Scumbag) HandleSpellCommand(channel string, args string) {
	response, err := bot.Spellcheck(args)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleSpellCommand()")
		return
	}

	bot.Log.WithField("response", response).Debug("HandleSpellCommand()")

	bot.Msg(channel, response)
}

// Searches for text like "some word (sp?) to spellcheck"
func (bot *Scumbag) SpellcheckLine(line *irc.Line) {
	channel := line.Args[0]

	if word, ok := bot.getWordFromLine(line); ok == true {
		response, err := bot.Spellcheck(word)
		if err != nil {
			bot.Log.WithField("error", err).Error("SpellcheckLine()")
			return
		}

		bot.Log.WithField("response", response).Debug("SpellcheckLine()")

		bot.Msg(channel, response)
	}
}

func (bot *Scumbag) Spellcheck(word string) (string, error) {
	echo := exec.Command("echo", word)
	aspell := exec.Command(ASPELL, "pipe")

	echoOut, err := echo.StdoutPipe()
	if err != nil {
		bot.Log.WithField("error", err).Error("Spellcheck()")
		return "", err
	}
	echo.Start()

	aspell.Stdin = echoOut
	output, err := aspell.Output()
	if err != nil {
		bot.Log.WithField("error", err).Error("Spellcheck()")
		return "", err
	}
	line := strings.Split(string(output[:]), "\n")[1]

	if strings.HasPrefix(line, "#") { // aspell's output starts with a '#' if no matches found.
		return "Beats me...", nil
	}

	spellMatch := aspellRegexp.FindStringSubmatch(line)
	if spellMatch != nil {
		return spellMatch[1], nil
	} else {
		return "GJ U CAN SPELL", nil
	}
}

func (bot *Scumbag) getWordFromLine(line *irc.Line) (string, bool) {
	msg := line.Args[1]
	match := wordRegexp.FindStringSubmatch(msg)
	if len(match) > 0 {
		return match[1], true
	} else {
		return "", false
	}
}
