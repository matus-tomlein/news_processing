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
	// Create postgres connection
	db, err := sql.Open("postgres", environment.PostgresConnectionString(envType))
	if err != nil { panic(err) }
	defer db.Close()

	messages := make(chan string)
	go readInput(messages)

	envType := "production"
	processLinks(envType, messages)
}
