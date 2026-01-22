package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"regexp"
	"slices"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmcdole/gofeed"
)

var WEBSITE_LIST = [6]string{"https://en.wikipedia.org/w/index.php?title=Special:NewPages&feed=rss",
"https://en.wikipedia.org/w/api.php?hidebots=1&hidecategorization=1&hideWikibase=1&urlversion=1&days=1&limit=50&action=feedrecentchanges&feedformat=rss", "https://www.theguardian.com/us/rss","https://www.espn.com/espn/rss/news", "https://feeds.bbci.co.uk/news/rss.xml", "http://rss.cnn.com/rss/cnn_topstories.rss"}


//term struct for our Dictionary
//For BM25 ranking - term frequency, Inverse Document Frequency(IDF)
type Term struct {
	doc_freq int
	term_freq int
}

type Rank struct {
	docID int
	score float64
}

//Results of the query search and needed data to return back to app
type Results struct{
	Title string `json:"title"`
	Link string `json:"link"`
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

    stmt, err := tx.Prepare("INSERT OR IGNORE INTO postings (term, doc_id, tf, doc_len) VALUES (?, ?, ?, ?)")
    if err != nil {
        log.Fatal("Prepare statement error:", err)
    }
    defer stmt.Close()

    for _, d := range batch {
        // Strip HTML and get words
        plainText := stripHTML(d.body)
        terms := tokenize(plainText)
				doc_len := len(terms)

				term_count := make(map[string]int)
				for _, term := range terms {
					term_count[term]++
				}

        for term, tf := range term_count{
            _, err := stmt.Exec(term, d.id, tf, doc_len)
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
	var feed *gofeed.Feed

	for _,web := range WEBSITE_LIST{
		var err error
		feed, err = fp.ParseURL(web)
		if err != nil {
			log.Fatal(err)
	}
		fmt.Printf("Feed Title: %s\n", feed.Title)
			for _, item := range feed.Items{
		_, err := db.Exec("INSERT OR IGNORE INTO docs(title, body, url) VALUES(?, ?, ?)", item.Title, item.Description, item.Link)
		if err != nil {
			fmt.Println("error inserting into table")
			log.Fatal(err)
		}

		fmt.Printf("Inserted Article %s\nLink: %s\n", item.Title, item.Link)
	}
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



func search(db *sql.DB, query string, dict Dictionary) []Results {
	terms := tokenize(strings.ToLower(query))
	if len(terms) == 0 {
		return nil
	}

	fmt.Printf("Search terms: %v\n", terms)

	scores := make(map[int]float64)
	k1 := 1.5
	b := 0.75

	var total_docs int
	var avg_len float64
	db.QueryRow("SELECT COUNT(*) FROM docs").Scan(&total_docs)
	db.QueryRow("SELECT AVG(doc_len) FROM postings").Scan(&avg_len)

	for _, term := range terms {
		tData, exists := dict[term]
		if !exists {
			continue
		}

		df := float64(tData.doc_freq)
		idf := math.Log((float64(total_docs)-df+0.5)/(df+0.5) + 1.0)

		rows, err := db.Query("SELECT doc_id, tf, doc_len FROM postings WHERE term = ?", term)
		if err != nil {
			continue
		}

		for rows.Next() {
			var id int
			var tf int
			var doc_len int
			rows.Scan(&id, &tf, &doc_len)
			
			tf_f := float64(tf)
			dl_f := float64(doc_len)

			numerator := tf_f * (k1 + 1)
			denominator := tf_f + k1*(1-b+b*(dl_f/avg_len))

			scores[id] += idf * (numerator / denominator)
		}
		rows.Close()
	}

	type Rank struct {
		DocID int
		Score float64
	}
	var rankedResults []Rank
	for id, score := range scores {
		rankedResults = append(rankedResults, Rank{DocID: id, Score: score})
	}

	slices.SortFunc(rankedResults, func(a, b Rank) int {
		if b.Score > a.Score {
			return 1
		} else if b.Score < a.Score {
			return -1
		}
		return 0
	})

	if len(rankedResults) > 20 {
		rankedResults = rankedResults[:20]
	}

	var finalData []Results
	for _, res := range rankedResults {
		var title, link string
		err := db.QueryRow("SELECT title, url FROM docs WHERE id = ?", res.DocID).Scan(&title, &link)
		if err != nil {
			continue
		}
		finalData = append(finalData, Results{
			Title: title,
			Link:  link,
		})
		fmt.Printf("Found: %s\n", title)
	}

	fmt.Printf("Found %d DocIDs\n", len(finalData))
	return finalData
} 


func main() {
	db, err := sql.Open("sqlite3", "./docs.db")
	if err != nil{
			log.Fatal(err)
	}
	defer db.Close()
	dict := loadDictionary(db)
	fetcher(db)
	buildIndex(db)
	listen(db, dict)
}
