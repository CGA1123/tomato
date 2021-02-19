package tomato

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

type Tomato interface {
	Start() (endsAt time.Time, err error)
	Stop() (timeRemaining time.Duration)
	Remaining() (timeRemaining time.Duration)
	IsDone() bool
}

type State struct {
	mut    sync.Mutex
	ends   time.Time
	tomato *time.Timer
}

func NewState() *State {
	return &State{}
}

func (s *State) Start() (time.Time, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.tomato != nil {
		return time.Time{}, ErrTomatoIsRunning
	}

	s.ends = time.Now().Add(Duration)
	s.tomato = time.AfterFunc(Duration, func() { s.Stop() })

	return s.ends, nil
}

func (s *State) Stop() time.Duration {
	s.mut.Lock()
	defer s.mut.Unlock()

	timeLeft := time.Until(s.ends)

	s.tomato = nil
	s.ends = time.Time{}

	return timeLeft
}

func (s *State) Remaining() time.Duration {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.tomato == nil {
		return time.Duration(0)
	}

	return time.Until(s.ends)
}

func (s *State) IsDone() bool {
	return s.tomato == nil
}

type RPCServer struct {
	tomato Tomato
}

func NewRPCServer(tomato Tomato) *RPCServer {
	return &RPCServer{tomato: tomato}
}

func (s *RPCServer) Start(args int, reply *time.Time) error {
	log.Printf("got Start")

	endTime, err := s.tomato.Start()
	*reply = endTime

	return err
}

func (s *RPCServer) Stop(args int, reply *time.Duration) error {
	log.Printf("got Stop")

	*reply = s.tomato.Stop()

	return nil
}

func (s *RPCServer) Remaining(args int, reply *time.Duration) error {
	log.Printf("got Remaining")

	*reply = s.tomato.Remaining()

	return nil
}

func (s *RPCServer) IsDone(args int, reply *bool) error {
	log.Printf("got IsDone")

	*reply = s.tomato.IsDone()

	return nil
}

type RPCClient struct {
	client *rpc.Client
}

func NewClient(socket string) (*RPCClient, error) {
	client, err := rpc.DialHTTP("unix", socket)
	if err != nil {
		return nil, err
	}

	return &RPCClient{client: client}, nil
}

func (c *RPCClient) Start() (time.Time, error) {
	var endsAt time.Time
	err := c.client.Call("Tomato.Start", 0, &endsAt)

	return endsAt, err
}

func (c *RPCClient) Stop() (time.Duration, error) {
	var left time.Duration
	err := c.client.Call("Tomato.Stop", 0, &left)

	return left, err
}

func (c *RPCClient) Remaining() (time.Duration, error) {
	var left time.Duration
	err := c.client.Call("Tomato.Remaining", 0, &left)

	return left, err
}

func (c *RPCClient) IsDone() (bool, error) {
	var done bool
	err := c.client.Call("Tomato.IsDone", 0, &done)

	return done, err
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
	"done":      WithClient(Done),
	"server":    Server}

func WithClient(f func(*RPCClient) error) func() error {
	return func() error {
		client, err := client()
		if err != nil {
			return err
		}

		return f(client)
	}
}

func Start(c *RPCClient) error {
	finish, err := c.Start()
	if err != nil {
		return err
	}

	log.Printf("timer will finish at %v", finish)
	return nil
}

func Stop(c *RPCClient) error {
	left, err := c.Stop()
	if err != nil {
		return err
	}

	log.Printf("there was %v left on the clock!", left)
	return nil
}

func Remaining(c *RPCClient) error {
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

func Done(c *RPCClient) error {
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

	server := NewRPCServer(NewState())
	rpc.RegisterName("Tomato", server)
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

func client() (*RPCClient, error) {
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
