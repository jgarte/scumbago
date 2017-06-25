package scumbag

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"database/sql"
	_ "github.com/lib/pq"

	log "github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	"github.com/jzelinskie/geddit"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/oauth2"
)

var VERSION = "1.2.0"
var BUILD = "HEAD"

const (
	CMD_ARG_REGEX = `(\w+)\s{1}\(sp\?\)`

	CMD_PREFIX = "?"

	CMD_ADMIN      = CMD_PREFIX + "admin"
	CMD_BEER       = CMD_PREFIX + "beer"
	CMD_FIGLET     = CMD_PREFIX + "fig"
	CMD_GITHUB     = CMD_PREFIX + "gh"
	CMD_HELP       = CMD_PREFIX + "help"
	CMD_REDDIT     = CMD_PREFIX + "reddit"
	CMD_SPELL      = CMD_PREFIX + "sp"
	CMD_TRUMP      = CMD_PREFIX + "trump"
	CMD_TWITTER    = CMD_PREFIX + "twitter"
	CMD_URBAN_DICT = CMD_PREFIX + "ud"
	CMD_URL        = CMD_PREFIX + "url"
	CMD_VERSION    = CMD_PREFIX + "version"
	CMD_WEATHER    = CMD_PREFIX + "weather"
	CMD_WIKI       = CMD_PREFIX + "wp"
	CMD_WOLFRAM    = CMD_PREFIX + "wolfram"

	// Default config file.
	CONFIG_FILE = "config/bot.json"

	// Default log file.
	LOG_FILE = "log/scumbag.log"
)

func Version() string {
	return fmt.Sprintf("scumbag v%s-%s", VERSION, BUILD)
}

type Scumbag struct {
	Config  *BotConfig
	DB      *sql.DB
	Log     *log.Logger
	Reddit  *geddit.Session
	Twitter *twitter.Client

	ircClients   map[string]*irc.Conn
	disconnected map[string]chan struct{}
}

func NewBot(configFile *string, logFilename *string) (*Scumbag, error) {
	botConfig, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	bot := &Scumbag{
		Config:       botConfig,
		disconnected: make(map[string]chan struct{}),
	}

	if err := bot.setupLogger(logFilename); err != nil {
		return nil, err
	}

	if err := bot.setupDatabase(); err != nil {
		return nil, err
	}

	bot.setupRedditSession()
	bot.setupTwitterClient()
	bot.setupIrcClients()
	bot.setupHandlers()

	return bot, nil
}

func (bot *Scumbag) Start() error {
	bot.Log.Info("Starting.")

	for _, client := range bot.ircClients {
		if err := client.Connect(); err != nil {
			bot.Log.WithFields(log.Fields{"err": err, "client": client}).Error("IRC Connection Error")
			return err
		}
	}

	return nil
}

func (bot *Scumbag) Wait() {
	bot.Log.Debug("Waiting...")
	for _, ch := range bot.disconnected {
		<-ch
	}
}

