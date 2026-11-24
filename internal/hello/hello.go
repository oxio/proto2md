package hello

import (
	"fmt"
	"strings"
)

// Greeting returns a greeting message for the given name, defaulting to "World".
func Greeting(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "World"
	}
	return fmt.Sprintf("Hello, %s!", trimmed)
}
