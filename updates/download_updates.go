package main

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/matus-tomlein/news_processing/environment"
	"github.com/matus-tomlein/news_processing/helpers"
	"time"
	"io/ioutil"
	_ "github.com/bmizerany/pq"
	"database/sql"
	"strings"
	"errors"
	"bytes"
)

func downloadJson(pageId int, url string, envType string) ([]byte, error) {
	done := make(chan error)
	cmd := exec.Command("phantomjs", fmt.Sprintf("%s/updates/links.js", environment.AppPath(envType)), url)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Start()
	var out string

	go func() {
		done <- func() (error) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(stdout)
			out = buf.String()
			err = cmd.Wait()
			return err
		}()
	}()
	select {
	case <-time.After(90 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			panic(err)
		}
		<-done // allow goroutine to exit
		return make([]byte, 0), errors.New("Process timeout, killed")
	case err := <-done:
		outString := string(out)
		i := strings.Index(outString, "[{")
		if i < 0 {
			return make([]byte, 0), errors.New("Failed to process links JSON")
		}
		return []byte(out[i:len(out)]), err
	}
	return make([]byte, 0), errors.New("Unknown error")
}

func download(pageId int, url string, envType string) (string, error) {
	out, err := downloadJson(pageId, url, envType)
	if err != nil { return "", err }

	t := time.Now()
	cacheFolderName := fmt.Sprintf("%d/go.%s", pageId, t.Format("20060102150405"))
	fileName := fmt.Sprintf("%s/parsed/%s.json", environment.CachePath(envType), cacheFolderName)

	err = os.MkdirAll(fmt.Sprintf("%s/parsed/%d", environment.CachePath(envType), pageId), 0777)
	if err != nil { panic(err) }
	err = ioutil.WriteFile(fileName, out, 0777)
	if err != nil { panic(err) }

	return cacheFolderName, nil
}

func updateIntervalForPagePriority(pagePriority int) int {
	if pagePriority == 10 {
		return 3
	} else if pagePriority == 5 {
		return 6
	} else if pagePriority == 3 {
		return 12
	}
	return 24
}

func main() {
	envType := environment.Current()

	messages := make(chan string)
	go helpers.ReadInput(messages)

	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer db.Close()

	for {
		// Select planned updates to execute
		rows, err := db.Query(`select planned_updates.id, page_id, pages.url, pages.priority, num_failed_accesses
			from planned_updates join pages on pages.id = planned_updates.page_id and pages.track = TRUE
			where execute_after < now() limit 10`)
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
			fmt.Println("Downloading update for page", pageId)
			cacheFolderName, err := download(pageId, url, envType)

			if err == nil { // if successful
				// insert into updates
				_, err = db.Exec(`insert into updates (page_id, cache_folder_name, parsed, created_at, updated_at)
					values ($1, $2, TRUE, now(), now())`,
					pageId, cacheFolderName)
				if err != nil { panic(err) }

				// reset num_failed_accesses in pages if it is > 0
				if numFailedAccesses > 0 {
					_, err = db.Exec("update pages set num_failed_accesses = 0 where id = $1",
						pageId)
					if err != nil { panic(err) }
				}
				fmt.Println("Update downloaded for page", pageId, "in folder", cacheFolderName)
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
			_, err = db.Exec(fmt.Sprintf("insert into planned_updates (execute_after, page_id, created_at, updated_at) values (now() + interval '%d hour', $1, now(), now())", updateIntervalForPagePriority(pagePriority)),
				pageId)
			if err != nil { panic(err) }
		}


		fmt.Println("Planned updates executed")
		time.Sleep(60 * time.Second)
		select {
			case msg := <-messages:
				if msg == "q" { os.Exit(0) }
			default:
		}
	}
}
