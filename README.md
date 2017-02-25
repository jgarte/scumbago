# Scumbag IRC Bot

## Dependencies

#### OS
* [Aspell](http://aspell.net/)
* [Figlet](http://www.figlet.org/)
* [Postgres](https://www.postgresql.org/)

#### Go
* go get [github.com/jzelinskie/geddit](https://github.com/jzelinskie/geddit)
* go get [github.com/fluffle/goirc](https://github.com/fluffle/goirc)
* go get [github.com/Sirupsen/logrus](https://github.com/Sirupsen/logrus)
* go get [golang.org/x/oauth2](https://godoc.org/golang.org/x/oauth2)
* go get [github.com/lib/pq](https://github.com/lib/pq)
* go get [github.com/kennygrant/sanitize](https://github.com/kennygrant/sanitize)
* go get [github.com/dghubble/go-twitter/twitter](https://github.com/dghubble/go-twitter)

## Setup

* Copy and edit `config/bot.json.example`
* Run `script/create_links_table.sql`

## Run

`$ go run main.go`
