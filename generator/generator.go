package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	TEMPLATE_METHOD      = "method_stub"
	TEMPLATE_METHOD_ONLY = "method_only"
	TEMPLATE_SERVICE     = "service_manifest"
	TEMPLATE_STRUCT      = "struct_stub"
)

// Generate processes the CodeGeneratorRequest and returns a CodeGeneratorResponse
func Generate(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	// Parse plugin options
	opts, err := parseOptions(req.GetParameter())
	if err != nil {
		return nil, err
	}

	var files []*pluginpb.CodeGeneratorResponse_File

	// Process each file to generate
	for _, fileName := range req.GetFileToGenerate() {
		var fileDesc *descriptorpb.FileDescriptorProto
		for _, fd := range req.GetProtoFile() {
			if fd.GetName() == fileName {
				fileDesc = fd
				break
			}
		}

		if fileDesc == nil {
			continue
		}

		// Process each service in the file
		for _, svc := range fileDesc.GetService() {
			generatedFiles, err := generateServiceFiles(fileDesc, svc, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to generate files for service %s: %w", svc.GetName(), err)
			}
			files = append(files, generatedFiles...)
		}
	}

	return &pluginpb.CodeGeneratorResponse{
		File: files,
	}, nil
}

// generateServiceFiles generates all files for a single service
func generateServiceFiles(fileDesc *descriptorpb.FileDescriptorProto, svc *descriptorpb.ServiceDescriptorProto, opts *Options) ([]*pluginpb.CodeGeneratorResponse_File, error) {
	ctx := buildContext(fileDesc, svc, opts)
	var files []*pluginpb.CodeGeneratorResponse_File

	// 1. Generate manifest file (always regenerated)
	manifestFiles, err := generateManifestFile(ctx)
	if err != nil {
		return nil, err
	}
	files = append(files, manifestFiles...)

	// 2. Generate struct file and methods based on mode
	if opts.Mode == modePerMethod {
		structFiles, err := generateStructFileIfNeeded(ctx, opts)
		if err != nil {
			return nil, err
		}
		files = append(files, structFiles...)

		methodFiles, err := generatePerMethodFiles(fileDesc, svc, ctx, opts)
		if err != nil {
			return nil, err
		}
		files = append(files, methodFiles...)
	} else {
		structFiles, err := generatePerServiceStructFile(fileDesc, svc, ctx, opts)
		if err != nil {
			return nil, err
		}
		files = append(files, structFiles...)
	}

	return files, nil
}

// generateManifestFile generates the service manifest file
func generateManifestFile(ctx Context) ([]*pluginpb.CodeGeneratorResponse_File, error) {
	manifestContent, err := renderTemplate(TEMPLATE_SERVICE, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render manifest template: %w", err)
	}

	return []*pluginpb.CodeGeneratorResponse_File{
		{
			Name:    &ctx.ManifestPath,
			Content: &manifestContent,
		},
	}, nil
}

// generateStructFileIfNeeded generates the struct file only if it doesn't exist
func generateStructFileIfNeeded(ctx Context, opts *Options) ([]*pluginpb.CodeGeneratorResponse_File, error) {
	fullStructPath := constructFullPath(opts.Out, ctx.StructPath)
	if !fileExists(fullStructPath) {
		structContent, err := renderTemplate(TEMPLATE_STRUCT, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to render struct template: %w", err)
		}

		return []*pluginpb.CodeGeneratorResponse_File{
			{
				Name:    &ctx.StructPath,
				Content: &structContent,
			},
		}, nil
	}
	return nil, nil
}

