package config

import (
	"os"
	"testing"
)

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue string
		setEnv       bool
		want         string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_STRING_1",
			envValue:     "custom_value",
			defaultValue: "default",
			setEnv:       true,
			want:         "custom_value",
		},
		{
			name:         "returns default when not set",
			key:          "TEST_STRING_2",
			envValue:     "",
			defaultValue: "default_value",
			setEnv:       false,
			want:         "default_value",
		},
		{
			name:         "returns default when empty string",
			key:          "TEST_STRING_3",
			envValue:     "",
			defaultValue: "fallback",
			setEnv:       true,
			want:         "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up env
			os.Unsetenv(tt.key)

			if tt.setEnv && tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvString(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue int
		setEnv       bool
		want         int
	}{
		{
			name:         "returns parsed int",
			key:          "TEST_INT_1",
			envValue:     "42",
			defaultValue: 10,
			setEnv:       true,
			want:         42,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_INT_2",
			envValue:     "",
			defaultValue: 100,
			setEnv:       false,
			want:         100,
		},
		{
			name:         "returns default when invalid",
			key:          "TEST_INT_3",
			envValue:     "not_a_number",
			defaultValue: 50,
			setEnv:       true,
			want:         50,
		},
		{
			name:         "handles negative numbers",
			key:          "TEST_INT_4",
			envValue:     "-5",
			defaultValue: 0,
			setEnv:       true,
			want:         -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvInt(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetEnvFloat(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue float64
		setEnv       bool
		want         float64
	}{
		{
			name:         "returns parsed float",
			key:          "TEST_FLOAT_1",
			envValue:     "0.7",
			defaultValue: 0.1,
			setEnv:       true,
			want:         0.7,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_FLOAT_2",
			envValue:     "",
			defaultValue: 0.5,
			setEnv:       false,
			want:         0.5,
		},
		{
			name:         "returns default when invalid",
			key:          "TEST_FLOAT_3",
			envValue:     "not_a_float",
			defaultValue: 0.3,
			setEnv:       true,
			want:         0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvFloat(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvFloat() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue bool
		setEnv       bool
		want         bool
	}{
		{
			name:         "returns true when set to true",
			key:          "TEST_BOOL_1",
			envValue:     "true",
			defaultValue: false,
			setEnv:       true,
			want:         true,
		},
		{
			name:         "returns false when set to false",
			key:          "TEST_BOOL_2",
			envValue:     "false",
			defaultValue: true,
			setEnv:       true,
			want:         false,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_BOOL_3",
			envValue:     "",
			defaultValue: true,
			setEnv:       false,
			want:         true,
		},
		{
			name:         "returns default when invalid",
			key:          "TEST_BOOL_4",
			envValue:     "not_a_bool",
			defaultValue: false,
			setEnv:       true,
			want:         false,
		},
		{
			name:         "handles 1 as true",
			key:          "TEST_BOOL_5",
			envValue:     "1",
			defaultValue: false,
			setEnv:       true,
			want:         true,
		},
		{
			name:         "handles 0 as false",
			key:          "TEST_BOOL_6",
			envValue:     "0",
			defaultValue: true,
			setEnv:       true,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvBool(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvStringArray(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue []string
		setEnv       bool
		want         []string
	}{
		{
			name:         "returns parsed array",
			key:          "TEST_ARR_1",
			envValue:     `["a", "b", "c"]`,
			defaultValue: []string{"default"},
			setEnv:       true,
			want:         []string{"a", "b", "c"},
		},
		{
			name:         "returns default when not set",
			key:          "TEST_ARR_2",
			envValue:     "",
			defaultValue: []string{"x", "y"},
			setEnv:       false,
			want:         []string{"x", "y"},
		},
		{
			name:         "returns default when invalid json",
			key:          "TEST_ARR_3",
			envValue:     "not_json",
			defaultValue: []string{"fallback"},
			setEnv:       true,
			want:         []string{"fallback"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvStringArray(tt.key, tt.defaultValue)

			if len(got) != len(tt.want) {
				t.Errorf("getEnvStringArray() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("getEnvStringArray()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestGetEnvThinkValue(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue any
		setEnv       bool
		wantBool     bool
		wantString   string
		expectBool   bool
	}{
		{
			name:         "returns true when set to true",
			key:          "TEST_THINK_1",
			envValue:     "true",
			defaultValue: false,
			setEnv:       true,
			wantBool:     true,
			expectBool:   true,
		},
		{
			name:         "returns false when set to false",
			key:          "TEST_THINK_2",
			envValue:     "false",
			defaultValue: true,
			setEnv:       true,
			wantBool:     false,
			expectBool:   true,
		},
		{
			name:         "returns string when not bool",
			key:          "TEST_THINK_3",
			envValue:     "custom_value",
			defaultValue: false,
			setEnv:       true,
			wantString:   "custom_value",
			expectBool:   false,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_THINK_4",
			envValue:     "",
			defaultValue: true,
			setEnv:       false,
			wantBool:     true,
			expectBool:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv && tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvThinkValue(tt.key, tt.defaultValue)

			if tt.expectBool {
				if b, ok := got.(bool); ok {
					if b != tt.wantBool {
						t.Errorf("getEnvThinkValue() = %v, want %v", b, tt.wantBool)
					}
				} else {
					t.Errorf("getEnvThinkValue() expected bool, got %T", got)
				}
			} else {
				if s, ok := got.(string); ok {
					if s != tt.wantString {
						t.Errorf("getEnvThinkValue() = %q, want %q", s, tt.wantString)
					}
				} else {
					t.Errorf("getEnvThinkValue() expected string, got %T", got)
				}
			}
		})
	}
}
