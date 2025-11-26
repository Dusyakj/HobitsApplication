package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}

func (c *Client) HealthCheck(ctx context.Context) error {
	state := c.conn.GetState()
	if state.String() == "READY" {
		return nil
	}
	return fmt.Errorf("connection not ready: %s", state)
}
