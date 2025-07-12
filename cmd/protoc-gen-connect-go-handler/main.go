package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jackchuka/protoc-gen-connect-go-handler/generator"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

const maxInputSize = 32 << 20 // 32 MiB

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "protoc-gen-connect-go-handler: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Limit input size for safety
	limitedReader := io.LimitReader(os.Stdin, maxInputSize)

	// Read the CodeGeneratorRequest from stdin
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Unmarshal the request
	var req pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Generate the response
	resp, err := generator.Generate(&req)
	if err != nil {
		// On error, return a response with the error message
		resp = &pluginpb.CodeGeneratorResponse{
			Error: proto.String(err.Error()),
		}
	}

	// Marshal the response
	respData, err := proto.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write the response to stdout
	if _, err := os.Stdout.Write(respData); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}
