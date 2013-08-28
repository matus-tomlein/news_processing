package main

import (
	"strconv"
	"fmt"
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

func processUpdatesWithLinksProcessingId(pageId, linksProcessingId int, db *sql.DB, ads *page_db.AdsFiltering, envType string) (int) {
	currentLinksProcessingIdPath := environment.CurrentLinksProcessingIdPath(envType)
	updateIds := make([]int, 0)

	rows, err := db.Query("select id, page_id from updates where links_processing_id = $1", linksProcessingId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var updateId, updatePageId int
		err = rows.Scan(&updateId, &updatePageId)
		if err != nil { panic(err) }
		updateIds = append(updateIds, updateId)
		if pageId < 0 { pageId = updatePageId }
	}
	if len(updateIds) > 0 {
		err = ioutil.WriteFile(currentLinksProcessingIdPath, []byte(strconv.Itoa(linksProcessingId)), 0644)
		if err != nil { panic(err) }

		processLinksInUpdates(pageId, updateIds, ads, envType)

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
	processUpdatesWithLinksProcessingId(-1, linksProcessingId, db, ads, envType)
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

	pageIds := make([]int, 0)
	rows, err := db.Query("select id from pages where priority >= 5")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var pageId int
		err = rows.Scan(&pageId)
		if err != nil { panic(err) }
		pageIds = append(pageIds, pageId)
	}

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
		for _, pageId := range pageIds {
			_, err = db.Exec("update updates set links_processing_id = $1 where id in (select id from updates where page_id = $2 and parsed = TRUE and links_processing_id IS NULL order by id limit 100)",
				linksProcessingId,
				pageId)
			if err != nil { panic(err) }

			linksProcessingId = processUpdatesWithLinksProcessingId(pageId, linksProcessingId, db, ads, envType)
			time.Sleep(1 * time.Second)

			select {
				case msg := <-messages:
					if msg == "q" { os.Exit(0) }
				default:
			}
		}
		fmt.Println("All pages processed")
		time.Sleep(60 * time.Second)

		select {
			case msg := <-messages:
				if msg == "q" { os.Exit(0) }
			default:
		}
	}
}
