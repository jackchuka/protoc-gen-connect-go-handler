package generator

import (
	"testing"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  *Options
		expectErr bool
	}{
		{
			name:  "minimal options",
			input: "out=gen",
			expected: &Options{
				Mode:       "per_service",
				DirPattern: "",
				ImplSuffix: "_handler",
				Out:        "gen",
			},
		},
		{
			name:  "per_method mode",
			input: "out=gen,mode=per_method",
			expected: &Options{
				Mode:       "per_method",
				DirPattern: "",
				ImplSuffix: "_handler",
				Out:        "gen",
			},
		},
		{
			name:  "custom suffix and pattern",
			input: "out=gen,impl_suffix=_impl,dir_pattern={package_path}/{service_snake}",
			expected: &Options{
				Mode:       "per_service",
				DirPattern: "{package_path}/{service_snake}",
				ImplSuffix: "_impl",
				Out:        "gen",
			},
		},
		{
			name:  "missing out",
			input: "mode=per_method,impl_suffix=_impl,dir_pattern={package_path}/{service_snake}",
			expected: &Options{
				Mode:       "per_method",
				DirPattern: "{package_path}/{service_snake}",
				ImplSuffix: "_impl",
				Out:        "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseOptions(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("parseOptions() expected error, got nil")
				}
				return
			}
			if opts.Mode != tt.expected.Mode {
				t.Errorf("Mode = %v, want %v", opts.Mode, tt.expected.Mode)
			}
			if opts.DirPattern != tt.expected.DirPattern {
				t.Errorf("DirPattern = %v, want %v", opts.DirPattern, tt.expected.DirPattern)
			}
			if opts.ImplSuffix != tt.expected.ImplSuffix {
				t.Errorf("ImplSuffix = %v, want %v", opts.ImplSuffix, tt.expected.ImplSuffix)
			}
		})
	}
}
