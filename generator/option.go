package generator

import (
	"fmt"
	"strings"
)

const (
	modePerService = "per_service"
	modePerMethod  = "per_method"
)

// Options represents the plugin configuration
type Options struct {
	Mode       string // "per_service" or "per_method"
	DirPattern string // directory pattern with placeholders
	ImplSuffix string // suffix for implementation files
	Out        string // output directory from buf.gen.yaml
}

// parseOptions parses the plugin parameter string
func parseOptions(parameter string) (*Options, error) {
	opts := &Options{
		Mode:       modePerService,
		DirPattern: "",
		ImplSuffix: "_handler",
		Out:        "",
	}

	for pair := range strings.SplitSeq(parameter, ",") {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])

		switch key {
		case "mode":
			if value == modePerService || value == modePerMethod {
				opts.Mode = value
			}
		case "dir_pattern":
			opts.DirPattern = value
		case "impl_suffix":
			opts.ImplSuffix = value
		case "out":
			opts.Out = value
		}
	}

	if opts.Out == "" {
		return nil, fmt.Errorf("missing required option 'out'")
	}

	return opts, nil
}
