package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLIHello(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"hello"})
	if err := root.Execute(); err != nil {
		t.Fatalf("hello failed: %v", err)
	}
	if got, want := out.String(), "Hello, World!\n"; got != want {
		t.Errorf("hello output = %q, want %q", got, want)
	}

	out.Reset()
	root = NewRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"hello", "--name", "Alice"})
	if err := root.Execute(); err != nil {
		t.Fatalf("hello --name failed: %v", err)
	}
	if got, want := out.String(), "Hello, Alice!\n"; got != want {
		t.Errorf("hello --name output = %q, want %q", got, want)
	}
}

func TestCLIVersion(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(out.String(), "go-cli-tool-template ") {
		t.Errorf("unexpected version output: %q", out.String())
	}
}

func TestCLIHelpIncludesExamples(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"Examples:",
		"go-cli-tool-template hello",
		"go-cli-tool-template hello --name Alice",
		"go-cli-tool-template version",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("root help output missing %q:\n%s", want, got)
		}
	}

	out.Reset()
	root = NewRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"hello", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("hello --help failed: %v", err)
	}

	got = out.String()
	for _, want := range []string{
		"Examples:",
		"go-cli-tool-template hello",
		"go-cli-tool-template hello --name Alice",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("hello help output missing %q:\n%s", want, got)
		}
	}
}
