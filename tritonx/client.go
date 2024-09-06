package tritonx

import (
	"context"

	tritongrpc "github.com/clinia/x/tritonx/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn

	Inference tritongrpc.GRPCInferenceServiceClient
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	opts := []grpc.DialOption{}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:      conn,
		Inference: tritongrpc.NewGRPCInferenceServiceClient(conn),
	}, nil
}

func (t *Client) Close() error {
	return t.conn.Close()
}
