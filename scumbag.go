package main

import (
	"crypto/tls"
	"fmt"
	"regexp"

	irc "github.com/fluffle/goirc/client"
)

var (
	// TODO: Map to store URLs keyed by nick. Store this in a database.
	urlDatabase = make(map[string][]string)

	// Channel to handle disconnect.
	quit = make(chan bool)
)

type Scumbag struct {
	Config *BotConfig
	client *irc.Conn
}

type DatabaseConfig struct {
	Name string "scumbag"
	Host string "localhost"
}

type BotConfig struct {
	Name     string
	Server   string
	Database *DatabaseConfig
}

func NewBot() *Scumbag {
	dbConfig := &DatabaseConfig{
		Name: "scumbag",
		Host: "localhost",
	}

	botConfig := &BotConfig{
		Name:     "scumbag_go",
		Server:   "irc.literat.us:9999",
		Database: dbConfig,
	}

	bot := Scumbag{Config: botConfig}

	bot.setupDatabase()
	bot.setupClient()
	bot.setupHandlers()

	return &bot
}

func (bot *Scumbag) setupDatabase() {
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

	fmt.Printf("-> URLs: %s\n", urlDatabase)
}

func (bot *Scumbag) saveURLs(nick string, msg string) {
	re := regexp.MustCompile(`((ftp|git|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(?:\/|\/([\w#!:.?+=&%@!\-\/]))?)`)

	if urls := re.FindAllString(msg, -1); urls != nil {
		for _, url := range urls {
			if notInArray(url, urlDatabase[nick]) {
				urlDatabase[nick] = append(urlDatabase[nick], url)
			}
		}
	}
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
	<-quit
}