// generatePerMethodFiles generates individual method files for per-method mode
func generatePerMethodFiles(fileDesc *descriptorpb.FileDescriptorProto, svc *descriptorpb.ServiceDescriptorProto, ctx Context, opts *Options) ([]*pluginpb.CodeGeneratorResponse_File, error) {
	var files []*pluginpb.CodeGeneratorResponse_File

	for _, method := range svc.GetMethod() {
		methodCtx := ctx
		methodCtx.Method = &MethodContext{
			Name:   method.GetName(),
			Input:  convertProtoTypeToGo(method.GetInputType(), fileDesc),
			Output: convertProtoTypeToGo(method.GetOutputType(), fileDesc),
		}

		methodFileBase := fmt.Sprintf("%s_%s",
			toSnakeCase(svc.GetName()), toSnakeCase(method.GetName()))
		methodPath := filepath.Join(ctx.Dir, methodFileBase+".go")
		methodCtx.MethodPath = methodPath

		// Only generate if method is not already implemented
		fullMethodPath := constructFullPath(opts.Out, methodPath)
		if !fileExists(fullMethodPath) {
			methodContent, err := renderTemplate(TEMPLATE_METHOD, methodCtx)
			if err != nil {
				return nil, fmt.Errorf("failed to render method template: %w", err)
			}

			files = append(files, &pluginpb.CodeGeneratorResponse_File{
				Name:    &methodPath,
				Content: &methodContent,
			})
		}
	}

	return files, nil
}

// generatePerServiceStructFile handles per-service mode by building the complete struct file with all methods
func generatePerServiceStructFile(fileDesc *descriptorpb.FileDescriptorProto, svc *descriptorpb.ServiceDescriptorProto, ctx Context, opts *Options) ([]*pluginpb.CodeGeneratorResponse_File, error) {
	fullStructPath := constructFullPath(opts.Out, ctx.StructPath)

	// Check if file already exists on disk (from previous runs)
	var existingContent string
	if fileExists(fullStructPath) {
		// Read existing file content
		content, err := os.ReadFile(fullStructPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing struct file: %w", err)
		}
		existingContent = string(content)
	} else {
		// Generate base struct content
		structContent, err := renderTemplate(TEMPLATE_STRUCT, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to render struct template: %w", err)
		}
		existingContent = structContent
	}

	// Build new methods to append
	var newMethods []string
	for _, method := range svc.GetMethod() {
		if !FuncExists(fullStructPath, ctx.StructName, method.GetName()) {
			methodCtx := ctx
			methodCtx.Method = &MethodContext{
				Name:   method.GetName(),
				Input:  convertProtoTypeToGo(method.GetInputType(), fileDesc),
				Output: convertProtoTypeToGo(method.GetOutputType(), fileDesc),
			}

			methodContent, err := renderTemplate(TEMPLATE_METHOD_ONLY, methodCtx)
			if err != nil {
				return nil, fmt.Errorf("failed to render method template: %w", err)
			}
			newMethods = append(newMethods, methodContent)
		}
	}

	// Combine existing content with new methods
	var finalContent string
	if len(newMethods) > 0 {
		finalContent = existingContent + "\n" + strings.Join(newMethods, "\n\n")
	} else {
		finalContent = existingContent
	}

	return []*pluginpb.CodeGeneratorResponse_File{
		{
			Name:    &ctx.StructPath,
			Content: &finalContent,
		},
	}, nil
}

// Context holds template data for code generation
type Context struct {
	PackageName  string
	StructName   string
	Receiver     string // e.g. "h"
	Service      *ServiceContext
	Method       *MethodContext
	ManifestPath string
	StructPath   string
	MethodPath   string
	Dir          string
	Mode         string
	ProtoImport  string
}

type ServiceContext struct {
	Name    string
	Methods []*MethodContext
}

type MethodContext struct {
	Name   string
	Input  string
	Output string
}

// buildContext creates a template context for a service
func buildContext(fileDesc *descriptorpb.FileDescriptorProto, svc *descriptorpb.ServiceDescriptorProto, opts *Options) Context {
	serviceName := svc.GetName()
	structName := serviceName + "Handler"

	// Build output directory
	dir := ""
	if opts.DirPattern != "" {
		dir = expandPlaceholders(opts.DirPattern, fileDesc, svc)
	}

	manifestPath := filepath.Join(dir, toSnakeCase(serviceName)+opts.ImplSuffix+".gen.go")
	structPath := filepath.Join(dir, toSnakeCase(serviceName)+opts.ImplSuffix+".go")

	// Build method contexts
	var methods []*MethodContext
	for _, method := range svc.GetMethod() {
		methods = append(methods, &MethodContext{
			Name:   method.GetName(),
			Input:  convertProtoTypeToGo(method.GetInputType(), fileDesc),
			Output: convertProtoTypeToGo(method.GetOutputType(), fileDesc),
		})
	}

	// Extract proto import path from go_package option
	protoImport := extractGoPackageImport(fileDesc.GetOptions().GetGoPackage())

	return Context{
		PackageName: generalizePackageName(fileDesc.GetPackage()),
		StructName:  structName,
		Receiver:    strings.ToLower(structName[:1]), // e.g. "h" for "Handler"
		Service: &ServiceContext{
			Name:    serviceName,
			Methods: methods,
		},
		ManifestPath: manifestPath,
		StructPath:   structPath,
		Dir:          dir,
		Mode:         opts.Mode,
		ProtoImport:  protoImport,
	}
}

