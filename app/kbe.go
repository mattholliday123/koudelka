package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	_"github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
)

var WEBSITE_LIST = [200]string{"https://en.wikipedia.org/w/index.php?title=Special:NewPages&feed=rss",
"https://en.wikipedia.org/w/api.php?hidebots=1&hidecategorization=1&hideWikibase=1&urlversion=1&days=1&limit=50&action=feedrecentchanges&feedformat=rss", "https://www.theguardian.com/us/rss","https://www.espn.com/espn/rss/news", "https://feeds.bbci.co.uk/news/rss.xml", "http://rss.cnn.com/rss/cnn_topstories.rss"}

type InvertedIndex map[string][]int

func stripHTML(input string) string {
    // Basic regex to remove everything inside < >
    re := regexp.MustCompile(`<(.|\n)*?>`)
    return re.ReplaceAllString(input, "")
}

func tokenize(input string) []string {
    // Simple implementation of Buttcher Sec 3.2: 
    // Convert to lower and split by non-alphanumeric characters
    re := regexp.MustCompile(`[a-zA-Z0-9]+`)
    return re.FindAllString(strings.ToLower(input), -1)
}
func buildIndex(db *sql.DB) InvertedIndex {

	index := make(InvertedIndex)

	rows, err := db.Query("SELECT id, body FROM docs")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Next()

	for rows.Next(){
		var id int
		var body string
		rows.Scan(&id, &body)
		plainText := stripHTML(body)
		terms := tokenize(plainText)

		for _, term := range terms {
			list := index[term]
			if len(list) == 0 || list[len(list)-1] != id{
				index[term] = append(index[term], id)
			}
		}
	}
	return index
}

//fetches from rss feeds and stores in db
func fetcher(db *sql.DB){
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(WEBSITE_LIST[0])
		if err != nil {
			log.Fatal(err)
	}

	fmt.Printf("Feed Title: %s\n", feed.Title)
		
	for _, item := range feed.Items{
		_, err = db.Exec("INSERT OR IGNORE INTO docs(title, body, url) VALUES(?, ?, ?)", item.Title, item.Description, item.Link)
		if err != nil {
			fmt.Println("error inserting into table")
			log.Fatal(err)
		}

		fmt.Printf("Inserted Article %s\nLink: %s\n", item.Title, item.Link)
	}
}

func main() {
	db, err := sql.Open("sqlite3", "./docs.db")
	if err != nil{
			log.Fatal(err)
	}
	defer db.Close()
	fetcher(db)
	buildIndex(db)
}
