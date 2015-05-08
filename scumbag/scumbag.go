package scumbag

import (
	"crypto/tls"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	mgo "gopkg.in/mgo.v2"
)

const (
	CMD_URL = "?url"

	// Default config file.
	CONFIG_FILE = "config/bot.json"

	// Default log file.
	LOG_FILE   = "log/scumbag.log"
	LOG_PREFIX = ""
)

var (
	// Channel to handle disconnect.
	quit = make(chan bool)
)

type Scumbag struct {
	Config *BotConfig
	Links  *mgo.Collection

	Log *log.Logger

	ircClient *irc.Conn
	dbSession *mgo.Session
}

func NewBot(configFile *string, logFilename *string) *Scumbag {
	botConfig := LoadConfig(configFile)
	bot := &Scumbag{Config: botConfig}

	bot.setupLogger(logFilename)
	bot.setupDatabase()
	bot.setupClient()
	bot.setupHandlers()

	return bot
}

func (bot *Scumbag) Start() {
	bot.Log.Info("Starting.")

	if err := bot.ircClient.Connect(); err != nil {
		bot.Log.WithField("error", err).Fatal("IRC Connection Error")
		return
	}

	// Wait for disconnect.
	<-quit
}

func (bot *Scumbag) Shutdown() {
	bot.Log.Info("Shutting down.")
	bot.dbSession.Close()
}

func (bot *Scumbag) setupLogger(logFilename *string) {
	logFile, err := os.OpenFile(*logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger := log.New()
	logger.Out = logFile
	logger.Level = log.DebugLevel

	bot.Log = logger
}

func (bot *Scumbag) setupDatabase() {
	session, err := mgo.Dial(bot.Config.Database.Host)
	if err != nil {
		bot.Log.WithField("error", err).Fatal("Database Connection Error")
		quit <- true
	}
	bot.dbSession = session

	databaseName := bot.Config.Database.Name
	linksCollection := bot.Config.Database.LinksCollection

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
		bot.Log.WithField("server", bot.Config.Server).Info("Connected to server.")
		bot.Log.WithField("channel", bot.Config.Channel).Info("Joined channel.")
		conn.Join(bot.Config.Channel)
	})

	bot.ircClient.HandleFunc("DISCONNECTED", func(conn *irc.Conn, line *irc.Line) {
		bot.Log.Info("Disconnected.")
		quit <- true
	})

	bot.ircClient.HandleFunc("PRIVMSG", bot.msgHandler)
}

// Handles normal PRIVMSG lines received from the server.
func (bot *Scumbag) msgHandler(conn *irc.Conn, line *irc.Line) {
	bot.Log.WithFields(log.Fields{
		"time": line.Time,
		"nick": line.Nick,
		"args": line.Args,
	}).Debug("Channel message.")

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
