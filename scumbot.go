package main

import (
	"crypto/tls"
	"fmt"
	"regexp"

	irc "github.com/fluffle/goirc/client"
)

type Scumbot struct {
	Client *irc.Conn
	Config map[string]string
}

var (
	// TODO: Map to store URLs keyed by nick. Store this in a database.
	urlDatabase = make(map[string][]string)

	bot Scumbot

	// Channel to handle disconnect.
	quit = make(chan bool)
)

func NewBot() Scumbot {
	fmt.Println("-> Setting up bot...")

	botConfig := map[string]string{
		"name":      "scumbot",
		"database":  "scumbag",
		"ircServer": "irc.literat.us:9999",
	}

	clientConfig := irc.NewConfig(botConfig["name"])
	clientConfig.Server = botConfig["ircServer"]

	// Setup SSL and skip cert verify.
	clientConfig.SSL = true
	clientConfig.SSLConfig = new(tls.Config)
	clientConfig.SSLConfig.InsecureSkipVerify = true

	clientConfig.NewNick = func(n string) string { return n + "^" }

	return Scumbot{Client: irc.Client(clientConfig), Config: botConfig}
}

func main() {
	fmt.Println("-> Starting...")

	bot = NewBot()
	setupHandlers()

	if err := bot.Client.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		quit <- true
	}

	// Wait for disconnect.
	<-quit
}

func setupHandlers() {
	fmt.Println("-> Setting up handlers...")

	bot.Client.HandleFunc("CONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("-> Connecting to #scumbot")
		conn.Join("#scumbot")
	})

	bot.Client.HandleFunc("DISCONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println(" -> Disconnected...")
		quit <- true
	})

	bot.Client.HandleFunc("PRIVMSG", msgHandler)
}

func msgHandler(conn *irc.Conn, line *irc.Line) {
	time := line.Time
	nick := line.Nick
	msg := line.Args[1]

	fmt.Printf("<- MSG(%s) %s: %s\n", time, nick, msg)

	bot.saveURLs(nick, msg)

	fmt.Printf("-> URLs: %s\n", urlDatabase)
}

func (b Scumbot) saveURLs(nick string, msg string) {
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
