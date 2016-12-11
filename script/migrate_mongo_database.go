package main

import (
	"fmt"
	"log"
	"time"

	"database/sql"
	_ "github.com/lib/pq"

	mgo "gopkg.in/mgo.v2"
)

const (
	MONGO_HOST             = "localhost"
	MONGO_DATABASE         = "scumbag"
	MONGO_LINKS_COLLECTION = "links"

	POSTGRES_DATABASE = "scumbag"
	POSTGRES_USERNAME = "ubuntu"
	POSTGRES_PASSWORD = "vagrant"
)

type Link struct {
	Nick      string
	Url       string
	Timestamp time.Time
}

func main() {
	mongoDb, err := mgo.Dial(MONGO_HOST)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoDb.Close()

	linksCollection := mongoDb.DB(MONGO_DATABASE).C(MONGO_LINKS_COLLECTION)

	var links []Link
	err = linksCollection.Find(nil).All(&links)
	if err != nil {
		log.Fatal(err)
	}

	postgresParams := fmt.Sprintf("dbname=%s user=%s password=%s", POSTGRES_DATABASE, POSTGRES_USERNAME, POSTGRES_PASSWORD)
	postgresDb, err := sql.Open("postgres", postgresParams)
	if err != nil {
		log.Fatal(err)
	}
	defer postgresDb.Close()

	migratedCount := 0

	for _, link := range links {
		err := postgresDb.QueryRow("SELECT * FROM links WHERE nick=$1 AND url=$2 AND created_at=$3 LIMIT 1;", link.Nick, link.Url, link.Timestamp).Scan(nil)

		// Only try to insert if no exact matching rows.
		if err == sql.ErrNoRows {
			if _, err := postgresDb.Exec("INSERT INTO links(nick, url, created_at) VALUES($1, $2, $3) RETURNING id;", link.Nick, link.Url, link.Timestamp); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("[%s] <%s> %s\n", link.Timestamp, link.Nick, link.Url)
			migratedCount += 1
		}
	}

	fmt.Printf("Migrated: %d\n", migratedCount)
}
