package test_v1

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"example.com/test/gen/proto/test/v1"
)

// Echo implements the Echo RPC
func (t *TestServiceHandler) Echo(
	ctx context.Context,
	req *connect.Request[testv1.EchoRequest],
) (*connect.Response[testv1.EchoResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented,
		errors.New("Echo not implemented"))
}
