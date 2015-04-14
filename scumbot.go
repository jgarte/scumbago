package main

import (
	"crypto/tls"
	"fmt"

	irc "github.com/fluffle/goirc/client"
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

	client.HandleFunc("PRIVMSG", func(conn *irc.Conn, line *irc.Line) {
		nick := line.Nick
		msg := line.Args[1]
		time := line.Time
		fmt.Printf("-> MSG(%s): %s: %s\n", time, nick, msg)
	})

	// Tell client to connect.
	if err := client.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
	}

	// Wait for disconnect.
	<-quit
}
