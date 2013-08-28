package page_db

import (
	"sort"
	"math"
	"bytes"
	"strconv"
	"fmt"
	_ "github.com/bmizerany/pq"
	"database/sql"
	"github.com/matus-tomlein/news_processing/environment"
)

type RankedLink struct {
	Url string
	Rank int
	Fontsize string
	UpdateId int
}

type RankedLinks []*RankedLink

func (s RankedLinks) Len() (int) { return len(s) }
func (s RankedLinks) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByRankAndFontsize struct{ RankedLinks }

func (s ByRankAndFontsize) Less(i, j int) (bool) {
	if s.RankedLinks[i].Rank != s.RankedLinks[j].Rank {
		return s.RankedLinks[i].Rank > s.RankedLinks[j].Rank
	}
	return s.RankedLinks[i].Fontsize > s.RankedLinks[j].Fontsize
}

func (ld *LinkDensity) RankLinksAndSaveToDb(updates []int, pageId int, envType string) {
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in RankLinksAndSaveToDb", r)
        }
    }()

	rankedLinks := make(RankedLinks, 0)

	// Create postgres connection
	postgres, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer postgres.Close()

	// Create sqlite connection
	sqlite, err := GetDatabase(pageId, envType)
	if err != nil { panic(err) }
	defer sqlite.Close()

	var buffer bytes.Buffer
	buffer.WriteString("select url, top, left, height, width, fontsize, update_id from links where update_id in (")
	for i, updateId := range updates {
		if i != 0 { buffer.WriteString(", ") }
		buffer.WriteString(strconv.Itoa(updateId))
	}
	buffer.WriteString(")")

	rows, err := sqlite.Query(buffer.String())
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var top, left, height, width float64
		var url, fontsize string
		var updateId int
		err = rows.Scan(&url, &top, &left, &height, &width, &fontsize, &updateId)
		if err != nil { panic(err) }
		top, left, height, width = math.Ceil(top / 10.0),
			math.Ceil(left / 10.0),
			math.Ceil(height / 10.0),
			math.Ceil(width / 10.0)
		maxRank := 0

		// Find the rank for the url as the maximum rank value in its space
		for w := 0.0; w < width; w++ {
			for h := 0.0; h < height; h++ {
				key := fmt.Sprint(left + w, "-", top + h)

				if _, ok := ld.Items[key]; ok {
					if ld.Items[key] > maxRank { maxRank = ld.Items[key] }
				}
			}
		}

		rankedLinks = append(rankedLinks, &RankedLink { url, maxRank, fontsize, updateId })
	}

	if len(rankedLinks) == 0 { return }

	// Create postgres transaction for inserting links
	tx, err := postgres.Begin()
	if err != nil {
		panic(err)
	}
	insertStmt, err := tx.Prepare("insert into links_to_downloads (url, rank, update_id, created_at, updated_at) values ($1, $2, $3, now(), now())")
	if err != nil {
		panic(err)
	}
	defer insertStmt.Close()

	// Save sorted links to db
	sort.Sort(ByRankAndFontsize{rankedLinks})
	for i, link := range rankedLinks {
		_, err = insertStmt.Exec(link.Url, i + 1, link.UpdateId)
		if err != nil { panic(err) }
	}

	tx.Commit()
}