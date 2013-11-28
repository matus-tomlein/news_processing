package main

import (
	"fmt"
	"github.com/matus-tomlein/news_processing/helpers"
	"github.com/matus-tomlein/news_processing/environment"
	"github.com/matus-tomlein/news_processing/link_search"
	"github.com/matus-tomlein/news_processing/web_server"
	"github.com/matus-tomlein/news_processing/article_downloader"
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
		fmt.Println("2 - Web server")
		fmt.Println("3 - Article downloader")
		fmt.Println("4 - Latest links")

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

		case 2:
			web_server.Serve()
			return

		case 3:
			article_downloader.StartDownload(environment.Current(), getMessagesChan())
			return

		case 4:
			link_search.FindLatestLinks()
			return

		}
	}
}
