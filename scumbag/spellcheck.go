package scumbag

import (
	// "os/exec"
	"regexp"

	irc "github.com/fluffle/goirc/client"
)

const (
	ASPELL = "/usr/bin/aspell"
)

var (
	wordRegexp = regexp.MustCompile(`(\w+)\s{1}\(sp\?\)`)
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
	bot.Log.WithField("word", word).Debug("Spellcheck()")
	return "spellcheck: " + word, nil
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
