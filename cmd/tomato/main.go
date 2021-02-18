package main

import (
	"log"

	"github.com/CGA1123/tomato"
)

// TODO: add flags and parse them, implement cli
// TODO: move server starting into here, add pidfile, auto fork and detach if
// cli cmd runs and server is not running.
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
