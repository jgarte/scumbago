package scumbag

import (
	"strings"

	irc "github.com/fluffle/goirc/client"
)

// HelpCommand shows help on all the commands.
type HelpCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

var helpCommands = []string{
	cmdFiglet,
	cmdCorona,
	cmdGame,
	cmdGithub,
	cmdHackerNews,
	cmdHelp,
	cmdMovie,
	cmdNews,
	cmdReddit,
	cmdSpell,
	cmdTwitter,
	cmdURL,
	cmdUptime,
	cmdUrbanDict,
	cmdWeather,
	cmdWiki,
	cmdWolfram,
}

// NewHelpCommand returns a new HelpCommand instance.
func NewHelpCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *HelpCommand {
	return &HelpCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *HelpCommand) Run(args ...string) {
	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("HelpCommand.Run(): No args")
		return
	}

	helpPhrase := args[0]
	switch helpPhrase {
	case strings.TrimLeft(cmdCorona, cmdPrefix):
		NewCoronaCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdFiglet, cmdPrefix):
		NewFigletCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdGame, cmdPrefix):
		NewGameCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdGithub, cmdPrefix):
		NewGithubCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdHackerNews, cmdPrefix):
		NewHackerNewsCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdNews, cmdPrefix):
		NewNewsCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdReddit, cmdPrefix):
		NewRedditCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdSpell, cmdPrefix):
		NewSpellcheckCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdTwitter, cmdPrefix):
		NewTwitterCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdUrbanDict, cmdPrefix):
		NewUrbanDictionaryCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdURL, cmdPrefix):
		NewLinkCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdUptime, cmdPrefix):
		NewUptimeCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdWeather, cmdPrefix):
		NewWeatherCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdWiki, cmdPrefix):
		NewWikiCommand(cmd.bot, cmd.conn, cmd.line).Help()
	case strings.TrimLeft(cmdWolfram, cmdPrefix):
		NewWolframAlphaCommand(cmd.bot, cmd.conn, cmd.line).Help()
	default:
		NewHelpCommand(cmd.bot, cmd.conn, cmd.line).Help()
	}
}

// Help shows the command help.
func (cmd *HelpCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("HelpCommand.Help()", err)
		return
	}

	// Strip cmdPrefix from the cmd* constants.
	help := make([]string, len(helpCommands))
	for i, command := range helpCommands {
		help[i] = strings.TrimLeft(command, cmdPrefix)
	}

	helpText := "commands: " + strings.Join(help, ", ")
	cmd.bot.Msg(cmd.conn, channel, helpText)
}
