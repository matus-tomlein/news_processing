package link_search

import (
	"fmt"
	"github.com/matus-tomlein/news_processing/page_db"
	"github.com/matus-tomlein/news_processing/environment"
	_ "github.com/bmizerany/pq"
	"database/sql"
	"time"
	"strconv"
	"strings"
	"encoding/json"
	"io/ioutil"
)

type SearchResult struct {
	LinkId uint32
	Title string
	Url string
	UpdateId int
	PageId int
	PageUrl string
	CreatedAt time.Time
}

func SearchPageDb(pageId int, query, pageUrl, envType string) ([]*SearchResult) {
	db, err := page_db.GetDatabase(pageId, envType)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	searchResults := make([]*SearchResult, 0)
	rows, err := db.Query(fmt.Sprintf("select url, update_id, inner_text from links where inner_text like '%%%s%%'", query))
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var url, title string
		var updateId int

		err = rows.Scan(&url, &updateId, &title)
		if err != nil { panic(err) }

		searchResult := &SearchResult {
			LinkId: helpers.Hash(url),
			Title: title,
			Url: url,
			UpdateId: updateId,
			PageId: pageId,
			PageUrl: pageUrl,
		}
		searchResults = append(searchResults, searchResult)
	}
	return searchResults
}

func SearchAllDbs(query, envType string) ([]*SearchResult) {
	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer db.Close()

	searchResults := make([]*SearchResult, 0)

	// Select planned updates to execute
	rows, err := db.Query(`select id, url from pages where priority > 0`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var url string
		err = rows.Scan(&id, &url)
		if err != nil {
			panic(err)
		}

		fmt.Println("Searching page", url)
		newResults := SearchPageDb(id, query, url, envType)
		for _, result := range newResults {
			searchResults = append(searchResults, result)
		}
	}
	return searchResults

}

func Search(query string) {
	
}

func StartSearch(messages chan string) {
	fmt.Println("Enter a search query:")

	query := <-messages
	fmt.Println(query)
	searchResults := SearchAllDbs(query, environment.Current())

	updateIds := make([]string, 0)
	for _, result := range searchResults {
		updateIds = append(updateIds, strconv.Itoa(result.UpdateId))
	}
	updateDates := make(map[int]time.Time)

	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(environment.Current()))
	if err != nil { panic(err) }
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("select created_at, id from updates where id in (%s)", strings.Join(updateIds, ", ")))
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var createdAt time.Time
		var updateId int
		err = rows.Scan(&createdAt, &updateId)
		if err != nil { panic(err) }
		updateDates[updateId] = createdAt
	}

	fmt.Println(len(searchResults))
	for _, result := range searchResults {
		result.CreatedAt = updateDates[result.UpdateId]
	}

	b, err := json.Marshal(searchResults)
	if err != nil { panic(err) }

	err = ioutil.WriteFile("output.json", b, 0644)
	if err != nil { panic(err) }
}
