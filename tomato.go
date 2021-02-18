package tomato

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

var (
	Duration           = 25 * time.Minute
	ErrTomatoIsRunning = fmt.Errorf("tomato is still runnning")
	Socket             = "/tmp/tomato.sock"
	LogFile            = "/tmp/tomato.log"
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

	timeLeft := s.ends.Sub(time.Now())

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

	return s.ends.Sub(time.Now())
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
