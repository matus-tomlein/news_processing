package page_db;

import (
	"database/sql"
	"github.com/matus-tomlein/news_processing/helpers"
	"github.com/matus-tomlein/news_processing/environment"
	_ "github.com/bmizerany/pq"
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"
)

func DumpLinks(path, envType string) {
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in DumpLinks", r)
        }
    }()

	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}

	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer db.Close()

	// Select planned updates to execute
	rows, err := db.Query(`select id, url from pages`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var pageId int
		var url string
		err = rows.Scan(&pageId, &url)
		if err != nil {
			panic(err)
		}

		fmt.Println("Dumping page", url)
		// Create sqlite connection
		sqlite, err := GetDatabase(pageId, envType)
		if err != nil { panic(err) }
		defer sqlite.Close()

		sqliteRows, err := sqlite.Query("select distinct url from links where inner_text <> ''")
		if err != nil {
			panic(err)
		}
		defer sqliteRows.Close()
		for sqliteRows.Next() {
			var linkUrl string
			err = sqliteRows.Scan(&linkUrl)
			if err != nil { panic(err) }

			linkId := helpers.Hash(linkUrl)

			s := []string{strconv.Itoa(pageId), " ", fmt.Sprintf("%d", linkId), " ", linkUrl, "\n"};

			n, err := io.WriteString(file, strings.Join(s, ""))
			if err != nil {
				fmt.Println(n, err)
			}
		}
	}
	file.Close()
}
