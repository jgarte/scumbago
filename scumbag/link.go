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

// Handler for "?url <nick_or_regex>"
func (bot *Scumbag) HandleUrlCommand(channel string, args string) {
	links, err := bot.SearchLinks(args)
	if err != nil {
		bot.Log.WithField("error", err).Error("HandleUrlCommand()")
		return
	}

	response := make([]string, len(links))
	for i, link := range links {
		response[i] = link.Url
	}

	bot.Msg(channel, strings.Join(response, URL_SEP))
}

func (bot *Scumbag) SaveURLs(line *irc.Line) {
	nick := line.Nick
	msg := line.Args[1]

	if urls := urlRegexp.FindAllString(msg, -1); urls != nil {
		for _, url := range urls {
			var urlMatch string

			err := bot.db.QueryRow("SELECT url FROM links WHERE url=$1;", url).Scan(&urlMatch)
			switch {
			case err == sql.ErrNoRows:
				// Link doesn't exist, so create one.
				if _, insertErr := bot.db.Exec("INSERT INTO links(nick, url, created_at) VALUES($1, $2, $3) RETURNING id;", nick, url, line.Time); insertErr != nil {
					bot.Log.WithFields(log.Fields{"insertErr": insertErr}).Fatal("SaveURLs()")
				}
				bot.Log.WithField("URL", url).Debug("SaveURLs(): New Link")

			case err != nil:
				bot.Log.WithFields(log.Fields{"err": err}).Fatal("SaveURLs()")

			default:
				bot.Log.WithFields(log.Fields{"url": url}).Debug("SaveURLs(): Existing Link")
			}
		}
	}
}

func (bot *Scumbag) SearchLinks(query string) ([]Link, error) {
	var results []Link

	// Regex search:  ?url /imgur/
	if strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/") {
		urlQuery := strings.Replace(query, "/", "", 2)
		err := bot.Links.Find(bson.M{"url": &bson.RegEx{Pattern: urlQuery, Options: "i"}}).Sort("-timestamp").Limit(SEARCH_LIMIT).All(&results)
		if err != nil {
			return results, err
		}
	} else {
		// Nick search:  ?url oshuma
		bot.Log.WithField("nick", query).Debug("SearchLinks(): Nick Search")

		rows, err := bot.db.Query(`SELECT nick, url FROM links WHERE nick = $1 LIMIT $2;`, query, SEARCH_LIMIT)
		if err != nil {
			bot.Log.WithField("err", err).Fatal("SearchLinks()")
		}
		defer rows.Close()

		for rows.Next() {
			link := Link{}
			err := rows.Scan(&link.Nick, &link.Url)
			if err != nil {
				bot.Log.WithField("err", err).Fatal("SearchLinks()")
			}

			results = append(results, link)
		}

		err = rows.Err()
		if err != nil {
			bot.Log.WithField("err", err).Fatal("SearchLinks()")
		}
	}

	return results, nil
}
