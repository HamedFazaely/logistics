package main

import (
	"log"

	"gitlab.com/Hamed1984/logistics/pkg/ui"
)

func main() {
	err := ui.StartServer()
	if err != nil {
		log.Fatal(err)
	}
}
