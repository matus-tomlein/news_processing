package main

import (
	"fmt"
	"github.com/matus-tomlein/news_processing/environment"
	"github.com/matus-tomlein/news_processing/helpers"
	"github.com/matus-tomlein/news_processing/link_search"
)

func getMessagesChan() (chan string) {
	messages := make(chan string)
	go helpers.ReadInput(messages)
	return messages
}

func main() {
	for {
		fmt.Println("Choose an action:")
		fmt.Println("1 - Link search")
		var i int
		_, err := fmt.Scanf("%d", &i)
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch (i) {
		case 1:
			link_search.StartSearch(getMessagesChan())
			return
		}
	}
}
