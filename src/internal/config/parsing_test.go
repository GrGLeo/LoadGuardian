package config_test

import (
	"os"
	"testing"

	"github.com/GrGLeo/LoadBalancer/src/internal/config"
)

func TestParseEnvs(t *testing.T) {
	os.Setenv("FOO", "bar")
	os.Setenv("BAZ", "qux")
	defer os.Unsetenv("FOO")
	defer os.Unsetenv("BAZ")

	// Test cases
	tests := []struct {
		name     string
		envs     []string
		expected []string
	}{
		{
			name:     "Valid environment variables",
			envs:     []string{"$FOO", "$BAZ"},
			expected: []string{"FOO=bar", "BAZ=qux"},
		},
		{
			name:     "Invalid environment variable",
			envs:     []string{"$FOO", "$INVALID"},
			expected: []string{"FOO=bar", "INVALID="},
		},
		{
			name:     "Empty input",
			envs:     []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedEnvs := config.ParseEnvs(tt.envs)
			if len(parsedEnvs) != len(tt.expected) {
				t.Errorf("ParseEnvs() length = %v, want %v", len(parsedEnvs), len(tt.expected))
			}

			for i, v := range parsedEnvs {
				if v != tt.expected[i] {
					t.Errorf("ParseEnvs()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}
