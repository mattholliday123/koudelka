package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
)

var WEBSITE_LIST = [200]string{"https://en.wikipedia.org/w/index.php?title=Special:NewPages&feed=rss",
"https://en.wikipedia.org/w/api.php?hidebots=1&hidecategorization=1&hideWikibase=1&urlversion=1&days=1&limit=50&action=feedrecentchanges&feedformat=rss", "https://www.theguardian.com/us/rss","https://www.espn.com/espn/rss/news", "https://feeds.bbci.co.uk/news/rss.xml", "http://rss.cnn.com/rss/cnn_topstories.rss"}

//term struct for our Dictionary
type Term struct {
	doc_freq int
	term_freq int
}

//Dictionary
type Dictionary map[string]*Term


func stripHTML(input string) string {
    // Basic regex to remove everything inside < >
    re := regexp.MustCompile(`<(.|\n)*?>`)
    return re.ReplaceAllString(input, "")
}

func tokenize(input string) []string {
    // Convert to lower and split by non-alphanumeric characters
    re := regexp.MustCompile(`[a-zA-Z0-9]+`)
    return re.FindAllString(strings.ToLower(input), -1)
}

//Build the postings list onto disk
func buildIndex(db *sql.DB){
	rows, err := db.Query("SELECT id, body FROM docs")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	tx, err := db.Begin()
    if err != nil {
        log.Fatal(err)
    }

    stmt, _ := tx.Prepare("INSERT OR IGNORE INTO postings (term, doc_id) VALUES (?, ?)")
	for rows.Next(){
		var id int
		var body string
		rows.Scan(&id, &body)
		plainText := stripHTML(body)
		terms := tokenize(plainText)

		for _, term := range terms {
			_, err := stmt.Exec(term,id)
			if err != nil{
				log.Printf("Error inserting into Postings List: %v", err)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
}
	fmt.Printf("Index built\n")
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

//load our dictionary into memory
func loadDictionary(db *sql.DB) Dictionary {

	dict := make(Dictionary)
	rows, _ := db.Query("SELECT term, COUNT(doc_id) FROM postings GROUP BY term")
	defer rows.Close()

	for rows.Next(){
		var term string
		var df int
		var tf int
		rows.Scan(&term, &df, &tf)
		dict[term] = &Term{doc_freq: df, term_freq: tf}
	}
	return dict
}

func intersect(p1, p2 []int) []int {
    combined := make([]int, 0)
    i, j := 0, 0

    for i < len(p1) && j < len(p2) {
        if p1[i] == p2[j] {
            combined = append(combined, p1[i])
            i++
            j++
        } else if p1[i] < p2[j] {
            i++
        } else {
            j++
        }
    }
    return combined
}

func search(db *sql.DB,query string, dict Dictionary) []int{

    //Look up the term in postings map
		rows, err := db.Query("SELECT doc_id FROM postings WHERE query = ? ORDER BY doc_id ASC", query)
		if err != nil{
			log.Fatal(err)
			return nil
		}
		defer rows.Close()

		//list of doc ids sorted
		var currentList []int
		for rows.Next(){
			var id int
			rows.Scan(&id)
			currentList = append(currentList, id)
		}
    

    terms := tokenize(strings.ToLower(query))
		if len(terms) == 0 {
			return nil 
		}

		sort.Slice(terms, func(i, j int) bool {
			return dict[terms[i]].doc_freq < dict[terms[j]].doc_freq
		})

  var results []int
    for i, term := range terms {
        // Check if term even exists in our dict
        if _, exists := dict[term]; !exists {
            return nil
        }
        if i == 0 {
            results = currentList
        } else {
            results = intersect(results, currentList)
        }

        //if results become empty, stop early
        if len(results) == 0 {
            break
        }
    }
    return results
}

func main() {
	db, err := sql.Open("sqlite3", "./docs.db")
	if err != nil{
			log.Fatal(err)
	}
	defer db.Close()
	fetcher(db)
	buildIndex(db)
	listen(db)
}
