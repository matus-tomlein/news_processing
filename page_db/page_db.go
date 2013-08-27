package page_db

import (
	"database/sql"
	"strings"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"io/ioutil"
    "encoding/json"
	"crypto/md5"
	"github.com/matus-tomlein/news_processing/helpers"
	"github.com/matus-tomlein/news_processing/environment"
)

type UpdateLink struct {
	Text string
	Url string
	Images []string
	Fontstyle string
	Fontsize string
	Bottom float32
	Top float32
	Left float32
	Right float32
	Height float32
	Width float32
}

type UpdateRoot struct {
	PageId int `json:"page_id"`
	Url string
	Links []UpdateLink
}

func GetDatabase(pageId int, envType string) (*sql.DB, error) {
	filename := environment.PageDbPath(700, envType)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		helpers.CopyFile(filename, environment.PageDbTemplatePath(envType))
		fmt.Printf("Created database: %s", filename)
	}

	return sql.Open("sqlite3", filename)
}

func CreateDatabase(pageId int, updates []int, envType string) {
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in CreateDatabase", r)
        }
    }()

	db, err := GetDatabase(pageId, envType)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	pageDomain := ""

	var ads AdsFiltering
	ads.Init(envType)

	for _, updateId := range updates {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in update", r)
			}
		}()

		filename := environment.UpdateJsonPath(updateId, envType)
		if _, err := os.Stat(filename); os.IsNotExist(err) { // Check if exists
			fmt.Printf("File does not exist: %s", filename)
			continue
		}
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println(err)
			continue
		}
		root := UpdateRoot{}
		err = json.Unmarshal(content, &root)
		if err != nil {
			panic(err)
		}
		if root.PageId != pageId { // Check if the update is the same page as requested in func arguments
			continue
		}
		if pageDomain == "" {
			pageDomain = helpers.GetDomain(root.Url)
		}

		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		stmt, err := tx.Prepare("insert into links (url, images, fontstyle, fontsize, top, left, height, width, inner_text, update_id, file_name, same_domain) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		selectStmt, err := tx.Prepare("select 1 from links where url = ?")
		if err != nil {
			panic(err)
		}
		defer selectStmt.Close()

		for _, link := range root.Links {
			if pageDomain != helpers.GetDomain(link.Url) {
				continue
			}

			// Check if url exists in db
			rows, err := selectStmt.Query(link.Url)
			if err != nil {
				panic(err)
			}
			if rows.Next() {
				continue
			}

			// Check if the link is an ad
			if ads.MatchUrl(link.Url) {
				continue
			}

			// Compute hash of url
			running_hash := md5.New();
			urlHash := running_hash.Sum([]byte(link.Url));

			// Save to db
			_, err = stmt.Exec(link.Url, strings.Join(link.Images, "\n"), link.Fontstyle, link.Fontsize,
				link.Top, link.Left, link.Height, link.Width, link.Text, updateId, urlHash, 1)
			if err != nil {
				panic(err)
			}
		}

		tx.Commit()
		fmt.Println("Processed update", updateId)
	}
}