func (bot *Scumbag) Shutdown() {
	bot.Log.Info("Shutting down.")

	for server, client := range bot.ircClients {
		bot.Log.WithField("server", server).Debug("Shutdown()")
		client.Quit("Fuck you. Fuck you. You're cool. I'm out.")
	}

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

// Sends a PRIVMSG to `channel_or_nick` on `conn.Config().Server`'s client.
func (bot *Scumbag) Msg(conn *irc.Conn, channel_or_nick string, message string) {
	bot.ircClients[conn.Config().Server].Privmsg(channel_or_nick, message)
}

func (bot *Scumbag) setupLogger(logFilename *string) error {
	logFile, err := os.OpenFile(*logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
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

	return nil
}

func (bot *Scumbag) setupDatabase() error {
	bot.Log.Debug("setupDatabase()")

	databaseParams := fmt.Sprintf("host=%s sslmode=%s dbname=%s user=%s password=%s", bot.Config.Database.Host, bot.Config.Database.SSL, bot.Config.Database.Name, bot.Config.Database.User, bot.Config.Database.Password)
	session, err := sql.Open("postgres", databaseParams)
	if err != nil {
		bot.Log.WithField("error", err).Fatal("Database Connection Error")
		return err
	}
	bot.DB = session

	return nil
}

func (bot *Scumbag) setupRedditSession() {
	bot.Log.Debug("setupRedditSession()")

	bot.Reddit = geddit.NewSession(Version())
}

func (bot *Scumbag) setupTwitterClient() {
	bot.Log.Debug("setupTwitterClient()")

	oauthConfig := &oauth2.Config{}
	oauthToken := &oauth2.Token{AccessToken: bot.Config.Twitter.AccessToken}
	httpClient := oauthConfig.Client(oauth2.NoContext, oauthToken)

	bot.Twitter = twitter.NewClient(httpClient)
}

func (bot *Scumbag) setupIrcClients() {
	bot.Log.Debug("setupIrcClients()")

	bot.ircClients = make(map[string]*irc.Conn)

	for _, serverConfig := range bot.Config.Servers {
		clientConfig := irc.NewConfig(serverConfig.Name)
		clientConfig.Server = serverConfig.Server

		// Setup SSL and skip cert verify.
		if serverConfig.SSL {
			clientConfig.SSL = true
			clientConfig.SSLConfig = new(tls.Config)
			clientConfig.SSLConfig.InsecureSkipVerify = true
		}

		clientConfig.NewNick = func(n string) string { return n + "_" }

		bot.ircClients[serverConfig.Server] = irc.Client(clientConfig)
		bot.disconnected[serverConfig.Server] = make(chan struct{})
	}
}

func (bot *Scumbag) setupHandlers() {
	bot.Log.Debug("setupHandlers()")

	for server, client := range bot.ircClients {
		bot.Log.WithField("server", server).Debug("setupHandlers()")

		serverConfig, err := bot.Config.Server(server)
		if err != nil {
			bot.Log.WithField("err", err).Error("setupHandlers()")
			continue
		}
		bot.Log.WithField("serverConfig", serverConfig).Debug("setupHandlers()")

		client.HandleFunc("CONNECTED", func(conn *irc.Conn, line *irc.Line) {
			bot.Log.WithField("server", conn.Config().Server).Info("Connected to server.")
			for _, channel := range serverConfig.Channels {
				conn.Join(channel)
			}
		})

		client.HandleFunc("DISCONNECTED", func(conn *irc.Conn, line *irc.Line) {
			bot.Log.WithField("server", conn.Config().Server).Info("Disconnected.")
			close(bot.disconnected[conn.Config().Server])
		})

		client.HandleFunc("PRIVMSG", bot.msgHandler)
	}
}

// Handles normal PRIVMSG lines received from the server.
func (bot *Scumbag) msgHandler(conn *irc.Conn, line *irc.Line) {
	bot.Log.WithFields(log.Fields{
		"conn.Config().Server": conn.Config().Server,
		"line.Time":            line.Time,
		"line.Nick":            line.Nick,
		"line.Args":            line.Args,
	}).Debug("Channel message.")

	// These functions check the line text and act accordingly.
	go bot.SaveURLs(conn, line)
	go bot.SpellcheckLine(conn, line)

	// This function handles explicit bot commands.
	go bot.processCommands(conn, line)
}

func (bot *Scumbag) processCommands(conn *irc.Conn, line *irc.Line) {
	if len(line.Args) <= 0 {
		bot.Log.WithField("line", line).Debug("processCommands(): Line has no args")
		return
	}

	fields := strings.Fields(line.Args[1])

	if len(fields) <= 0 {
		bot.Log.WithField("line", line).Debug("processCommands(): No fields in line args")
		return
	}

	commandName := fields[0]
	args := strings.Join(fields[1:], " ") // FIXME: This is pretty hackish; just pass a slice of args to Command.Run() below.

	var command Command
	switch commandName {
	case CMD_ADMIN:
		command = NewAdminCommand(bot, conn, line)
	case CMD_BEER:
		command = NewBeerCommand(bot, conn, line)
	case CMD_FIGLET:
		command = NewFigletCommand(bot, conn, line)
	case CMD_GITHUB:
		command = NewGithubCommand(bot, conn, line)
	case CMD_HELP:
		command = NewHelpCommand(bot, conn, line)
	case CMD_REDDIT:
		command = NewRedditCommand(bot, conn, line)
	case CMD_SPELL:
		command = NewSpellcheckCommand(bot, conn, line)
	case CMD_TRUMP:
		command = NewTrumpCommand(bot, conn, line)
	case CMD_TWITTER:
		command = NewTwitterCommand(bot, conn, line)
	case CMD_URBAN_DICT:
		command = NewUrbanDictionaryCommand(bot, conn, line)
	case CMD_URL:
		command = NewLinkCommand(bot, conn, line)
	case CMD_VERSION:
		command = NewVersionCommand(bot, conn, line)
	case CMD_WEATHER:
		command = NewWeatherCommand(bot, conn, line)
	case CMD_WIKI:
		command = NewWikiCommand(bot, conn, line)
	case CMD_WOLFRAM:
		command = NewWolframAlphaCommand(bot, conn, line)
	}

	if command != nil {
		command.Run(args)
	}
}

func getContent(requestUrl string) ([]byte, error) {
	response, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	return content, nil
}
