package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/CGA1123/tomato"
)

func run(listener net.Listener) error {
	errorC := make(chan error, 1)
	shutdownC := make(chan os.Signal, 1)

	go func(errC chan<- error) {
		errorC <- http.Serve(listener, nil)
	}(errorC)

	signal.Notify(shutdownC, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errorC:
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return err
	case <-shutdownC:
		return nil
	}
}

func main() {
	defer func() {
		log.Printf("Shutting down...")
		log.Printf("Removing socket...")
		os.RemoveAll(tomato.Socket)
		log.Printf("👋")
	}()

	f, err := os.Create(tomato.LogFile)
	if err != nil {
		log.Printf("error creating logfile: %v", err)
	}
	defer f.Close()

	log.Printf("Will write logs to: %v", tomato.LogFile)
	log.SetOutput(f)
	log.SetPrefix(tomato.LogPrefix)

	log.Printf("Starting server at [%s]...", tomato.Socket)

	server := tomato.NewRPCServer(tomato.NewState())
	rpc.RegisterName("Tomato", server)
	rpc.HandleHTTP()

	listener, err := net.Listen("unix", tomato.Socket)
	if err != nil {
		log.Printf("Error opening socket: %v", err)
		return
	}

	if err := run(listener); err != nil {
		log.Printf("error running server: %v", err)
		return
	}
}
