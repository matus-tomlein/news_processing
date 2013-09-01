package main

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/matus-tomlein/news_processing/environment"
	"github.com/matus-tomlein/news_processing/helpers"
	"github.com/matus-tomlein/news_processing/page_db"
	"time"
	"encoding/json"
	"io/ioutil"
	_ "github.com/bmizerany/pq"
	"database/sql"
)

func downloadJson(pageId int, url string, envType string) ([]byte, error) {
	out, err := exec.Command("phantomjs", fmt.Sprintf("%s/updates/links.js", environment.AppPath(envType)), url).Output()

	if err != nil {
		fmt.Println(err)
		return make([]byte, 0), err
	}
	links := make([]page_db.UpdateLink, 0)
	err = json.Unmarshal(out, &links)
	if err != nil {
		fmt.Println(err)
		return make([]byte, 0), err
	}
	return out, err
}

func download(pageId int, url string, envType string) (string, error) {
	out, err := downloadJson(pageId, url, envType)
	if err != nil { return "", err }

	t := time.Now()
	cacheFolderName := fmt.Sprintf("%d/go.%s", pageId, t.Format("20060102150405"))
	fileName := fmt.Sprintf("%s/parsed/%s.json", environment.CachePath(envType), cacheFolderName)

	err = os.Mkdir(fmt.Sprintf("%s/parsed/%d", environment.CachePath(envType), pageId), 0777)
	if err != nil { panic(err) }
	err = ioutil.WriteFile(fileName, out, 0777)
	if err != nil { panic(err) }

	return cacheFolderName, nil
}

func updateIntervalForPagePriority(pagePriority int) int {
	if pagePriority == 10 {
		return 3
	} else if pagePriority == 5 {
		return 12
	}
	return 24
}

func main() {
	envType := "test"

	messages := make(chan string)
	go helpers.ReadInput(messages)

	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer db.Close()

	for {
		rows, err := db.Query("select id, page_id, pages.url, pages.priority, num_failed_accesses from planned_updates where execute_after < now() join pages on pages.id = planned_updates.page_id and pages.track = TRUE")
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			var plannedUpdateId, pageId, pagePriority, numFailedAccesses int
			var url string

			err = rows.Scan(&plannedUpdateId, &pageId, &url, &pagePriority, &numFailedAccesses)
			if err != nil { panic(err) }

			// download the update
			cacheFolderName, err := download(pageId, url, envType)

			if err == nil { // if successful
				// insert into updates
				_, err = db.Exec("insert into updates (page_id, cache_folder_name, parsed, created_at, updated_at) values (?, ?, TRUE, now(), now())",
					pageId, cacheFolderName)
				if err != nil { panic(err) }

				// reset num_failed_accesses in pages if it is > 0
				if numFailedAccesses > 0 {
					_, err = db.Exec("update pages set num_failed_accesses = 0 where id = $1",
						pageId)
					if err != nil { panic(err) }
				}
			} else { // if failed to download update, increase num_failed_accesses in pages
				numFailedAccesses++
				_, err = db.Exec("update pages set num_failed_accesses = $1 where id = $2",
					numFailedAccesses, pageId)
				if err != nil { panic(err) }
			}

			// delete planned_update
			_, err = db.Exec("delete from planned_updates where id = $1", plannedUpdateId)
			if err != nil { panic(err) }

			// create new planned_update
			_, err = db.Exec(fmt.Sprintf("insert into planned_updates (execute_after, page_id, created_at, updated_at) values (now() + interval '%d hour', ?, now(), now())", updateIntervalForPagePriority(pagePriority)),
				pageId)
			if err != nil { panic(err) }
		}


		fmt.Println("Planned updates executed")
		time.Sleep(120 * time.Second)
		select {
			case msg := <-messages:
				if msg == "q" { os.Exit(0) }
			default:
		}
	}
}