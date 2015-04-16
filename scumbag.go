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

	ircClient *irc.Conn
	dbSession *mgo.Session
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

func (bot *Scumbag) Start() {
	fmt.Println("-> Starting...")

	if err := bot.ircClient.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		quit <- true
	}
}

func (bot *Scumbag) Shutdown() {
	fmt.Println("-> Shutting down...")
	bot.dbSession.Close()
}

func (bot *Scumbag) setupDatabase() {
	session, err := mgo.Dial(bot.Config.DB.Host)
	if err != nil {
		fmt.Printf("Database connection error: %s\n", err)
		quit <- true
	}
	bot.dbSession = session

	databaseName := bot.Config.DB.Name
	linksCollection := bot.Config.DB.LinksCollection

	bot.Links = bot.dbSession.DB(databaseName).C(linksCollection)
}

func (bot *Scumbag) setupClient() {
	clientConfig := irc.NewConfig(bot.Config.Name)
	clientConfig.Server = bot.Config.Server

	// Setup SSL and skip cert verify.
	clientConfig.SSL = true
	clientConfig.SSLConfig = new(tls.Config)
	clientConfig.SSLConfig.InsecureSkipVerify = true

	clientConfig.NewNick = func(n string) string { return n + "^" }

	bot.ircClient = irc.Client(clientConfig)
}

func (bot *Scumbag) setupHandlers() {
	bot.ircClient.HandleFunc("CONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("-> Connecting to #scumbag")
		conn.Join("#scumbag")
	})

	bot.ircClient.HandleFunc("DISCONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println(" -> Disconnected...")
		quit <- true
	})

	bot.ircClient.HandleFunc("PRIVMSG", bot.msgHandler)
}

// Handles normal PRIVMSG lines received from the server.
func (bot *Scumbag) msgHandler(conn *irc.Conn, line *irc.Line) {
	fmt.Printf("<- MSG(%v) %v: %v\n", line.Time, line.Nick, line.Args)

	bot.saveURLs(line)
}

func (bot *Scumbag) saveURLs(line *irc.Line) {
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

func main() {
	bot := NewBot()
	bot.Start()
	defer bot.Shutdown()

	// Wait for disconnect.
	<-quit
}
