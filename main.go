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
	page_db.CreateOrUpdateDatabase(700, []int { 717, 10995, 105185,   109233, 109332, 111329 }, ads, "test")
}

func main1() {
	messages := make(chan string)
	go readInput(messages)

	envType := "test"
	processLinks(envType, messages)
}