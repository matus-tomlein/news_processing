package link_search

import (
	"fmt"
	"github.com/matus-tomlein/news_processing/page_db"
	"github.com/matus-tomlein/news_processing/helpers"
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

func getLinksFromPageUpdates(pageId int, pageUrl, updates string, updatesCreatedAt map[int]time.Time, searchResults []*SearchResult, envType string) ([]*SearchResult) {
	db, err := page_db.GetDatabase(pageId, envType)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select url, update_id, inner_text from links where update_id in (" + updates + ")")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	return getSearchResults(rows, updatesCreatedAt, searchResults, pageId, pageUrl)

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

	return getSearchResults(rows, nil, searchResults, pageId, pageUrl)
}

func getSearchResults(rows *sql.Rows, updatesCreatedAt map[int]time.Time, searchResults []*SearchResult, pageId int, pageUrl string) ([]*SearchResult) {
	for rows.Next() {
		var url, title string
		var updateId int

		err := rows.Scan(&url, &updateId, &title)
		if err != nil { panic(err) }

		searchResult := &SearchResult {
			LinkId: helpers.Hash(url),
			Title: title,
			Url: url,
			UpdateId: updateId,
			PageId: pageId,
			PageUrl: pageUrl,
		}
		if updatesCreatedAt != nil {
			searchResult.CreatedAt = updatesCreatedAt[updateId]
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

func FindLatestLinks() {
	envType := environment.Current()

	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer db.Close()

	searchResults := make([]*SearchResult, 0)
	pageUpdates := make(map[int]string)
	pageUrls := make(map[int]string)
	updatesCreatedAt := make(map[int]time.Time)

	// Select planned updates to execute
	rows, err := db.Query(`SELECT updates.id AS update_id, pages.id AS page_id, pages.url AS url, updates.created_at FROM updates JOIN pages ON pages.id = updates.page_id AND pages.priority >= 5 WHERE updates.created_at > now() - '3 days'::interval AND parsed = TRUE ORDER BY page_id;`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var updateId, pageId int
		var url string
		var createdAt time.Time

		err = rows.Scan(&updateId, &pageId, &url, &createdAt)
		if err != nil {
			panic(err)
		}

		if _, ok := pageUpdates[pageId]; ok {
			pageUpdates[pageId] = strings.Join([]string{ pageUpdates[pageId], strconv.Itoa(updateId) }, ",")
		} else {
			pageUpdates[pageId] = strconv.Itoa(updateId)
		}
		pageUrls[pageId] = url
		updatesCreatedAt[updateId] = createdAt
	}

	for pageId, updates := range pageUpdates {
		searchResults = getLinksFromPageUpdates(pageId, pageUrls[pageId], updates, updatesCreatedAt, searchResults, envType)
	}

	outputResults(searchResults)
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
	outputResults(searchResults)
}

func outputResults(searchResults []*SearchResult) {
	b, err := json.Marshal(searchResults)
	if err != nil { panic(err) }

	err = ioutil.WriteFile("output.json", b, 0644)
	if err != nil { panic(err) }
}
