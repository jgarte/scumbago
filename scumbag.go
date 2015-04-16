package main

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"time"

	irc "github.com/fluffle/goirc/client"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// Channel to handle disconnect.
	quit = make(chan bool)
)

type Scumbag struct {
	Config *BotConfig
	Links  *mgo.Collection

	client *irc.Conn
}

type DatabaseConfig struct {
	Name            string
	Host            string
	LinksCollection string
}

type BotConfig struct {
	Name   string
	Server string
	DB     *DatabaseConfig
}

func NewBot() *Scumbag {
	dbConfig := &DatabaseConfig{
		Name:            "scumbag",
		Host:            "localhost",
		LinksCollection: "links",
	}

	botConfig := &BotConfig{
		Name:   "scumbag_go",
		Server: "irc.literat.us:9999",
		DB:     dbConfig,
	}

	bot := Scumbag{Config: botConfig}

	bot.setupDatabase()
	bot.setupClient()
	bot.setupHandlers()

	return &bot
}

func (bot *Scumbag) setupDatabase() {
	session, err := mgo.Dial(bot.Config.DB.Host)
	if err != nil {
		fmt.Printf("Database connection error: %s\n", err)
		quit <- true
	}

	databaseName := bot.Config.DB.Name
	linksCollection := bot.Config.DB.LinksCollection
	bot.Links = session.DB(databaseName).C(linksCollection)
}

func (bot *Scumbag) setupClient() {
	clientConfig := irc.NewConfig(bot.Config.Name)
	clientConfig.Server = bot.Config.Server

	// Setup SSL and skip cert verify.
	clientConfig.SSL = true
	clientConfig.SSLConfig = new(tls.Config)
	clientConfig.SSLConfig.InsecureSkipVerify = true

	clientConfig.NewNick = func(n string) string { return n + "^" }

	bot.client = irc.Client(clientConfig)
}

func (bot *Scumbag) setupHandlers() {
	bot.client.HandleFunc("CONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("-> Connecting to #scumbag")
		conn.Join("#scumbag")
	})

	bot.client.HandleFunc("DISCONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println(" -> Disconnected...")
		quit <- true
	})

	bot.client.HandleFunc("PRIVMSG", bot.msgHandler)
}

// Handles normal PRIVMSG lines received from the server.
func (bot *Scumbag) msgHandler(conn *irc.Conn, line *irc.Line) {
	time := line.Time
	nick := line.Nick
	msg := line.Args[1]

	fmt.Printf("<- MSG(%s) %s: %s\n", time, nick, msg)

	bot.saveURLs(nick, msg)
}

func (bot *Scumbag) saveURLs(nick string, msg string) {
	re := regexp.MustCompile(`((ftp|git|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(?:\/|\/([\w#!:.?+=&%@!\-\/]))?)`)

	if urls := re.FindAllString(msg, -1); urls != nil {
		for _, url := range urls {
			link := Link{}

			if err := bot.Links.Find(bson.M{"nick": nick, "url": url}).One(&link); err != nil {
				// Link doesn't exist, so create one.
				link.Nick = nick
				link.Url = url
				link.Timestamp = time.Now()

				if err := bot.Links.Insert(link); err != nil {
					fmt.Println("ERROR: ", err)
					continue // With the next URL match.
				}
			}
		}
	}
}

type Link struct {
	Nick      string
	Url       string
	Timestamp time.Time
}

func notInArray(value string, array []string) bool {
	return !inArray(value, array)
}

func inArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println("-> Starting...")

	bot := NewBot()

	if err := bot.client.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		quit <- true
	}

	// Wait for disconnect.
	if <-quit {
		fmt.Println("-> Shutting down...")
	}
}
