package client

import (
	"context"
	"time"

	"github.com/CGA1123/tomato/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	client pb.TomatoServiceClient
}

func New(socket string) (*Client, error) {
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
