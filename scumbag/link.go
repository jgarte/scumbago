package scumbag

import (
	"database/sql"
	"regexp"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	log "github.com/sirupsen/logrus"
)

const (
	searchLimit = 5
	urlSep      = " | "

	urlHelp = cmdPrefix + "url <username> or /<search>/"
)

var (
	urlRegexp = regexp.MustCompile(`((ftp|git|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(?:\/|\/([\w#!:.?+=&%@!\-\/]))?)`)
)

// Link represents a saved URL link.
type Link struct {
	Nick      string
	URL       string
	Server    string
	Channel   string
	CreatedAt time.Time
}

// LinkCommand interacts with the databased-saved URL links.
type LinkCommand struct {
	BaseCommand

	bot  *Scumbag
	conn *irc.Conn
	line *irc.Line
}

// NewLinkCommand returns a new LinkCommand instance.
func NewLinkCommand(bot *Scumbag, conn *irc.Conn, line *irc.Line) *LinkCommand {
	return &LinkCommand{bot: bot, conn: conn, line: line}
}

// Run is the command handler for "<cmdPrefix>url <nick_or_regex>"
func (cmd *LinkCommand) Run(args ...string) {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("LinkCommand.Run()")
		return
	}

	query := args[0]
	if query == "" {
		cmd.bot.Log.Debug("LinkCommand.Run(): No query")
		return
	}

	links, err := cmd.SearchLinks(query)
	if err != nil {
		cmd.bot.Log.WithField("error", err).Error("LinkCommand.Run()")
		return
	}

	response := make([]string, len(links))
	for i, link := range links {
		response[i] = link.URL
	}

	cmd.bot.Msg(cmd.conn, channel, strings.Join(response, urlSep))
}

// Help shows the command help.
func (cmd *LinkCommand) Help() {
	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("LinkCommand.Help()")
		return
	}

	cmd.bot.Msg(cmd.conn, channel, urlHelp)
}

// SaveURLs is called from a goroutine to save links from `conn.Config().Server` and `line`.
func (bot *Scumbag) SaveURLs(conn *irc.Conn, line *irc.Line) {
	link := NewLinkCommand(bot, conn, line)
	channel, err := link.Channel(line)
	if err != nil {
		bot.Log.WithField("err", err).Error("SaveURLs()")
		return
	}

	nick := line.Nick
	server := conn.Config().Server
	msg := line.Args[1]

	if urls := urlRegexp.FindAllString(msg, -1); urls != nil {
		if ignoredNick(bot, server, nick) {
			bot.Log.WithFields(log.Fields{"server": server, "nick": nick}).Debug("SaveURLs(): Ignored nick.")
			return
		}

		for _, url := range urls {
			var urlMatch string

			err := bot.DB.QueryRow("SELECT url FROM links WHERE url=$1 AND server=$2 AND channel=$3;", url, server, channel).Scan(&urlMatch)
			switch {
			case err == sql.ErrNoRows:
				// Link doesn't exist, so create one.
				if _, insertErr := bot.DB.Exec("INSERT INTO links(nick, url, server, channel, created_at) VALUES($1, $2, $3, $4, $5) RETURNING id;", nick, url, server, channel, line.Time); insertErr != nil {
					bot.Log.WithFields(log.Fields{"insertErr": insertErr}).Error("SaveURLs()")
				}
				bot.Log.WithFields(log.Fields{"URL": url, "server": server, "channel": channel}).Debug("SaveURLs(): New Link")

			case err != nil:
				bot.Log.WithFields(log.Fields{"err": err}).Error("SaveURLs()")

			default:
				bot.Log.WithFields(log.Fields{"url": url}).Debug("SaveURLs(): Existing Link")
			}
		}
	}
}

// SearchLinks searches the links database for query.
func (cmd *LinkCommand) SearchLinks(query string) ([]*Link, error) {
	var results []*Link

	channel, err := cmd.Channel(cmd.line)
	if err != nil {
		cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
		return results, err
	}

	// Regex search:  <cmdPrefix>url /imgur/
	if strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/") {
		urlQuery := strings.Replace(query, "/", "", 2)

		rows, err := cmd.bot.DB.Query(`SELECT nick, url, server, channel FROM links WHERE url ILIKE '%' || $1 || '%' AND server=$2 AND channel=$3 ORDER BY created_at DESC LIMIT $4;`, urlQuery, cmd.conn.Config().Server, channel, searchLimit)
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			link := Link{}
			err := rows.Scan(&link.Nick, &link.URL, &link.Server, &link.Channel)
			if err != nil {
				cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
				return nil, err
			}

			results = append(results, &link)
		}

		err = rows.Err()
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
			return nil, err
		}
	} else {
		// Nick search:  <cmdPrefix>url oshuma
		cmd.bot.Log.WithField("nick", query).Debug("LinkCommand.SearchLinks(): Nick Search")

		rows, err := cmd.bot.DB.Query(`SELECT nick, url, server, channel FROM links WHERE nick=$1 AND server=$2 AND channel=$3 ORDER BY created_at DESC LIMIT $4;`, query, cmd.conn.Config().Server, channel, searchLimit)
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			link := Link{}
			err := rows.Scan(&link.Nick, &link.URL, &link.Server, &link.Channel)
			if err != nil {
				cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
				return nil, err
			}

			results = append(results, &link)
		}

		err = rows.Err()
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
			return nil, err
		}
	}

	return results, nil
}

func ignoredNick(bot *Scumbag, server, nick string) bool {
	var result bool
	err := bot.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM ignored_nicks WHERE server=$1 AND nick=$2);", server, nick).Scan(&result)
	if err != nil && err != sql.ErrNoRows {
		bot.Log.WithField("err", err).Error("ignoredNick()")
		return false
	}
	return result
}
