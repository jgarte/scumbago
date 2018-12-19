# Scumbag IRC Bot

## Dependencies

#### OS
* [Aspell](http://aspell.net/)
* [Figlet](http://www.figlet.org/)
* [Postgres](https://www.postgresql.org/)

News command powered by [News API](https://newsapi.org/).

## Setup

* Copy and edit `config/bot.json.example`
* Run `script/001-create_links_table.sql`
* Run `script/002-add_server_and_channel_to_links.sql`
* Run `script/003-create_ignored_nicks_table.sql`

## Run

`$ go run main.go`
