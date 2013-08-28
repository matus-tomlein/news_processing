package main

import (
	"fmt"
	"github.com/matus-tomlein/news_processing/page_db"
)

func readInput(messages chan string) {
	fmt.Println("Type q to quit")
	for {
		var input string
		_, err := fmt.Scanf("%s", &input)
		if err != nil { panic(err) }
		messages <- input
	}
}

func main() {
	ads := &page_db.AdsFiltering{}
	ads.Init("test")
	page_db.CreateOrUpdateDatabase(700, "http://www.boston.com", []page_db.UpdateInfo {
		{1236, "1371183281-p"},
	}, ads, "test")
}

func main1() {
	messages := make(chan string)
	go readInput(messages)

	envType := "test"
	processLinks(envType, messages)
}