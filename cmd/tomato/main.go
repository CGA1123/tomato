package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/CGA1123/tomato"
	"github.com/soellman/pidfile"
)

func serverRunning() bool {
	pid, err := pidfileContents(tomato.PidFile)
	if err != nil {
		return false
	}

	return pidIsRunning(pid)
}

func pidfileContents(filename string) (int, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return 0, pidfile.ErrFileInvalid
	}

	return pid, nil
}

func pidIsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))

	if err != nil && err.Error() == "no such process" {
		return false
	}

	if err != nil && err.Error() == "os: process already finished" {
		return false
	}

	return true
}

var Commands map[string]func() error = map[string]func() error{
	"start":     WithClient(Start),
	"stop":      WithClient(Stop),
	"remaining": WithClient(Remaining),
	"done":      WithClient(Done),
	"server":    Server}

func WithClient(f func(*tomato.RPCClient) error) func() error {
	return func() error {
		client, err := client()
		if err != nil {
			return err
		}

		return f(client)
	}
}

func Start(c *tomato.RPCClient) error {
	finish, err := c.Start()
	if err != nil {
		return err
	}

	log.Printf("timer will finish at %v", finish)
	return nil
}

func Stop(c *tomato.RPCClient) error {
	left, err := c.Stop()
	if err != nil {
		return err
	}

	log.Printf("there was %v left on the clock!", left)
	return nil
}

func Remaining(c *tomato.RPCClient) error {
	left, err := c.Remaining()
	if err != nil {
		return err
	}

	if left == time.Duration(0) {
		log.Printf("the clock was not running!")
		log.Printf("use `tomato start` to get going.")

		return nil
	}

	log.Printf("there is %v left on the clock!", left)
	return nil
}

var ErrNotDone = errors.New("not done")

func Done(c *tomato.RPCClient) error {
	ok, err := c.IsDone()
	if err != nil {
		return err
	}

	if !ok {
		return ErrNotDone
	}

	return nil
}

func Server() error {
	defer func() {
		log.Printf("Shutting down...")
		log.Printf("Removing socket...")
		os.RemoveAll(tomato.Socket)
		log.Printf("👋")
	}()

	if err := pidfile.WriteControl(tomato.PidFile, os.Getpid(), true); err != nil {
		return err
	}
	defer pidfile.Remove(tomato.PidFile)

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
		return fmt.Errorf("error opening socket: %w", err)
	}

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

func client() (*tomato.RPCClient, error) {
	log.SetFlags(0)
	log.SetPrefix(tomato.LogPrefix)

	if !serverRunning() {
		log.Printf("tomato server is not running")
		log.Printf("you need to start it before attempting to use tomato")
		return nil, errors.New("")
	}

	c, err := tomato.NewClient(tomato.Socket)
	if err != nil {
		log.Printf("is the server running? start it with tomato-server")
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return c, err
}

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	if len(os.Args) != 2 {
		exitCode = 1
		log.Printf("usage: tomato {start,stop,remaining,done}")
		return
	}

	cmd, ok := Commands[os.Args[1]]
	if !ok {
		exitCode = 1
		log.Printf("usage: tomato {start,stop,remaining,done}")
		return
	}

	if err := cmd(); err != nil {
		if err == ErrNotDone {
			exitCode = 3
			return
		}

		log.Printf("error: %v", err)
		exitCode = 1
		return
	}
}
