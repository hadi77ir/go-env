package env

import (
	"os"
	"testing"
)

func TestExpandEnvVars(t *testing.T) {
	// Setup test environment variables
	testEnvVars := map[string]string{
		"USER":        "testuser",
		"HOME":        "/home/testuser",
		"PATH":        "/usr/bin:/bin",
		"EMPTY":       "",
		"SHELL":       "/bin/bash",
		"TEST_VAR":    "test_value",
		"_UNDERSCORE": "underscore_value",
		"VAR123":      "mixed_alphanumeric",
	}

	// Set test environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	// Clean up after tests
	defer func() {
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
		// Clean up variables that might be set during tests
		os.Unsetenv("NEW_VAR")
		os.Unsetenv("ASSIGNED_VAR")
	}()

	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// Basic $var tests
		{
			name:    "simple variable substitution",
			args:    args{input: "Hello $USER"},
			want:    "Hello testuser",
			wantErr: false,
		},
		{
			name:    "variable at beginning",
			args:    args{input: "$USER is logged in"},
			want:    "testuser is logged in",
			wantErr: false,
		},
		{
			name:    "variable at end",
			args:    args{input: "Current user: $USER"},
			want:    "Current user: testuser",
			wantErr: false,
		},
		{
			name:    "multiple variables",
			args:    args{input: "$USER lives in $HOME"},
			want:    "testuser lives in /home/testuser",
			wantErr: false,
		},
		{
			name:    "undefined variable",
			args:    args{input: "Value: $UNDEFINED"},
			want:    "Value: ",
			wantErr: false,
		},
		{
			name:    "empty variable",
			args:    args{input: "Value: $EMPTY"},
			want:    "Value: ",
			wantErr: false,
		},

		// ${var} tests
		{
			name:    "braced variable",
			args:    args{input: "User: ${USER}"},
			want:    "User: testuser",
			wantErr: false,
		},
		{
			name:    "braced variable with text",
			args:    args{input: "${USER}name"},
			want:    "testusername",
			wantErr: false,
		},
		{
			name:    "empty braced variable",
			args:    args{input: "Value: ${}"},
			want:    "Value: ${}",
			wantErr: false,
		},
		{
			name:    "undefined braced variable",
			args:    args{input: "Value: ${UNDEFINED}"},
			want:    "Value: ",
			wantErr: false,
		},

		// ${var:-default} tests
		{
			name:    "default value with set variable",
			args:    args{input: "${USER:-defaultuser}"},
			want:    "testuser",
			wantErr: false,
		},
		{
			name:    "default value with unset variable",
			args:    args{input: "${UNSET:-defaultvalue}"},
			want:    "defaultvalue",
			wantErr: false,
		},
		{
			name:    "default value with empty variable",
			args:    args{input: "${EMPTY:-notempty}"},
			want:    "notempty",
			wantErr: false,
		},
		{
			name:    "default value complex",
			args:    args{input: "Shell: ${MISSING_SHELL:-/bin/sh}"},
			want:    "Shell: /bin/sh",
			wantErr: false,
		},

		// ${var:+alt} tests
		{
			name:    "alternative value with set variable",
			args:    args{input: "${USER:+user_exists}"},
			want:    "user_exists",
			wantErr: false,
		},
		{
			name:    "alternative value with unset variable",
			args:    args{input: "${UNSET:+should_not_appear}"},
			want:    "",
			wantErr: false,
		},
		{
			name:    "alternative value with empty variable",
			args:    args{input: "${EMPTY:+should_not_appear}"},
			want:    "",
			wantErr: false,
		},

		// ${var:=default} tests
		{
			name:    "assign default with set variable",
			args:    args{input: "${USER:=should_not_assign}"},
			want:    "testuser",
			wantErr: false,
		},
		{
			name:    "assign default with unset variable",
			args:    args{input: "${NEW_VAR:=assigned_value}"},
			want:    "assigned_value",
			wantErr: false,
		},

		// ${var:?error} tests
		{
			name:    "error with set variable",
			args:    args{input: "${USER:?should not error}"},
			want:    "testuser",
			wantErr: false,
		},
		{
			name:    "error with unset variable",
			args:    args{input: "${UNSET:?variable is required}"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "error with empty variable",
			args:    args{input: "${EMPTY:?variable cannot be empty}"},
			want:    "",
			wantErr: true,
		},

		// Variable name validation tests
		{
			name:    "invalid variable starting with digit",
			args:    args{input: "$123invalid"},
			want:    "$123invalid",
			wantErr: false,
		},
		{
			name:    "invalid braced variable starting with digit",
			args:    args{input: "${123invalid:-default}"},
			want:    "${123invalid:-default}",
			wantErr: false,
		},
		{
			name:    "invalid variable with special chars",
			args:    args{input: "${VAR-WITH-HYPHENS:-default}"},
			want:    "${VAR-WITH-HYPHENS:-default}",
			wantErr: false,
		},
		{
			name:    "valid variable starting with underscore",
			args:    args{input: "$_UNDERSCORE"},
			want:    "underscore_value",
			wantErr: false,
		},
		{
			name:    "valid variable with numbers",
			args:    args{input: "$VAR123"},
			want:    "mixed_alphanumeric",
			wantErr: false,
		},
		{
			name:    "variable name too long",
			args:    args{input: "${VERY_LONG_VARIABLE_NAME_THAT_IS_MORE_THAN_SIXTY_FOUR_CHARACTERS_LONG:-default}"},
			want:    "${VERY_LONG_VARIABLE_NAME_THAT_IS_MORE_THAN_SIXTY_FOUR_CHARACTERS_LONG:-default}",
			wantErr: false,
		},

		// Edge cases
		{
			name:    "dollar at end",
			args:    args{input: "Value ends with $"},
			want:    "Value ends with $",
			wantErr: false,
		},
		{
			name:    "double dollar",
			args:    args{input: "$$USER"},
			want:    "$testuser",
			wantErr: false,
		},
		{
			name:    "unclosed brace",
			args:    args{input: "${USER"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "no text",
			args:    args{input: ""},
			want:    "",
			wantErr: false,
		},
		{
			name:    "just dollar",
			args:    args{input: "$"},
			want:    "$",
			wantErr: false,
		},
		{
			name:    "complex mixed expression",
			args:    args{input: "$USER:${HOME}:${SHELL:-/bin/sh}:${UNDEFINED:-default}"},
			want:    "testuser:/home/testuser:/bin/bash:default",
			wantErr: false,
		},
		{
			name:    "nested braces",
			args:    args{input: "${PATH:-${HOME}/bin}"},
			want:    "/usr/bin:/bin",
			wantErr: false,
		},
		{
			name:    "multiple operators in text",
			args:    args{input: "User ${USER:+exists} in ${HOME:-/tmp} using ${SHELL:-sh}"},
			want:    "User exists in /home/testuser using /bin/bash",
			wantErr: false,
		},

		// Special character handling
		{
			name:    "variable followed by alphanumeric",
			args:    args{input: "${USER}123"},
			want:    "testuser123",
			wantErr: false,
		},
		{
			name:    "variable in path-like string",
			args:    args{input: "/usr/local/${USER}/bin"},
			want:    "/usr/local/testuser/bin",
			wantErr: false,
		},
		{
			name:    "colon in default value",
			args:    args{input: "${UNSET:-http://example.com}"},
			want:    "http://example.com",
			wantErr: false,
		},
		{
			name:    "equals in default value",
			args:    args{input: "${UNSET:-key=value}"},
			want:    "key=value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandEnv(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExpandEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExpandEnvVarsAssignment tests the assignment operator separately
// since it modifies environment state
func TestExpandEnvVarsAssignment(t *testing.T) {
	// Clean up before and after
	os.Unsetenv("ASSIGN_TEST")
	defer os.Unsetenv("ASSIGN_TEST")

	// Test assignment with unset variable
	result, err := ExpandEnv("${ASSIGN_TEST:=new_value}")
	if err != nil {
		t.Errorf("ExpandEnv() error = %v, expected no error", err)
	}
	if result != "new_value" {
		t.Errorf("ExpandEnv() got = %v, want = new_value", result)
	}

	// Verify the variable was actually set
	if os.Getenv("ASSIGN_TEST") != "new_value" {
		t.Errorf("Environment variable was not set correctly, got = %v, want = new_value", os.Getenv("ASSIGN_TEST"))
	}

	// Test assignment with already set variable (should not change)
	result2, err2 := ExpandEnv("${ASSIGN_TEST:=different_value}")
	if err2 != nil {
		t.Errorf("ExpandEnv() error = %v, expected no error", err2)
	}
	if result2 != "new_value" {
		t.Errorf("ExpandEnv() got = %v, want = new_value", result2)
	}

	// Verify the variable was not changed
	if os.Getenv("ASSIGN_TEST") != "new_value" {
		t.Errorf("Environment variable should not have changed, got = %v, want = new_value", os.Getenv("ASSIGN_TEST"))
	}
}

// BenchmarkExpandEnvVars provides performance benchmarks
func BenchmarkExpandEnvVars(b *testing.B) {
	os.Setenv("BENCH_VAR", "benchmark_value")
	defer os.Unsetenv("BENCH_VAR")

	testCases := []struct {
		name  string
		input string
	}{
		{"simple", "$BENCH_VAR"},
		{"braced", "${BENCH_VAR}"},
		{"default", "${MISSING:-default}"},
		{"complex", "$BENCH_VAR uses ${HOME:-/tmp} and ${SHELL:-/bin/sh}"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ExpandEnv(tc.input)
			}
		})
	}
}
