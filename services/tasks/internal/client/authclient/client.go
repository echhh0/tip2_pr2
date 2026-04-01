package authclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tip2_pr2/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn   *grpc.ClientConn
	client proto.AuthServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		strings.TrimSpace(addr),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create grpc client: %w", err)
	}

	return &Client{
		conn:   conn,
		client: proto.NewAuthServiceClient(conn),
	}, nil
}

func (c *Client) Verify(ctx context.Context, token string) (bool, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.Verify(ctx, &proto.VerifyRequest{
		Token: token,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return false, 401, nil
			case codes.DeadlineExceeded:
				return false, 503, fmt.Errorf("auth verify timeout: %w", err)
			case codes.Unavailable:
				return false, 503, fmt.Errorf("auth service unavailable: %w", err)
			default:
				return false, 502, fmt.Errorf("auth grpc error: %w", err)
			}
		}

		return false, 502, fmt.Errorf("auth grpc request failed: %w", err)
	}

	return resp.Valid, 200, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
