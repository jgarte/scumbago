package scumbag

import (
	"crypto/tls"
	"fmt"
	"strings"

	irc "github.com/fluffle/goirc/client"
	mgo "gopkg.in/mgo.v2"
)

const (
	CMD_URL = "?url"
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

func NewBot() *Scumbag {
	dbConfig := &DatabaseConfig{
		Name:            "scumbag",
		Host:            "localhost",
		LinksCollection: "links",
	}

	botConfig := &BotConfig{
		Name:    "scumbag_go",
		Server:  "irc.literat.us:9999",
		Channel: "#scumbag",
		DB:      dbConfig,
	}

	bot := &Scumbag{Config: botConfig}

	bot.setupDatabase()
	bot.setupClient()
	bot.setupHandlers()

	return bot
}

func (bot *Scumbag) Start() {
	fmt.Println("-> Starting...")

	if err := bot.ircClient.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		quit <- true
	}

	// Wait for disconnect.
	<-quit
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

	clientConfig.NewNick = func(n string) string { return n + "_" }

	bot.ircClient = irc.Client(clientConfig)
}

func (bot *Scumbag) setupHandlers() {
	bot.ircClient.HandleFunc("CONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Printf("-> Connected to %s\n", bot.Config.Server)
		fmt.Printf("-> Joining %s\n", bot.Config.Channel)
		conn.Join(bot.Config.Channel)
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

	go SaveURLs(bot, line)
	go bot.processCommands(line)
}

func (bot *Scumbag) processCommands(line *irc.Line) {
	channel := line.Args[0]
	command, args := bot.getCommand(line)

	switch command {
	case CMD_URL:
		HandleUrlCommand(bot, channel, args)
	}
}

func (bot *Scumbag) getCommand(line *irc.Line) (string, string) {
	fields := strings.Fields(line.Args[1])

	command := fields[0]
	args := strings.Join(fields[1:], " ")

	return command, args
}
