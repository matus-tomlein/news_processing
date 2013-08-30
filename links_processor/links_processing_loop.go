package main

import (
	"strconv"
	"fmt"
	"strings"
	"os"
	"github.com/matus-tomlein/news_processing/page_db"
	"time"
	"io/ioutil"
	_ "github.com/bmizerany/pq"
	"database/sql"
	"github.com/matus-tomlein/news_processing/environment"
)

func processLinksInUpdates(pageId int, updates []int, ads *page_db.AdsFiltering, envType string) {
	fmt.Println("Processing page", pageId, "with", len(updates), "updates")
	page_db.CreateOrUpdateDatabase(pageId, updates, ads, envType)
	ld := page_db.NewLinkDensity(pageId, envType)
	newUpdates := ld.Update(pageId, updates, envType)
	ld.RankLinksAndSaveToDb(newUpdates, pageId, envType)
}

func processUpdatesWithLinksProcessingId(linksProcessingId int, db *sql.DB, ads *page_db.AdsFiltering, envType string) (int) {
	currentLinksProcessingIdPath := environment.CurrentLinksProcessingIdPath(envType)
	pageUpdates := make(map[int][]int)
	updatesFound := false

	rows, err := db.Query("select id, page_id from updates where links_processing_id = $1", linksProcessingId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		updatesFound = true
		var updateId, updatePageId int

		err = rows.Scan(&updateId, &updatePageId)
		if err != nil { panic(err) }

		if updateIds, ok := pageUpdates[updatePageId]; ok {
			pageUpdates[updatePageId] = append(updateIds, updateId)
		} else {
			pageUpdates[updatePageId] = []int { updateId }
		}
	}

	if updatesFound {
		// Create a temporary file to mark the ongoing process in case of crash
		err = ioutil.WriteFile(currentLinksProcessingIdPath, []byte(strconv.Itoa(linksProcessingId)), 0644)
		if err != nil { panic(err) }

		// Process updates for each page
		for pageId, updateIds := range pageUpdates {
			processLinksInUpdates(pageId, updateIds, ads, envType)
		}

		// Remove the temporary file
		err = os.Remove(currentLinksProcessingIdPath)
		if err != nil { panic(err) }

		linksProcessingId++
	}

	return linksProcessingId
}

func recoverUnfinishedLinksProcessing(db *sql.DB, ads *page_db.AdsFiltering, envType string) {
	b, err := ioutil.ReadFile(environment.CurrentLinksProcessingIdPath(envType))
	if err != nil { panic(err) }
	linksProcessingId, err := strconv.Atoi(string(b))
	if err != nil { panic(err) }
	processUpdatesWithLinksProcessingId(linksProcessingId, db, ads, envType)
}

func processLinks(envType string, messages chan string) {
	currentLinksProcessingIdPath := environment.CurrentLinksProcessingIdPath(envType)
	ads := &page_db.AdsFiltering{}
	ads.Init(envType)

	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer db.Close()

	if _, err := os.Stat(currentLinksProcessingIdPath); err == nil {
		fmt.Printf("Recovering unfinished processing")
	    recoverUnfinishedLinksProcessing(db, ads, envType)
	}

	pageIds := make([]string, 0)
	rows, err := db.Query("select id from pages where priority >= 5")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var pageId int
		err = rows.Scan(&pageId)
		if err != nil { panic(err) }
		pageIds = append(pageIds, strconv.Itoa(pageId))
	}
	if len(pageIds) == 0 {
		panic("No pages to process")
	}
	pageIdsString := strings.Join(pageIds, ",")

	linksProcessingId := 0
	rows, err = db.Query("select COALESCE(max(links_processing_id), 0) AS links_processing_id from updates where links_processing_id IS NOT NULL")
	if err != nil { panic(err) }
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&linksProcessingId)
		if err != nil { panic(err) }
		linksProcessingId++
	}

	for {
		_, err = db.Exec("update updates set links_processing_id = $1 where id in (select id from updates where page_id in (" +
			pageIdsString + ") and parsed = TRUE and links_processing_id IS NULL order by id limit 100)",
			linksProcessingId)
		if err != nil { panic(err) }

		linksProcessingId = processUpdatesWithLinksProcessingId(linksProcessingId, db, ads, envType)

		select {
			case msg := <-messages:
				if msg == "q" { os.Exit(0) }
			default:
		}

		fmt.Println("All pages processed")
		time.Sleep(120 * time.Second)

		select {
			case msg := <-messages:
				if msg == "q" { os.Exit(0) }
			default:
		}
	}
}
