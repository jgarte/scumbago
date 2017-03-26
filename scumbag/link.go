package scumbag

import (
	"database/sql"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
)

const (
	SEARCH_LIMIT = 5
	URL_SEP      = " | "
)

var (
	urlRegexp = regexp.MustCompile(`((ftp|git|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(?:\/|\/([\w#!:.?+=&%@!\-\/]))?)`)
)

type Link struct {
	Nick      string
	Url       string
	CreatedAt time.Time
}

type LinkCommand struct {
	bot     *Scumbag
	channel string
	conn    *irc.Conn
}

// Handler for "?url <nick_or_regex>"
func (cmd *LinkCommand) Run(args ...string) {
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
		response[i] = link.Url
	}

	cmd.bot.Msg(cmd.conn, cmd.channel, strings.Join(response, URL_SEP))
}

// Called from a goroutine to save links from `line`.
func (bot *Scumbag) SaveURLs(line *irc.Line) {
	nick := line.Nick
	msg := line.Args[1]

	if urls := urlRegexp.FindAllString(msg, -1); urls != nil {
		for _, url := range urls {
			var urlMatch string

			err := bot.DB.QueryRow("SELECT url FROM links WHERE url=$1;", url).Scan(&urlMatch)
			switch {
			case err == sql.ErrNoRows:
				// Link doesn't exist, so create one.
				if _, insertErr := bot.DB.Exec("INSERT INTO links(nick, url, created_at) VALUES($1, $2, $3) RETURNING id;", nick, url, line.Time); insertErr != nil {
					bot.Log.WithFields(log.Fields{"insertErr": insertErr}).Error("SaveURLs()")
				}
				bot.Log.WithField("URL", url).Debug("SaveURLs(): New Link")

			case err != nil:
				bot.Log.WithFields(log.Fields{"err": err}).Error("SaveURLs()")

			default:
				bot.Log.WithFields(log.Fields{"url": url}).Debug("SaveURLs(): Existing Link")
			}
		}
	}
}

func (cmd *LinkCommand) SearchLinks(query string) ([]*Link, error) {
	var results []*Link

	// Regex search:  ?url /imgur/
	if strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/") {
		urlQuery := strings.Replace(query, "/", "", 2)

		rows, err := cmd.bot.DB.Query(`SELECT nick, url FROM links WHERE url ILIKE '%' || $1 || '%' ORDER BY created_at DESC LIMIT $2;`, urlQuery, SEARCH_LIMIT)
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			link := Link{}
			err := rows.Scan(&link.Nick, &link.Url)
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
		// Nick search:  ?url oshuma
		cmd.bot.Log.WithField("nick", query).Debug("LinkCommand.SearchLinks(): Nick Search")

		rows, err := cmd.bot.DB.Query(`SELECT nick, url FROM links WHERE nick = $1 ORDER BY created_at DESC LIMIT $2;`, query, SEARCH_LIMIT)
		if err != nil {
			cmd.bot.Log.WithField("err", err).Error("LinkCommand.SearchLinks()")
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			link := Link{}
			err := rows.Scan(&link.Nick, &link.Url)
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
