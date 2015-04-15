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

type Scumbot struct {
	Client *irc.Conn
	Config *BotConfig
}

type DatabaseConfig struct {
	Name string "scumbag"
	Host string "localhost"
}

type BotConfig struct {
	Name     string "scumbot"
	Server   string "irc.literat.us:9999"
	Database *DatabaseConfig
}

func NewBot() *Scumbot {
	dbConfig := &DatabaseConfig{
		Name: "scumbag",
		Host: "localhost",
	}

	botConfig := &BotConfig{
		Name:     "scumbot",
		Server:   "irc.literat.us:9999",
		Database: dbConfig,
	}

	clientConfig := irc.NewConfig(botConfig.Name)
	clientConfig.Server = botConfig.Server

	// Setup SSL and skip cert verify.
	clientConfig.SSL = true
	clientConfig.SSLConfig = new(tls.Config)
	clientConfig.SSLConfig.InsecureSkipVerify = true

	clientConfig.NewNick = func(n string) string { return n + "^" }

	bot := Scumbot{Client: irc.Client(clientConfig), Config: botConfig}
	bot.setupHandlers()

	return &bot
}

func (bot *Scumbot) setupHandlers() {
	bot.Client.HandleFunc("CONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("-> Connecting to #scumbot")
		conn.Join("#scumbot")
	})

	bot.Client.HandleFunc("DISCONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println(" -> Disconnected...")
		quit <- true
	})

	bot.Client.HandleFunc("PRIVMSG", bot.msgHandler)
}

// Handles normal PRIVMSG lines received from the server.
func (bot *Scumbot) msgHandler(conn *irc.Conn, line *irc.Line) {
	time := line.Time
	nick := line.Nick
	msg := line.Args[1]

	fmt.Printf("<- MSG(%s) %s: %s\n", time, nick, msg)

	bot.saveURLs(nick, msg)

	fmt.Printf("-> URLs: %s\n", urlDatabase)
}

func (bot *Scumbot) saveURLs(nick string, msg string) {
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

	if err := bot.Client.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		quit <- true
	}

	// Wait for disconnect.
	<-quit
}
