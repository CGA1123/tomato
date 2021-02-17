package main

import (
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.DialHTTP("unix", "/tmp/tomato.sock")
	if err != nil {
		log.Printf("error: %e", err)
		return
	}

	var reply string
	if err := client.Call("Ping.Do", "ping", &reply); err != nil {
		log.Printf("error: %e", err)
		return
	}

	log.Printf("got: %s", reply)
}
