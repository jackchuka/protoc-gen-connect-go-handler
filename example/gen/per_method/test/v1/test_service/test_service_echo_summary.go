package test_v1

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"example.com/test/gen/proto/test/v1"
)

// EchoSummary implements the EchoSummary RPC
func (t *TestServiceHandler) EchoSummary(
	ctx context.Context,
	req *connect.Request[testv1.EchoSummaryRequest],
) (*connect.Response[testv1.EchoSummaryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented,
		errors.New("EchoSummary not implemented"))
}
