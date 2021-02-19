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
	"sync"
	"syscall"
	"time"

	"github.com/soellman/pidfile"
)

var (
	Duration           = 25 * time.Minute
	ErrTomatoIsRunning = fmt.Errorf("tomato is still runnning")
	Socket             = "/tmp/tomato.sock"
	LogFile            = "/tmp/tomato.log"
	PidFile            = "/tmp/tomato.pid"
	LogPrefix          = "🍅 "
)

type Tomato struct {
	mut    sync.Mutex
	ends   time.Time
	tomato *time.Timer
}

func NewTomato() *Tomato {
	return &Tomato{}
}

func (s *Tomato) do(action string, f func() error) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	log.Printf("Got %s", action)

	return f()
}

func (s *Tomato) stop() time.Duration {
	if s.tomato == nil {
		return time.Duration(0)
	}

	remaining := s.remaining()

	s.tomato = nil
	s.ends = time.Now()

	return remaining
}

func (s *Tomato) start() (time.Time, error) {
	if s.tomato != nil {
		return time.Now(), ErrTomatoIsRunning
	}

	s.tomato = time.AfterFunc(Duration, func() { s.stop() })
	s.ends = time.Now().Add(Duration)

	return s.ends, nil
}

func (s *Tomato) remaining() time.Duration {
	if s.tomato == nil {
		return time.Duration(0)
	}

	return time.Until(s.ends)
}

func (s *Tomato) Start(args int, reply *time.Time) error {
	return s.do("Start", func() error {
		ends, err := s.start()

		*reply = ends

		return err
	})
}

func (s *Tomato) Stop(args int, reply *time.Duration) error {
	return s.do("Start", func() error {
		*reply = s.stop()

		return nil
	})
}

func (s *Tomato) Remaining(args int, reply *time.Duration) error {
	return s.do("Start", func() error {
		*reply = s.remaining()

		return nil
	})
}

type Client struct {
	client *rpc.Client
}

func NewClient(socket string) (*Client, error) {
	client, err := rpc.DialHTTP("unix", socket)
	if err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

func (c *Client) Start() (time.Time, error) {
	var endsAt time.Time
	err := c.client.Call("Tomato.Start", 0, &endsAt)

	return endsAt, err
}

func (c *Client) Stop() (time.Duration, error) {
	var left time.Duration
	err := c.client.Call("Tomato.Stop", 0, &left)

	return left, err
}

func (c *Client) Remaining() (time.Duration, error) {
	var left time.Duration
	err := c.client.Call("Tomato.Remaining", 0, &left)

	return left, err
}

func serverRunning() bool {
	pid, err := pidfileContents(PidFile)
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
	"server":    Server}

func WithClient(f func(*Client) error) func() error {
	return func() error {
		client, err := client()
		if err != nil {
			return err
		}

		return f(client)
	}
}

func Start(c *Client) error {
	finish, err := c.Start()
	if err != nil {
		return err
	}

	log.Printf("timer will finish at %v", finish)
	return nil
}

func Stop(c *Client) error {
	left, err := c.Stop()
	if err != nil {
		return err
	}

	log.Printf("there was %v left on the clock!", left)
	return nil
}

func Remaining(c *Client) error {
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

func Server() error {
	defer func() {
		log.Printf("Shutting down...")
		log.Printf("Removing socket...")
		os.RemoveAll(Socket)
		log.Printf("👋")
	}()

	if err := pidfile.WriteControl(PidFile, os.Getpid(), true); err != nil {
		return err
	}
	defer pidfile.Remove(PidFile)

	f, err := os.Create(LogFile)
	if err != nil {
		log.Printf("error creating logfile: %v", err)
	}
	defer f.Close()

	log.Printf("Will write logs to: %v", LogFile)
	log.SetOutput(f)
	log.SetPrefix(LogPrefix)

	log.Printf("Starting server at [%s]...", Socket)

	server := NewTomato()
	rpc.Register(server)
	rpc.HandleHTTP()

	listener, err := net.Listen("unix", Socket)
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

func client() (*Client, error) {
	log.SetFlags(0)
	log.SetPrefix(LogPrefix)

	if !serverRunning() {
		log.Printf("tomato server is not running")
		log.Printf("you need to start it before attempting to use tomato")
		return nil, errors.New("")
	}

	c, err := NewClient(Socket)
	if err != nil {
		log.Printf("is the server running? start it with tomato server")
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return c, err
}

func usage() {
	log.Printf("usage: tomato {start,stop,remaining,server}")
}

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	if len(os.Args) != 2 {
		exitCode = 1
		usage()
		return
	}

	cmd, ok := Commands[os.Args[1]]
	if !ok {
		exitCode = 1
		usage()
		return
	}

	if err := cmd(); err != nil {
		log.Printf("error: %v", err)
		exitCode = 1
		return
	}
}
