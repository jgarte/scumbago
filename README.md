# Scumbag IRC Bot

## Dependencies

#### OS
* [Aspell](http://aspell.net/)
* [Figlet](http://www.figlet.org/)
* [Postgres](https://www.postgresql.org/)

#### Go
* go get [github.com/fluffle/goirc](https://github.com/fluffle/goirc)
* go get [github.com/Sirupsen/logrus](https://github.com/Sirupsen/logrus)
* go get [github.com/lib/pq](https://github.com/lib/pq)

## Setup

* Copy and edit `config/bot.json.example`
* Run `script/create_links_table.sql`

## Run

`$ go run main.go`
