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

//Results of the query search and needed data to return back to app
type Results struct{
	title string
	link string
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
func buildIndex(db *sql.DB) {
    rows, err := db.Query("SELECT id, body FROM docs")
    if err != nil {
        log.Fatal("Query error:", err)
    }
    
    type doc struct {
        id   int
        body string
    }
    var batch []doc

    for rows.Next() {
        var d doc
        if err := rows.Scan(&d.id, &d.body); err != nil {
            log.Println("Scan error:", err)
            continue
        }
        batch = append(batch, d)
    }
    rows.Close() 

    tx, err := db.Begin()
    if err != nil {
        log.Fatal("Transaction begin error:", err)
    }

    stmt, err := tx.Prepare("INSERT OR IGNORE INTO postings (term, doc_id) VALUES (?, ?)")
    if err != nil {
        log.Fatal("Prepare statement error:", err)
    }
    defer stmt.Close()

    for _, d := range batch {
        // Strip HTML and get words
        plainText := stripHTML(d.body)
        terms := tokenize(plainText)

        for _, term := range terms {
            if term == "" { continue }
            _, err := stmt.Exec(term, d.id)
            if err != nil {
                log.Printf("Insert error for term [%s]: %v", term, err)
            }
        }
    }

    // 5. Finalize
    if err := tx.Commit(); err != nil {
        log.Fatal("Commit error:", err)
    }
    log.Println("Successfully built index in 'postings' table!")
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

func loadDictionary(db *sql.DB) Dictionary {
    dict := make(Dictionary)
    // Only select two columns since we aren't calculating tf yet
    rows, err := db.Query("SELECT term, COUNT(doc_id) FROM postings GROUP BY term")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var term string
        var df int
        // Removed tf from scan to match the SELECT statement
        if err := rows.Scan(&term, &df); err != nil {
            log.Fatal(err)
        }
        dict[term] = &Term{doc_freq: df, term_freq: 0}
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

func search(db *sql.DB, query string, dict Dictionary) []Results{
    terms := tokenize(strings.ToLower(query))
    if len(terms) == 0 {
        return nil
    }
		fmt.Printf("Search terms: %v\n", terms)

    sort.Slice(terms, func(i, j int) bool {
        // Handle cases where term isn't in dict to avoid nil pointer
        dfI, dfJ := 0, 0
        if t, ok := dict[terms[i]]; ok { dfI = t.doc_freq }
        if t, ok := dict[terms[j]]; ok { dfJ = t.doc_freq }
        return dfI < dfJ
    })

    var results []int

    for i, term := range terms {
        if _, exists := dict[term]; !exists {
						fmt.Printf("Term [%s] not found in Dictionary\n", term)
            return nil 
        }

        rows, err := db.Query("SELECT doc_id FROM postings WHERE term = ? ORDER BY doc_id ASC", term)
        if err != nil {
            continue
        }

        var currentList []int
        for rows.Next() {
            var id int
            rows.Scan(&id)
            currentList = append(currentList, id)
        }
        rows.Close()

        if i == 0 {
            results = currentList
        } else {
            results = intersect(results, currentList)
        }

        if len(results) == 0 {
            break
        }
    }

    data := []Results{}
    for _, id := range results {
        var title string
				var link string
        db.QueryRow("SELECT title, url FROM docs where id = ?", id).Scan(&title, &link)
				var d Results
				d.link = link
				d.title = title
        data = append(data, d)
    }
		fmt.Printf("Found %d DocIDs\n", len(results))
    return data
}


func main() {
	db, err := sql.Open("sqlite3", "./docs.db")
	if err != nil{
			log.Fatal(err)
	}
	defer db.Close()
	dict := loadDictionary(db)
	//fetcher(db)
	buildIndex(db)
	listen(db, dict)
}
