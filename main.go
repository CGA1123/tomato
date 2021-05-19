package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/CGA1123/tomato/pb"
	"github.com/soellman/pidfile"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	Duration           = 25 * time.Minute
	ErrTomatoIsRunning = fmt.Errorf("tomato is still runnning")
	Socket             = "/tmp/tomato.sock"
	LogFile            = "/tmp/tomato.log"
	PidFile            = "/tmp/tomato.pid"
	LogPrefix          = "🍅 "
	Quiet              = false
	ErrNotRunning      = errors.New("not running")
	Commands           = map[string]func() error{
		"start":     WithClient(Start),
		"stop":      WithClient(Stop),
		"remaining": WithClient(Remaining),
		"running":   WithClient(Running),
		"kill":      Kill,
		"up":        Up,
		"server":    Server,
		"help":      Help}
)

type tomatoServiceServer struct {
	pb.UnimplementedTomatoServiceServer

	mut    sync.Mutex
	ends   time.Time
	tomato *time.Timer
}

func (s *tomatoServiceServer) stop() time.Duration {
	if s.tomato == nil {
		return time.Duration(0)
	}

	remaining := s.remaining()

	s.tomato = nil
	s.ends = time.Now()

	return remaining
}

func (s *tomatoServiceServer) start() (time.Time, error) {
	if s.tomato != nil {
		return time.Now(), ErrTomatoIsRunning
	}

	s.tomato = time.AfterFunc(Duration, func() { s.stop() })
	s.ends = time.Now().Add(Duration)

	return s.ends, nil
}

func (s *tomatoServiceServer) remaining() time.Duration {
	if s.tomato == nil {
		return time.Duration(0)
	}

	return time.Until(s.ends)
}
func (s *tomatoServiceServer) Start(ctx context.Context, _ *emptypb.Empty) (*timestamppb.Timestamp, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	ends, err := s.start()

	return timestamppb.New(ends), err

}

func (s *tomatoServiceServer) Stop(ctx context.Context, _ *emptypb.Empty) (*durationpb.Duration, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	return durationpb.New(s.stop()), nil
}

func (s *tomatoServiceServer) Running(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	running := s.tomato != nil

	return wrapperspb.Bool(running), nil
}

func (s *tomatoServiceServer) Remaining(ctx context.Context, _ *emptypb.Empty) (*durationpb.Duration, error) {
	return durationpb.New(s.remaining()), nil
}

type Client struct {
	client pb.TomatoServiceClient
}

func NewClient(socket string) (*Client, error) {
	conn, err := grpc.Dial("unix://"+socket, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{client: pb.NewTomatoServiceClient(conn)}, nil
}

func (c *Client) Start() (time.Time, error) {
	endsAt, err := c.client.Start(context.Background(), &emptypb.Empty{})
	if err != nil {
		return time.Now(), err
	}

	return endsAt.AsTime(), err
}

func (c *Client) Stop() (time.Duration, error) {
	left, err := c.client.Stop(context.Background(), &emptypb.Empty{})
	if err != nil {
		return time.Duration(0), err
	}

	return left.AsDuration(), err
}

func (c *Client) Remaining() (time.Duration, error) {
	left, err := c.client.Remaining(context.Background(), &emptypb.Empty{})
	if err != nil {
		return time.Duration(0), err
	}

	return left.AsDuration(), err
}

func (c *Client) Running() (bool, error) {
	running, err := c.client.Running(context.Background(), &emptypb.Empty{})
	if err != nil {
		return false, err
	}

	return running.GetValue(), err
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

func WithClient(f func(*Client) error) func() error {
	return func() error {
		client, err := client()
		if err != nil {
			return err
		}

		return f(client)
	}
}

func Help() error {
	usage()
	return nil
}

func Up() error {
	if !serverRunning() {
		return ErrNotRunning
	}

	return nil
}

func Kill() error {
	if !serverRunning() {
		return errors.New("server is not running")
	}

	pid, err := pidfileContents(PidFile)
	if err != nil {
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}

	return process.Signal(syscall.SIGTERM)
}

func Running(c *Client) error {
	running, err := c.Running()
	if err != nil {
		return err
	}

	if !running {
		return ErrNotRunning
	}

	return nil
}

func Start(c *Client) error {
	finish, err := c.Start()
	if err != nil {
		return err
	}

	fmtd := finish.Format("15:04")
	if Quiet {
		fmt.Println(fmtd)
	} else {
		log.Printf("timer will finish at %v", fmtd)
	}

	return nil
}

func Stop(c *Client) error {
	left, err := c.Stop()
	if err != nil {
		return err
	}

	minutes := left.Round(time.Minute).Minutes()

	if Quiet {
		fmt.Printf("%.0f\n", minutes)
	} else {
		log.Printf("there was %.0f minute(s) left on the clock!", minutes)
	}

	return nil
}

func Remaining(c *Client) error {
	left, err := c.Remaining()
	if err != nil {
		return err
	}

	if left == time.Duration(0) {
		if Quiet {
			fmt.Println("0")
		} else {
			log.Printf("the clock was not running!")
			log.Printf("use `tomato start` to get going.")
		}
		return nil
	}

	minutes := left.Round(time.Minute).Minutes()

	if Quiet {
		fmt.Printf("%.0f\n", minutes)
	} else {
		log.Printf("there are %.0f minutes left on the clock!", minutes)
	}
	return nil
}

func Server() error {
	if err := pidfile.WriteControl(PidFile, os.Getpid(), true); err != nil {
		return err
	}

	f, err := os.Create(LogFile)
	if err != nil {
		log.Printf("error creating logfile: %v", err)
	}

	defer func() {
		log.Printf("Shutting down...")
		log.Printf("Removing pidfile...")
		pidfile.Remove(PidFile)
		log.Printf("Removing socket...")
		os.RemoveAll(Socket)
		log.Printf("👋")
		f.Close()
	}()

	log.Printf("Will write logs to: %v", LogFile)
	log.SetOutput(f)
	log.SetPrefix(LogPrefix)
	log.Printf("Starting server at [%s]...", Socket)

	listener, err := net.Listen("unix", Socket)
	if err != nil {
		return fmt.Errorf("error opening socket: %w", err)
	}

	server := grpc.NewServer()
	pb.RegisterTomatoServiceServer(server, &tomatoServiceServer{})

	errorC := make(chan error, 1)
	shutdownC := make(chan os.Signal, 1)

	go func(errC chan<- error) {
		errorC <- server.Serve(listener)
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
		return nil, errors.New("tomato server is not running")
	}

	c, err := NewClient(Socket)
	if err != nil {
		log.Printf("is the server running? start it with tomato server")
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return c, err
}

func usage() {
	log.Printf("usage: tomato [-quiet] {start,stop,remaining,server,running,kill,up,help}")
}

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	flag.BoolVar(&Quiet, "quiet", false, "")
	flag.Parse()

	cmd, ok := Commands[flag.Arg(0)]
	if !ok {
		exitCode = 1
		usage()
		return
	}

	if err := cmd(); err != nil {
		if err == ErrNotRunning {
			exitCode = 33
		}

		log.Printf("error: %v", err)
		exitCode = 1
		return
	}
}
