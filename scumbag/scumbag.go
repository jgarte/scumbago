package scumbag

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"database/sql"
	_ "github.com/lib/pq"

	log "github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	"github.com/jzelinskie/geddit"
)

const (
	CMD_ARG_REGEX = `(\w+)\s{1}\(sp\?\)`

	CMD_ADMIN  = "?admin"
	CMD_FIGLET = "?fig"
	CMD_REDDIT = "?reddit"
	CMD_SPELL  = "?sp"
	CMD_URL    = "?url"

	// Default config file.
	CONFIG_FILE = "config/bot.json"

	// Default log file.
	LOG_FILE = "log/scumbag.log"

	REDDIT_USER_AGENT = "scumbag"
)

var (
	// Channel to handle disconnect.
	quit = make(chan bool)
)

type Scumbag struct {
	Config *BotConfig
	DB     *sql.DB
	Log    *log.Logger
	Reddit *geddit.Session

	ircClient *irc.Conn
}

func NewBot(configFile *string, logFilename *string) *Scumbag {
	botConfig := LoadConfig(configFile)
	bot := &Scumbag{Config: botConfig}

	bot.setupLogger(logFilename)
	bot.setupDatabase()
	bot.setupRedditSession()
	bot.setupClient()
	bot.setupHandlers()

	return bot
}

func (bot *Scumbag) Start() {
	bot.Log.Info("Starting.")

	go func() {
		bot.Log.Debug("Setting up signal handler.")
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt)
		<-signalChannel
		bot.Log.Debug("SIGINT received.")
		quit <- true
	}()

	if err := bot.ircClient.Connect(); err != nil {
		bot.Log.WithField("error", err).Fatal("IRC Connection Error")
		quit <- true
	}

	// Wait for disconnect.
	<-quit
}

func (bot *Scumbag) Shutdown() {
	bot.Log.Info("Shutting down.")
	bot.DB.Close()
}

func (bot *Scumbag) Admin(nick string) bool {
	for _, n := range bot.Config.Admins {
		if n == nick {
			return true
		}
	}

	return false
}

// Sends a PRIVMSG to `channel_or_nick`.
func (bot *Scumbag) Msg(channel_or_nick string, message string) {
	bot.ircClient.Privmsg(channel_or_nick, message)
}

func (bot *Scumbag) setupLogger(logFilename *string) {
	logFile, err := os.OpenFile(*logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger := log.New()
	logger.Out = logFile

	switch bot.Config.LogLevel {
	case "Panic":
		logger.Level = log.PanicLevel
	case "Fatal":
		logger.Level = log.FatalLevel
	case "Error":
		logger.Level = log.ErrorLevel
	case "Warn":
		logger.Level = log.WarnLevel
	case "Info":
		logger.Level = log.InfoLevel
	case "Debug":
		logger.Level = log.DebugLevel
	default:
		logger.Level = log.InfoLevel
	}

	bot.Log = logger
}

func (bot *Scumbag) setupDatabase() {
	databaseParams := fmt.Sprintf("dbname=%s user=%s password=%s", bot.Config.Database.Name, bot.Config.Database.User, bot.Config.Database.Password)
	session, err := sql.Open("postgres", databaseParams)
	if err != nil {
		bot.Log.WithField("error", err).Fatal("Database Connection Error")
		quit <- true
	}
	bot.DB = session
}

func (bot *Scumbag) setupRedditSession() {
	bot.Reddit = geddit.NewSession(REDDIT_USER_AGENT)
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
		conn.Join(bot.Config.Channel)
		bot.Log.WithField("channel", bot.Config.Channel).Info("Joined channel.")
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

	// These functions check the line text and act accordingly.
	go bot.SaveURLs(line)
	go bot.SpellcheckLine(line)

	// This function handles explicit bot commands ("?url", "?sp", etc)
	go bot.processCommands(line)
}

func (bot *Scumbag) processCommands(line *irc.Line) {
	channel := line.Args[0]
	command, args := bot.getCommand(line)

	switch command {
	case CMD_ADMIN:
		bot.HandleAdminCommand(channel, args, line)
	case CMD_FIGLET:
		bot.HandleFigletCommand(channel, args)
	case CMD_REDDIT:
		bot.HandleRedditCommand(channel, args)
	case CMD_SPELL:
		bot.HandleSpellCommand(channel, args)
	case CMD_URL:
		bot.HandleUrlCommand(channel, args)
	}
}

func (bot *Scumbag) getCommand(line *irc.Line) (string, string) {
	fields := strings.Fields(line.Args[1])

	command := fields[0]
	args := strings.Join(fields[1:], " ")

	return command, args
}
