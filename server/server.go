package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CGA1123/tomato/pb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	Duration = 25 * time.Minute
)

type Server struct {
	pb.UnimplementedTomatoServiceServer

	mut    sync.Mutex
	ends   time.Time
	tomato *time.Timer
}

func New() *Server {
	return &Server{}
}

func (s *Server) stop() time.Duration {
	if s.tomato == nil {
		return time.Duration(0)
	}

	remaining := s.remaining()

	s.tomato = nil
	s.ends = time.Now()

	return remaining
}

func (s *Server) start() (time.Time, error) {
	if s.tomato != nil {
		return time.Now(), fmt.Errorf("tomato is still runnning")
	}

	s.tomato = time.AfterFunc(Duration, func() { s.stop() })
	s.ends = time.Now().Add(Duration)

	return s.ends, nil
}

func (s *Server) remaining() time.Duration {
	if s.tomato == nil {
		return time.Duration(0)
	}

	return time.Until(s.ends)
}
func (s *Server) Start(ctx context.Context, _ *emptypb.Empty) (*timestamppb.Timestamp, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	ends, err := s.start()

	return timestamppb.New(ends), err

}

func (s *Server) Stop(ctx context.Context, _ *emptypb.Empty) (*durationpb.Duration, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	return durationpb.New(s.stop()), nil
}

func (s *Server) Running(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	running := s.tomato != nil

	return wrapperspb.Bool(running), nil
}

func (s *Server) Remaining(ctx context.Context, _ *emptypb.Empty) (*durationpb.Duration, error) {
	return durationpb.New(s.remaining()), nil
}
