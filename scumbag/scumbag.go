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

const (
	CMD_ARG_REGEX = `(\w+)\s{1}\(sp\?\)`

	CMD_ADMIN      = "?admin"
	CMD_FIGLET     = "?fig"
	CMD_GITHUB     = "?gh"
	CMD_REDDIT     = "?reddit"
	CMD_SPELL      = "?sp"
	CMD_TRUMP      = "?trump"
	CMD_TWITTER    = "?twitter"
	CMD_URBAN_DICT = "?ud"
	CMD_URL        = "?url"
	CMD_WEATHER    = "?weather"
	CMD_WIKI       = "?wp"

	// Default config file.
	CONFIG_FILE = "config/bot.json"

	// Default log file.
	LOG_FILE = "log/scumbag.log"

	REDDIT_USER_AGENT = "scumbag v0.666"
)

type Command interface {
	Run(args ...string)
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

	databaseParams := fmt.Sprintf("dbname=%s user=%s password=%s", bot.Config.Database.Name, bot.Config.Database.User, bot.Config.Database.Password)
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

	bot.Reddit = geddit.NewSession(REDDIT_USER_AGENT)
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
	go bot.SaveURLs(line)
	go bot.SpellcheckLine(line)

	// This function handles explicit bot commands ("?url", "?sp", etc)
	go bot.processCommands(conn, line)
}

func (bot *Scumbag) processCommands(conn *irc.Conn, line *irc.Line) {
	if len(line.Args) <= 0 {
		bot.Log.WithField("line", line).Debug("processCommands(): Line has no args")
		return
	}

	channel := line.Args[0]
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
		command = &AdminCommand{bot: bot, channel: channel, conn: conn, line: line}
	case CMD_FIGLET:
		command = &FigletCommand{bot: bot, channel: channel, conn: conn}
	case CMD_GITHUB:
		command = &GithubCommand{bot: bot, channel: channel, conn: conn}
	case CMD_REDDIT:
		command = &RedditCommand{bot: bot, channel: channel, conn: conn}
	case CMD_SPELL:
		command = &SpellcheckCommand{bot: bot, channel: channel, conn: conn}
	case CMD_TRUMP:
		command = &TrumpCommand{bot: bot, channel: channel, conn: conn}
	case CMD_TWITTER:
		command = &TwitterCommand{bot: bot, channel: channel, conn: conn}
	case CMD_URBAN_DICT:
		command = &UrbanDictionaryCommand{bot: bot, channel: channel, conn: conn}
	case CMD_URL:
		command = &LinkCommand{bot: bot, channel: channel, conn: conn}
	case CMD_WEATHER:
		command = &WeatherCommand{bot: bot, channel: channel, conn: conn}
	case CMD_WIKI:
		command = &WikiCommand{bot: bot, channel: channel, conn: conn}
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
