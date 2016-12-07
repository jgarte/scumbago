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
					bot.Log.WithFields(log.Fields{"insertErr": insertErr}).Fatal(insertErr)
				}
				bot.Log.WithField("URL", url).Info("-> New Link")

			case err != nil:
				bot.Log.WithFields(log.Fields{"err": err}).Fatal(err)

			default:
				bot.Log.WithFields(log.Fields{"url": url}).Info("-> Existing Link")
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
		err := bot.Links.Find(bson.M{"nick": query}).Sort("-timestamp").Limit(SEARCH_LIMIT).All(&results)
		if err != nil {
			return results, err
		}
	}

	return results, nil
}
