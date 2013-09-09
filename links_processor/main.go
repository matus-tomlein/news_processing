package main

import (
	"fmt"
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
	messages := make(chan string)
	go readInput(messages)

	envType := "production"
	processLinks(envType, messages)
}