// generalizePackageName converts a package name to a more Go-friendly format
func generalizePackageName(pkg string) string {
	// Remove leading dot if present
	pkg = strings.TrimPrefix(pkg, ".")

	// Replace dots with underscores for Go package compatibility
	pkg = strings.ReplaceAll(pkg, ".", "_")

	// Ensure it starts with a letter (Go package names should not start with a number)
	if len(pkg) > 0 && ('0' <= pkg[0] && pkg[0] <= '9') {
		pkg = "pkg_" + pkg
	}

	return pkg
}

// expandPlaceholders expands placeholders in directory patterns
func expandPlaceholders(pattern string, fileDesc *descriptorpb.FileDescriptorProto, svc *descriptorpb.ServiceDescriptorProto) string {
	pkg := fileDesc.GetPackage()
	serviceName := svc.GetName()

	replacements := map[string]string{
		"{package}":       pkg,
		"{package_path}":  strings.ReplaceAll(pkg, ".", "/"),
		"{service}":       serviceName,
		"{service_snake}": toSnakeCase(serviceName),
	}

	result := pattern
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteByte('_')
		}
		if 'A' <= r && r <= 'Z' {
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// convertProtoTypeToGo converts a protobuf type name to Go type name
func convertProtoTypeToGo(protoType string, fileDesc *descriptorpb.FileDescriptorProto) string {
	// Remove leading dot if present
	protoType = strings.TrimPrefix(protoType, ".")

	// Split the type name to get package and type
	parts := strings.Split(protoType, ".")
	if len(parts) < 2 {
		return protoType // fallback to original if can't parse
	}

	// Extract the message name (last part)
	messageName := parts[len(parts)-1]

	// Extract package parts (all but last)
	protoPackage := strings.Join(parts[:len(parts)-1], ".")

	// Check if this type is from the current file
	if protoPackage == fileDesc.GetPackage() {
		// Extract Go package name from go_package option
		goPackageOption := fileDesc.GetOptions().GetGoPackage()
		goPackage := extractGoPackageName(goPackageOption)
		return goPackage + "." + messageName
	}

	// For external packages, fall back to the underscore approach
	// This would need to be enhanced to look up the actual go_package for external types
	goPackage := strings.ReplaceAll(protoPackage, ".", "_")
	return goPackage + "." + messageName
}

// extractGoPackageName extracts the package name from go_package option
// Example: "example/gen/test/v1;testv1" -> "testv1"
// Example: "example/gen/test/v1" -> "v1" (last part of path)
func extractGoPackageName(goPackage string) string {
	if goPackage == "" {
		return ""
	}

	// go_package format: "import/path;package_name" or just "import/path"
	if semicolon := strings.Index(goPackage, ";"); semicolon != -1 {
		return goPackage[semicolon+1:]
	}

	// If no semicolon, use the last part of the import path
	parts := strings.Split(goPackage, "/")
	return parts[len(parts)-1]
}

// extractGoPackageImport extracts the import path from go_package option
// Example: "example/gen/test/v1;testv1" -> "example/gen/test/v1"
func extractGoPackageImport(goPackage string) string {
	if goPackage == "" {
		return ""
	}

	// go_package format: "import/path;package_name" or just "import/path"
	if semicolon := strings.Index(goPackage, ";"); semicolon != -1 {
		return goPackage[:semicolon]
	}

	return goPackage
}

// constructFullPath builds the full file path including output directory
func constructFullPath(outputDir, relativePath string) string {
	if outputDir == "" {
		// If no output directory specified, use relative path as-is
		return relativePath
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path if we can't get cwd
		return filepath.Join(outputDir, relativePath)
	}

	// Combine: current_working_directory + output_directory + relative_path
	return filepath.Join(cwd, outputDir, relativePath)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
