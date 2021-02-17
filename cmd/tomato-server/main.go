// tomato-server entrypoint creates the tomato rpc server to be used interacted
// with by the tomato CLI
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

const (
	Tomato = "🍅"
	Socket = "/tmp/tomato.sock"
)

type Ping struct{}

func (p Ping) Do(str string, reply *string) error {
	log.Printf("%s got: %s", Tomato, str)
	*reply = str

	return nil
}

func main() {
	log.Println(fmt.Sprintf("%s Starting server at [%s]...", Tomato, Socket))

	ping := &Ping{}
	if err := rpc.Register(ping); err != nil {
		log.Printf("%s Failed to register rpc: %e", Tomato, err)
		return
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("unix", Socket)
	if err != nil {
		log.Printf("%s Error opening socket: %e", Tomato, err)
		return
	}

	http.Serve(listener, nil)
}
