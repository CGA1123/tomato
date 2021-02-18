package main

import (
	"log"

	"github.com/CGA1123/tomato"
)

// TODO: add flags and parse them, implement cli
func main() {
	log.SetFlags(0)
	log.SetPrefix(tomato.LogPrefix)

	client, err := tomato.NewClient(tomato.Socket)
	if err != nil {
		log.Printf("error creating client: %v", err)
		log.Printf("is the server running? start it with tomato-server")
		return
	}

	ends, err := client.Start()
	if err != nil {
		if err.Error() == tomato.ErrTomatoIsRunning.Error() {
			log.Printf("tomato is still running!")
			log.Printf("use tomato stop or tomato start -f")
		} else {
			log.Printf("error: %v", err)
		}
		return
	}

	log.Printf("got: %v", ends)
}
