package main

import (
	"crypto/tls"
	"fmt"
	"regexp"

	irc "github.com/fluffle/goirc/client"
)

var (
	// TODO: Map to store URLs keyed by nick. Store these in a database.
	urlDatabase = make(map[string][]string)
)

func main() {
	fmt.Println("-> Starting...")

	config := irc.NewConfig("scumbot")
	config.Server = "irc.example.com:1234"

	// Setup SSL and skip cert verify.
	config.SSL = true
	config.SSLConfig = new(tls.Config)
	config.SSLConfig.InsecureSkipVerify = true

	config.NewNick = func(n string) string { return n + "^" }

	client := irc.Client(config)

	// Channel to handle disconnect.
	quit := make(chan bool)

	client.HandleFunc("CONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println("-> Connecting to #scumbot")
		conn.Join("#scumbot")
	})

	client.HandleFunc("DISCONNECTED", func(conn *irc.Conn, line *irc.Line) {
		fmt.Println(" -> Disconnected...")
		quit <- true
	})

	client.HandleFunc("PRIVMSG", msgHandler)

	// Tell client to connect.
	if err := client.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		quit <- true
	}

	// Wait for disconnect.
	<-quit
}

func msgHandler(conn *irc.Conn, line *irc.Line) {
	time := line.Time
	nick := line.Nick
	msg := line.Args[1]

	fmt.Printf("<- MSG(%s) %s: %s\n", time, nick, msg)

	saveURLs(nick, msg)

	fmt.Printf("-> URLs: %s\n", urlDatabase)
}

func saveURLs(nick string, msg string) {
	re := regexp.MustCompile(`((ftp|git|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(?:\/|\/([\w#!:.?+=&%@!\-\/]))?)`)

	if urlMatches := re.FindAllString(msg, -1); urlMatches != nil {
		for _, url := range urlMatches {
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
