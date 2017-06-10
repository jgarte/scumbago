package scumbag

import (
	"os/exec"
	"regexp"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	ASPELL          = "/usr/bin/aspell"
	ASPELL_REGEXP   = `\A&\s\w+\s\d+\s\d+:\s(.+)\z`
	SPELLCHECK_HELP = "?sp <word>"
)

var (
	aspellRegexp = regexp.MustCompile(ASPELL_REGEXP)
	wordRegexp   = regexp.MustCompile(CMD_ARG_REGEX)
)

type SpellcheckCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

func NewSpellcheckCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *SpellcheckCommand {
	return &SpellcheckCommand{bot: bot, conn: conn, line: line}
}

// Handler for "?sp <word>"
func (cmd *SpellcheckCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("SpellcheckCommand.Run()")
		return
	}

	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("SpellcheckCommand.Run(): No args")
		return
	}

	word := args[0]
	if word == "" {
		cmd.bot.Log.Debug("SpellcheckCommand.Run(): No word")
		return
	}

	response, err := cmd.Spellcheck(word)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("SpellcheckCommand.Run()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, response)
}

func (cmd *SpellcheckCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("SpellcheckCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, SPELLCHECK_HELP)
}

// Called from a goroutine to search for text like "some word (sp?) to spellcheck"
func (bot *Scumbag) SpellcheckLine(conn *irc.Conn, line *irc.Line) {
	if len(line.Args) <= 0 {
		return
	}

	cmd := NewSpellcheckCommand(bot, conn, line)

	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("AdminCommand.Run()")
		return
	}

	if word, ok := cmd.getWordFromLine(line); ok == true {
		response, err := cmd.Spellcheck(word)
		if err != nil {
			cmd.bot.Log.WithField("error", err).Error("Scumbag.SpellcheckLine()")
			return
		}

		cmd.bot.Msg(cmd.conn, channel, response)
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
