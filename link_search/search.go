package link_search

import (
	"fmt"
)

func Search(query string) {
	
}

func StartSearch(messages chan string) {
	fmt.Println("Enter a search query:")

	query := <-messages
	fmt.Println(query)
}
