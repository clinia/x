package tritonx

import (
	"context"

	tritongrpc "github.com/clinia/x/tritonx/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TritonClient struct {
	conn *grpc.ClientConn

	inferenceServiceClient tritongrpc.GRPCInferenceServiceClient
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
		conn:                   conn,
		inferenceServiceClient: tritongrpc.NewGRPCInferenceServiceClient(conn),
	}, nil
}
