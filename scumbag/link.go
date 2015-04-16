package scumbag

import (
	"fmt"
	"regexp"
	"time"

	irc "github.com/fluffle/goirc/client"
	"gopkg.in/mgo.v2/bson"
)

type Link struct {
	Nick      string
	Url       string
	Timestamp time.Time
}

func SaveURLs(bot *Scumbag, line *irc.Line) {
	nick := line.Nick
	msg := line.Args[1]

	re := regexp.MustCompile(`((ftp|git|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(?:\/|\/([\w#!:.?+=&%@!\-\/]))?)`)

	if urls := re.FindAllString(msg, -1); urls != nil {
		for _, url := range urls {
			link := Link{}

			if err := bot.Links.Find(bson.M{"nick": nick, "url": url}).One(&link); err != nil {
				// Link doesn't exist, so create one.
				link.Nick = nick
				link.Url = url
				link.Timestamp = line.Time

				if err := bot.Links.Insert(link); err != nil {
					fmt.Println("ERROR: ", err)
					continue // With the next URL match.
				} else {
					fmt.Printf("-> LINK: %v\n", link)
				}
			}
		}
	}
}

func SearchLinks(query string) []string {
	panic("Finish this.")
	return make([]string, 1)
}
