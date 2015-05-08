package scumbag

import (
	"regexp"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	"gopkg.in/mgo.v2/bson"
)

const (
	SEARCH_LIMIT = 5
)

var (
	urlRegex = regexp.MustCompile(`((ftp|git|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(?:\/|\/([\w#!:.?+=&%@!\-\/]))?)`)
)

type Link struct {
	Nick      string
	Url       string
	Timestamp time.Time
}

func SaveURLs(bot *Scumbag, line *irc.Line) {
	nick := line.Nick
	msg := line.Args[1]

	if urls := urlRegex.FindAllString(msg, -1); urls != nil {
		for _, url := range urls {
			var link Link

			if err := bot.Links.Find(bson.M{"nick": nick, "url": url}).One(&link); err != nil {
				// Link doesn't exist, so create one.
				link.Nick = nick
				link.Url = url
				link.Timestamp = line.Time

				if err := bot.Links.Insert(link); err != nil {
					bot.Log.WithField("error", err).Error("SaveURLs()")
					continue // With the next URL match.
				} else {
					bot.Log.WithField("link", link).Info("-> Link")
				}
			}
		}
	}
}

func SearchLinks(bot *Scumbag, query string) ([]Link, error) {
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
