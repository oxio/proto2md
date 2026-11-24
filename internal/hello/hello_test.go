package hello

import "testing"

func TestGreeting(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "default", input: "", want: "Hello, World!"},
		{name: "whitespace", input: "   ", want: "Hello, World!"},
		{name: "custom name", input: "Go", want: "Hello, Go!"},
		{name: "trimmed name", input: "  Alice  ", want: "Hello, Alice!"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Greeting(tc.input); got != tc.want {
				t.Errorf("Greeting(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
