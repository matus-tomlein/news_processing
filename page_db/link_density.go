package page_db

import (
	"encoding/json"
	"os"
	"math"
	"strconv"
	"fmt"
	"bytes"
	"github.com/matus-tomlein/news_processing/environment"
	"io/ioutil"
)

type LinkDensity struct {
	Items map[string] int
	Updates []int
}

func NewLinkDensity(pageId int, envType string) *LinkDensity {
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in NewLinkDensity", r)
        }
    }()

	filename := environment.LinkDensityPath(pageId, envType)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		ld := &LinkDensity { Items: make(map[string]int) }
		return ld
	}
	b, err := ioutil.ReadFile(filename)
    if err != nil { panic(err) }
	ld := &LinkDensity {}
	err = json.Unmarshal(b, &ld)
    if err != nil { panic(err) }
	return ld
}

func (ld *LinkDensity) Update(pageId int, allUpdates []int, envType string) (updates []int) {
	defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered in Update", r)
        }
    }()

	updates = make([]int, 0)
	for _, updateId := range allUpdates {
		if (!contains(ld.Updates, updateId)) {
			updates = append(updates, updateId)
			ld.Updates = append(ld.Updates, updateId)
		} else {
			fmt.Println("Skipped update", updateId)
		}
	}

	db, err := GetDatabase(pageId, envType)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var buffer bytes.Buffer
	buffer.WriteString("select top, left, height, width from links where update_id in (")
	for i, updateId := range updates {
		if i != 0 { buffer.WriteString(", ") }
		buffer.WriteString(strconv.Itoa(updateId))
	}
	buffer.WriteString(")")

	rows, err := db.Query(buffer.String())
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var top, left, height, width float64
		rows.Scan(&top, &left, &height, &width)
		top, left, height, width = math.Ceil(top / 10.0),
			math.Ceil(left / 10.0),
			math.Ceil(height / 10.0),
			math.Ceil(width / 10.0)

		for w := 0.0; w < width; w++ {
			for h := 0.0; h < height; h++ {
				key := fmt.Sprint(left + w, "-", top + h)

				if _, ok := ld.Items[key]; ok {
					ld.Items[key]++
				} else {
					ld.Items[key] = 1
				}
			}
		}
	}
	ld.save(pageId, envType)
	return
}

func (ld *LinkDensity) save(pageId int, envType string) {
	b, err := json.Marshal(ld)
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = ioutil.WriteFile(environment.LinkDensityPath(pageId, envType), b, 0644)
	if err != nil { panic(err) }
}

func contains(s []int, e int) bool {
    for _, a := range s { if a == e { return true } }
    return false
}