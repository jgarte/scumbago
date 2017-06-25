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

// Used when just "help" is given.
var COMMANDS = []string{
	CMD_BEER,
	CMD_FIGLET,
	CMD_GITHUB,
	CMD_HELP,
	CMD_REDDIT,
	CMD_SPELL,
	CMD_TRUMP,
	CMD_TWITTER,
	CMD_URBAN_DICT,
	CMD_URL,
	CMD_WEATHER,
	CMD_WIKI,
	CMD_WOLFRAM,
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
	case strings.TrimLeft(CMD_BEER, CMD_PREFIX):
		NewBeerCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_FIGLET, CMD_PREFIX):
		NewFigletCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_GITHUB, CMD_PREFIX):
		NewGithubCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_REDDIT, CMD_PREFIX):
		NewRedditCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_SPELL, CMD_PREFIX):
		NewSpellcheckCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_TRUMP, CMD_PREFIX):
		NewTrumpCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_TWITTER, CMD_PREFIX):
		NewTwitterCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_URBAN_DICT, CMD_PREFIX):
		NewUrbanDictionaryCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_URL, CMD_PREFIX):
		NewLinkCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_WEATHER, CMD_PREFIX):
		NewWeatherCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_WIKI, CMD_PREFIX):
		NewWikiCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(CMD_WOLFRAM, CMD_PREFIX):
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

	// Strip CMD_PREFIX from the CMD_* constants.
	helpCommands := make([]string, len(COMMANDS))
	for i, command := range COMMANDS {
		helpCommands[i] = strings.TrimLeft(command, CMD_PREFIX)
	}

	helpText := "commands: " + strings.Join(helpCommands, ", ")
	cmd.bot.Msg(cmd.conn, channel, helpText)
}
