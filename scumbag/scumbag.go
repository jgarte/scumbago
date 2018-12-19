package scumbag

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"database/sql"
	// Don't need a named import for a database driver.
	_ "github.com/lib/pq"

	irc "github.com/fluffle/goirc/client"
	"github.com/jzelinskie/geddit"
	newsapi "github.com/kaelanb/newsapi-go"
	log "github.com/sirupsen/logrus"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/oauth2"
)

// Version is a rarely updated string...
var Version = "1.6.0"

// BuildTag is updated from the current git SHA when the Docker image is pushed.
var BuildTag = "HEAD"

const (
	cmdArgRegex = `(\w+)\s{1}\(sp\?\)`

	cmdPrefix = "?"

	cmdAdmin     = cmdPrefix + "admin"
	cmdBeer      = cmdPrefix + "beer"
	cmdFiglet    = cmdPrefix + "fig"
	cmdGithub    = cmdPrefix + "gh"
	cmdHelp      = cmdPrefix + "help"
	cmdMovie     = cmdPrefix + "movie"
	cmdNews      = cmdPrefix + "news"
	cmdReddit    = cmdPrefix + "reddit"
	cmdSpell     = cmdPrefix + "sp"
	cmdTwitter   = cmdPrefix + "twitter"
	cmdURL       = cmdPrefix + "url"
	cmdUrbanDict = cmdPrefix + "ud"
	cmdVersion   = cmdPrefix + "version"
	cmdWeather   = cmdPrefix + "weather"
	cmdWiki      = cmdPrefix + "wp"
	cmdWolfram   = cmdPrefix + "wolfram"

	// ConfigFile is the path to the default config file.
	ConfigFile = "config/bot.json"

	// LogFile is the path to the default log file.
	LogFile = "log/scumbag.log"
)

// VersionString returns a formatted version string.
func VersionString() string {
	return fmt.Sprintf("scumbag v%s-%s", Version, BuildTag)
}

// Scumbag is the main bot struct.
type Scumbag struct {
	Config  *BotConfig
	DB      *sql.DB
	Log     *log.Logger
	News    *newsapi.Client
	Reddit  *geddit.Session
	Twitter *twitter.Client

	ircClients   map[string]*irc.Conn
	disconnected map[string]chan struct{}
}

// NewBot returns a new Scumbag instance.
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

	bot.setupNewsClient()
	bot.setupRedditSession()
	bot.setupTwitterClient()
	bot.setupIrcClients()
	bot.setupHandlers()

	return bot, nil
}

// Start connects the bot to the configured IRC servers.
func (bot *Scumbag) Start() error {
	bot.Log.Info("Starting.")

	// Keeps track of how many servers in bot.ircClients that fail to connect.
	connectErrors := 0

	for _, client := range bot.ircClients {
		if err := bot.connectClient(client); err != nil {
			bot.Log.WithFields(log.Fields{"err": err, "client": client}).Error("IRC Connection Error")
			connectErrors++
		}
	}

	if len(bot.ircClients) == connectErrors {
		return errors.New("could not connect to any servers")
	}

	return nil
}

// Wait keeps the bot running until a disconnect is received.
func (bot *Scumbag) Wait() {
	bot.Log.Debug("Waiting...")
	for _, ch := range bot.disconnected {
		<-ch
	}
}

// Shutdown sanely shuts down the bot.
func (bot *Scumbag) Shutdown() {
	bot.Log.Info("Shutting down.")

	for server, client := range bot.ircClients {
		bot.Log.WithField("server", server).Debug("Shutdown()")
		if client.Connected() {
			client.Quit("Fuck you. Fuck you. You're cool. I'm out.")
		} else {
			close(bot.disconnected[server])
		}
	}

	bot.DB.Close()
}

// Admin returns true if the given nick string is an admin.
func (bot *Scumbag) Admin(nick string) bool {
	for _, n := range bot.Config.Admins {
		if n == nick {
			return true
		}
	}

	return false
}

// Msg sends a PRIVMSG to `channel_or_nick` on `conn.Config().Server`'s client.
func (bot *Scumbag) Msg(conn *irc.Conn, channelOrNick string, message string) {
	bot.ircClients[conn.Config().Server].Privmsg(channelOrNick, message)
}

func (bot *Scumbag) connectClient(client *irc.Conn) error {
	if err := client.Connect(); err != nil {
		return err
	}
	return nil
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

func (bot *Scumbag) setupNewsClient() {
	bot.Log.Debug("setupNewsClient()")
	bot.News = newsapi.New(bot.Config.News.Key)
}

func (bot *Scumbag) setupRedditSession() {
	bot.Log.Debug("setupRedditSession()")

	bot.Reddit = geddit.NewSession(VersionString())
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

			err := bot.connectClient(client)
			if err != nil {
				return
			}

			bot.disconnected[serverConfig.Server] = make(chan struct{})
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
	}).Debug("Scumbag.msgHandler(): Channel message.")

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
	case cmdAdmin:
		command = NewAdminCommand(bot, conn, line)
	case cmdBeer:
		command = NewBeerCommand(bot, conn, line)
	case cmdFiglet:
		command = NewFigletCommand(bot, conn, line)
	case cmdGithub:
		command = NewGithubCommand(bot, conn, line)
	case cmdHelp:
		command = NewHelpCommand(bot, conn, line)
	case cmdMovie:
		command = NewMovieCommand(bot, conn, line)
	case cmdNews:
		command = NewNewsCommand(bot, conn, line)
	case cmdReddit:
		command = NewRedditCommand(bot, conn, line)
	case cmdSpell:
		command = NewSpellcheckCommand(bot, conn, line)
	case cmdTwitter:
		command = NewTwitterCommand(bot, conn, line)
	case cmdUrbanDict:
		command = NewUrbanDictionaryCommand(bot, conn, line)
	case cmdURL:
		command = NewLinkCommand(bot, conn, line)
	case cmdVersion:
		command = NewVersionCommand(bot, conn, line)
	case cmdWeather:
		command = NewWeatherCommand(bot, conn, line)
	case cmdWiki:
		command = NewWikiCommand(bot, conn, line)
	case cmdWolfram:
		command = NewWolframAlphaCommand(bot, conn, line)
	default:
		bot.Log.WithField("commandName", commandName).Debug("Scumbag.processCommands(): Unknown command")
	}

	if command != nil {
		command.Run(args)
	}
}

func getContent(requestURL string) ([]byte, error) {
	response, err := http.Get(requestURL)
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
