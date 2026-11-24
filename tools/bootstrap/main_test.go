package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGitHubRepo(t *testing.T) {
	tests := []struct {
		url     string
		want    string
		wantErr bool
	}{
		{url: "git@github.com:acme/mytool.git", want: "acme/mytool"},
		{url: "git@github.com:acme/mytool", want: "acme/mytool"},
		{url: "ssh://git@github.com/acme/mytool.git", want: "acme/mytool"},
		{url: "https://github.com/acme/mytool.git", want: "acme/mytool"},
		{url: "https://github.com/acme/mytool", want: "acme/mytool"},
		{url: "https://gitlab.com/acme/mytool.git", wantErr: true},
		{url: "", wantErr: true},
	}

	for _, tc := range tests {
		got, err := parseGitHubRepo(tc.url)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseGitHubRepo(%q) expected error, got %q", tc.url, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseGitHubRepo(%q) unexpected error: %v", tc.url, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseGitHubRepo(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestRunReplacesAndCleansUp(t *testing.T) {
	tmp := t.TempDir()

	writeFile(t, filepath.Join(tmp, "go.mod"), "module github.com/oxio/go-cli-tool-template\n", 0o600)
	writeFile(t, filepath.Join(tmp, "install.sh"), "#!/bin/bash\nGITHUB_REPO=\"oxio/go-cli-tool-template\"\nBINARY_NAME=\"go-cli-tool-template\"\n", 0o755)
	writeFile(t, filepath.Join(tmp, "LICENSE"), "MIT License\n\nCopyright (c) 2026 Oxio\n", 0o600)
	writeFile(t, filepath.Join(tmp, "README.md"), `# go-cli-tool-template

<!-- BEGIN TEMPLATE DESCRIPTION -->
Any template description text here.
<!-- END TEMPLATE DESCRIPTION -->

<!-- BEGIN TEMPLATE USAGE -->
## Using This Template

Run task bootstrap.

<!-- END TEMPLATE USAGE -->
## Installation

go install github.com/oxio/go-cli-tool-template@latest
go-cli-tool-template hello
`, 0o600)
	writeFile(t, filepath.Join(tmp, "cmd", "root.go"), `package cmd

func f() {
	_ = struct {
		Use   string
		Short string
		Long  string
	}{
		Use:   "go-cli-tool-template",
		Short: "A template CLI tool written in Go",
		Long: `+"`"+`go-cli-tool-template is a starter template for building Go CLI tools
with Cobra, Taskfile, golangci-lint, and GitHub Actions releases.`+"`"+`,
	}
}
`, 0o600)
	writeFile(t, filepath.Join(tmp, "Taskfile.yml"), `version: '3'
vars:
  BINARY: bin/go-cli-tool-template
tasks:
  # BEGIN TEMPLATE BOOTSTRAP
  bootstrap:
    interactive: true
    cmds:
      - go run ./tools/bootstrap -interactive {{.CLI_ARGS}}

  # END TEMPLATE BOOTSTRAP
  build:
    cmds:
      - go build .
`, 0o600)
	writeFile(t, filepath.Join(tmp, "tools", "bootstrap", "main.go"), "package main\n", 0o600)

	var out bytes.Buffer
	err := run([]string{
		"-repo", "acme/my-repo",
		"-binary", "mycli",
		"-description", "Convert widgets into gadgets",
		"-author", "Acme Corp",
	}, tmp, strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	gotGoMod := readFile(t, filepath.Join(tmp, "go.mod"))
	if gotGoMod != "module github.com/acme/my-repo\n" {
		t.Errorf("go.mod = %q", gotGoMod)
	}

	gotInstall := readFile(t, filepath.Join(tmp, "install.sh"))
	if !strings.Contains(gotInstall, `GITHUB_REPO="acme/my-repo"`) || !strings.Contains(gotInstall, `BINARY_NAME="mycli"`) {
		t.Errorf("unexpected install.sh:\n%s", gotInstall)
	}
	info, err := os.Stat(filepath.Join(tmp, "install.sh"))
	if err != nil {
		t.Fatalf("stat install.sh: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("install.sh lost executable bit: %v", info.Mode())
	}

	gotReadme := readFile(t, filepath.Join(tmp, "README.md"))
	if strings.Contains(gotReadme, "go-cli-tool-template") || strings.Contains(gotReadme, "Using This Template") || strings.Contains(gotReadme, "BEGIN TEMPLATE") || strings.Contains(gotReadme, "Any template description") {
		t.Errorf("README.md still has template content:\n%s", gotReadme)
	}
	for _, want := range []string{
		"# mycli",
		"Convert widgets into gadgets",
		"go install github.com/acme/my-repo@latest",
		"mycli hello",
	} {
		if !strings.Contains(gotReadme, want) {
			t.Errorf("README.md missing %q:\n%s", want, gotReadme)
		}
	}

	gotRoot := readFile(t, filepath.Join(tmp, "cmd", "root.go"))
	for _, want := range []string{
		`Use:   "mycli",`,
		`Short: "Convert widgets into gadgets",`,
		`Long:  "Convert widgets into gadgets",`,
	} {
		if !strings.Contains(gotRoot, want) {
			t.Errorf("cmd/root.go missing %q:\n%s", want, gotRoot)
		}
	}

	gotLicense := readFile(t, filepath.Join(tmp, "LICENSE"))
	if !strings.Contains(gotLicense, "Acme Corp") || strings.Contains(gotLicense, "Oxio") {
		t.Errorf("unexpected LICENSE:\n%s", gotLicense)
	}

	gotTaskfile := readFile(t, filepath.Join(tmp, "Taskfile.yml"))
	if strings.Contains(gotTaskfile, "bootstrap:") || strings.Contains(gotTaskfile, "BEGIN TEMPLATE BOOTSTRAP") {
		t.Errorf("Taskfile.yml still contains bootstrap task:\n%s", gotTaskfile)
	}
	if !strings.Contains(gotTaskfile, "BINARY: bin/mycli") {
		t.Errorf("Taskfile.yml binary not updated:\n%s", gotTaskfile)
	}

	if _, err := os.Stat(filepath.Join(tmp, "tools")); !os.IsNotExist(err) {
		t.Errorf("expected tools/ to be removed, got err=%v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestFetchGitHubDescription(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/repos/acme/spark" {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     "404 Not Found",
					Body:       io.NopCloser(strings.NewReader(`{}`)),
					Header:     make(http.Header),
				}, nil
			}
			if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
				t.Errorf("Authorization header = %q, want %q", got, "Bearer secret-token")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(`{"description":"  Fast Spark CLI  "}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	got, err := fetchGitHubDescription(context.Background(), client, "https://api.github.com", "acme/spark", "secret-token")
	if err != nil {
		t.Fatalf("fetchGitHubDescription failed: %v", err)
	}
	if got != "Fast Spark CLI" {
		t.Errorf("got %q, want %q", got, "Fast Spark CLI")
	}

	if _, err := fetchGitHubDescription(context.Background(), client, "https://api.github.com", "acme/missing", ""); err == nil {
		t.Errorf("expected error for 404 response")
	}
}

func TestRunInteractive(t *testing.T) {
	origFetch := fetchRepoDescription
	t.Cleanup(func() { fetchRepoDescription = origFetch })
	fetchRepoDescription = func(string) string { return "Fetched from GitHub" }

	tmp := t.TempDir()

	writeFile(t, filepath.Join(tmp, "go.mod"), "module github.com/oxio/go-cli-tool-template\n", 0o600)
	writeFile(t, filepath.Join(tmp, "LICENSE"), "MIT License\n\nCopyright (c) 2026 Oxio\n", 0o600)
	writeFile(t, filepath.Join(tmp, "README.md"), "# go-cli-tool-template\n\n<!-- BEGIN TEMPLATE DESCRIPTION -->\nOld template description.\n<!-- END TEMPLATE DESCRIPTION -->\n", 0o600)

	// Enter repo, accept default binary (Enter), accept fetched description (Enter), enter custom author.
	input := "acme/spark\n\n\nAcme Inc\n"
	var out bytes.Buffer
	if err := run([]string{"-interactive"}, tmp, strings.NewReader(input), &out); err != nil {
		t.Fatalf("interactive run failed: %v", err)
	}

	gotOut := out.String()
	for _, prompt := range []string{
		"GitHub repository (owner/repo):",
		"Binary name [spark]:",
		"Short description (optional) [Fetched from GitHub]:",
		"License copyright holder [acme]:",
	} {
		if !strings.Contains(gotOut, prompt) {
			t.Errorf("interactive output missing prompt %q:\n%s", prompt, gotOut)
		}
	}

	if got := readFile(t, filepath.Join(tmp, "go.mod")); got != "module github.com/acme/spark\n" {
		t.Errorf("go.mod = %q", got)
	}
	if got := readFile(t, filepath.Join(tmp, "README.md")); !strings.Contains(got, "# spark") || !strings.Contains(got, "Fetched from GitHub") {
		t.Errorf("unexpected README.md:\n%s", got)
	}
	if got := readFile(t, filepath.Join(tmp, "LICENSE")); !strings.Contains(got, "Acme Inc") {
		t.Errorf("unexpected LICENSE:\n%s", got)
	}
}

func TestRunRefusesTemplateRepo(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"-repo", templateRepo}, t.TempDir(), strings.NewReader(""), &out)
	if err == nil || !strings.Contains(err.Error(), "match the template itself") {
		t.Fatalf("expected template guard error, got %v", err)
	}
}

func writeFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
