# Scumbag IRC Bot

## Dependencies

#### OS
* [Aspell](http://aspell.net/)
* [Figlet](http://www.figlet.org/)
* [Postgres](https://www.postgresql.org/)

#### Go
* go get -u [github.com/jzelinskie/geddit](https://github.com/jzelinskie/geddit)
* go get -u [github.com/fluffle/goirc](https://github.com/fluffle/goirc)
* go get -u [github.com/Sirupsen/logrus](https://github.com/Sirupsen/logrus)
* go get -u [golang.org/x/oauth2](https://godoc.org/golang.org/x/oauth2)
* go get -u [github.com/lib/pq](https://github.com/lib/pq)
* go get -u [github.com/kennygrant/sanitize](https://github.com/kennygrant/sanitize)
* go get -u [github.com/dghubble/go-twitter/twitter](https://github.com/dghubble/go-twitter)
* go get -u [github.com/kaelanb/newsapi-go](https://github.com/kaelanb/newsapi-go)

News command powered by [News API](https://newsapi.org/).

## Setup

* Copy and edit `config/bot.json.example`
* Run `script/001-create_links_table.sql`
* Run `script/002-add_server_and_channel_to_links.sql`
* Run `script/003-create_ignored_nicks_table.sql`

## Run

`$ go run main.go`
