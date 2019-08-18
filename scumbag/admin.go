package scumbag

import (
	"database/sql"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

const (
	cmdIgnore   = "ignore"
	cmdUnignore = "unignore"
	cmdNick     = "nick"
)

// AdminCommand handles bot admin.
type AdminCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewAdminCommand returns a new AdminCommand instance.
func NewAdminCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *AdminCommand {
	return &AdminCommand{bot: bot, conn: conn, line: line}
}

// Run runs the command.
func (cmd *AdminCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.LogError("AdminCommand.Run()", err)
		return
	}

	if !cmd.bot.Admin(cmd.line.Nick) {
		cmd.bot.Msg(cmd.conn, channel, "Fuck off.")
		return
	}

	if len(args) <= 0 {
		cmd.bot.Log.WithField("args", args).Debug("AdminCommand.Run(): No args")
		return
	}

	fields := strings.Fields(args[0])

	if len(fields) > 1 {
		command := fields[0]
		commandArgs := strings.Join(fields[1:], " ")

		switch command {
		case cmdIgnore:
			cmd.ignoreNick(cmd.conn.Config().Server, channel, commandArgs)
		case cmdUnignore:
			cmd.unignoreNick(cmd.conn.Config().Server, channel, commandArgs)
		case cmdNick:
			client := cmd.bot.ircClients[cmd.conn.Config().Server]
			client.Nick(commandArgs)
		}
	} else {
		cmd.bot.Log.WithField("args", args).Error("AdminCommand.Run(): Could not get command args")
	}
}

func (cmd *AdminCommand) ignoreNick(server, channel, nick string) {
	var nickMatch string
	err := cmd.bot.DB.QueryRow("SELECT nick FROM ignored_nicks WHERE server=$1 AND nick=$2;", server, nick).Scan(&nickMatch)
	if err == sql.ErrNoRows {
		// New nick ignore; create one.
		if _, insertErr := cmd.bot.DB.Exec("INSERT INTO ignored_nicks(server, nick, created_at) VALUES($1, $2, $3) RETURNING id;", server, nick, cmd.line.Time); insertErr != nil {
			cmd.bot.LogError("AdminCommand.ignoreNick()", insertErr)
		} else {
			cmd.bot.Msg(cmd.conn, channel, "Ignoring: "+nick)
		}
	}
}

func (cmd *AdminCommand) unignoreNick(server, channel, nick string) {
	_, err := cmd.bot.DB.Exec("DELETE FROM ignored_nicks WHERE server=$1 AND nick=$2;", server, nick)
	if err != nil {
		cmd.bot.LogError("AdminCommand.unignoreNick()", err)
	} else {
		cmd.bot.Msg(cmd.conn, channel, "Unignoring: "+nick)
	}
}
