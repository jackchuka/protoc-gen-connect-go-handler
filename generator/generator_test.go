package generator

import (
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestService", "test_service"},
		{"APIService", "a_p_i_service"},
		{"simpleTest", "simple_test"},
		{"Test", "test"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandPlaceholders(t *testing.T) {
	pkg := "test.v1"
	fileDesc := &descriptorpb.FileDescriptorProto{
		Package: &pkg,
	}

	serviceName := "TestService"
	svc := &descriptorpb.ServiceDescriptorProto{
		Name: &serviceName,
	}

	tests := []struct {
		pattern  string
		expected string
	}{
		{"{package}", "test.v1"},
		{"{package_path}", "test/v1"},
		{"{service}", "TestService"},
		{"{service_snake}", "test_service"},
		{"{package_path}/{service_snake}", "test/v1/test_service"},
		{"handler", "handler"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := expandPlaceholders(tt.pattern, fileDesc, svc)
			if result != tt.expected {
				t.Errorf("expandPlaceholders(%v) = %v, want %v", tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	// Create a simple test request
	pkg := "test.v1"
	serviceName := "TestService"
	methodName := "Echo"
	inputType := "test.v1.EchoRequest"
	outputType := "test.v1.EchoResponse"
	fileName := "test/test_service.proto"
	parameter := "out=gen"

	req := &pluginpb.CodeGeneratorRequest{
		Parameter:      &parameter,
		FileToGenerate: []string{fileName},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			{
				Name:    &fileName,
				Package: &pkg,
				Service: []*descriptorpb.ServiceDescriptorProto{
					{
						Name: &serviceName,
						Method: []*descriptorpb.MethodDescriptorProto{
							{
								Name:       &methodName,
								InputType:  &inputType,
								OutputType: &outputType,
							},
						},
					},
				},
			},
		},
	}

	resp, err := Generate(req)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if len(resp.File) == 0 {
		t.Fatal("Expected generated files, got none")
	}

	// Check that we got the expected files
	var manifestFile, structFile *pluginpb.CodeGeneratorResponse_File
	for _, file := range resp.File {
		if file.GetName() == "test_service_handler.gen.go" {
			manifestFile = file
		}
		if file.GetName() == "test_service_handler.go" {
			structFile = file
		}
	}

	if manifestFile == nil {
		t.Error("Expected manifest file not generated")
	}
	if structFile == nil {
		t.Error("Expected struct file not generated")
	}

	if manifestFile != nil && !contains(manifestFile.GetContent(), "TestServiceHandler") {
		t.Error("Manifest file should contain TestServiceHandler")
	}

	if structFile != nil && !contains(structFile.GetContent(), "*TestServiceHandler) Echo") {
		t.Error("Struct file should contain Echo method")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
