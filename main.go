package main

import (
	"github.com/matus-tomlein/news_processing/page_db"
)

func main() {
	envType = "test"
	ld := page_db.NewLinkDensity(700, envType)
	page_db.CreateDatabase(700, []int{717, 10995, 105185, 109233, 109332, 111329}, envType)
	ld.Update(700, []int{717, 10995, 105185, 109233, 109332, 111329}, envType)
}