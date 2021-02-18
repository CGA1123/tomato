package main

import (
	"log"

	"github.com/CGA1123/tomato"
)

func main() {
	log.SetFlags(0)

	client, err := tomato.NewClient(tomato.Socket)
	if err != nil {
		log.Printf("error creating client: %v", err)
		return
	}

	ends, err := client.Start()
	if err != nil {
		if err.Error() == tomato.ErrTomatoIsRunning.Error() {
			log.Printf("tomato is till running! use tomato stop or tomato start -f")
		} else {
			log.Printf("error: %v", err)
		}
		return
	}

	log.Printf("got: %v", ends)
}
