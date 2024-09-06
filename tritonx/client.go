package tritonx

import (
	"context"

	tritongrpc "github.com/clinia/x/tritonx/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TritonClient struct {
	conn *grpc.ClientConn

	Inference tritongrpc.GRPCInferenceServiceClient
}

func NewTritonClient(ctx context.Context, cfg Config) (*TritonClient, error) {
	opts := []grpc.DialOption{}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, err
	}

	return &TritonClient{
		conn:      conn,
		Inference: tritongrpc.NewGRPCInferenceServiceClient(conn),
	}, nil
}

func (t *TritonClient) Close() error {
	return t.conn.Close()
}
