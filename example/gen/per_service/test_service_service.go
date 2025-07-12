package test_v1
import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"example.com/test/gen/proto/test/v1"
)

// TestServiceHandler handles TestService RPCs
type TestServiceHandler struct {
	// Add your dependencies here (DB, logger, etc.)
}

// NewTestServiceHandler creates a new TestServiceHandler handler
func NewTestServiceHandler() *TestServiceHandler {
	return &TestServiceHandler{}
}

// Echo implements the Echo RPC
func (t *TestServiceHandler) Echo(
	ctx context.Context,
	req *connect.Request[testv1.EchoRequest],
) (*connect.Response[testv1.EchoResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented,
		errors.New("Echo not implemented"))
}

// EchoSummary implements the EchoSummary RPC
func (t *TestServiceHandler) EchoSummary(
	ctx context.Context,
	req *connect.Request[testv1.EchoSummaryRequest],
) (*connect.Response[testv1.EchoSummaryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented,
		errors.New("EchoSummary not implemented"))
}
