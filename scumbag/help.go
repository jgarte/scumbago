package scumbag

import (
	"strings"

	irc "github.com/fluffle/goirc/client"
)

type HelpCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// These are basically the same as the CMD_* constants, only without the leading "?".
// TODO: Just use the CMD_* constants and strip out the leading "?".
const (
	HELP_BEER       = "beer"
	HELP_FIGLET     = "fig"
	HELP_GITHUB     = "gh"
	HELP_HELP       = "help"
	HELP_REDDIT     = "reddit"
	HELP_SPELL      = "sp"
	HELP_TRUMP      = "trump"
	HELP_TWITTER    = "twitter"
	HELP_URBAN_DICT = "ud"
	HELP_URL        = "url"
	HELP_WEATHER    = "weather"
	HELP_WIKI       = "wp"
	HELP_WOLFRAM    = "wolfram"
)

// Used when just "?help" is given.
var COMMANDS = []string{
	HELP_BEER,
	HELP_FIGLET,
	HELP_GITHUB,
	HELP_HELP,
	HELP_REDDIT,
	HELP_SPELL,
	HELP_TRUMP,
	HELP_TWITTER,
	HELP_URBAN_DICT,
	HELP_URL,
	HELP_WEATHER,
	HELP_WIKI,
	HELP_WOLFRAM,
}

func NewHelpCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *HelpCommand {
	return &HelpCommand{bot: bot, conn: conn, line: line}
}

func (cmd *HelpCommand) Run(args ...string) {
	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("HelpCommand.Run(): No args")
		return
	}

	helpPhrase := args[0]
	switch helpPhrase {
	case HELP_BEER:
		NewBeerCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_FIGLET:
		NewFigletCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_GITHUB:
		NewGithubCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_REDDIT:
		NewRedditCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_SPELL:
		NewSpellcheckCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_TRUMP:
		NewTrumpCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_TWITTER:
		NewTwitterCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_URBAN_DICT:
		NewUrbanDictionaryCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_URL:
		NewLinkCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_WEATHER:
		NewWeatherCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_WIKI:
		NewWikiCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case HELP_WOLFRAM:
		NewWolframAlphaCommand(cmd.bot, cmd.conn, cmd.line).Help()
	default:
		NewHelpCommand(cmd.bot, cmd.conn, cmd.line).Help()
	}
}

func (cmd *HelpCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("HelpCommand.Help()")
		return
	}

	helpText := "commands: " + strings.Join(COMMANDS, ", ")
	cmd.bot.Msg(cmd.conn, channel, helpText)
}
