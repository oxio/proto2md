package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	templateRepo   = "oxio/go-cli-tool-template"
	templateBinary = "go-cli-tool-template"
	templateOwner  = "Oxio"

	readmeDescBegin    = "<!-- BEGIN TEMPLATE DESCRIPTION -->"
	readmeDescEnd      = "<!-- END TEMPLATE DESCRIPTION -->"
	readmeUsageBegin   = "<!-- BEGIN TEMPLATE USAGE -->"
	readmeUsageEnd     = "<!-- END TEMPLATE USAGE -->"
	taskBootstrapBegin = "  # BEGIN TEMPLATE BOOTSTRAP"
	taskBootstrapEnd   = "  # END TEMPLATE BOOTSTRAP"

	templateRootShort = `Short: "A template CLI tool written in Go",`
	templateRootLong  = "Long: `" + templateBinary + " is a starter template for building Go CLI tools\nwith Cobra, Taskfile, golangci-lint, and GitHub Actions releases.`,"
)

var (
	githubRemoteRe = regexp.MustCompile(`github\.com[:/]([^/\s]+/[^/\s]+)$`)
	validRepoRe    = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)
	validBinaryRe  = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	licenseLineRe  = regexp.MustCompile(`(?m)^Copyright \(c\) \d{4} .*$`)

	fetchRepoDescription = defaultFetchRepoDescription
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: get working directory: %v\n", err)
		os.Exit(1)
	}
	if err := run(os.Args[1:], cwd, os.Stdin, os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, rootDir string, stdin io.Reader, stdout io.Writer) error {
	fsFlags := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	fsFlags.SetOutput(stdout)

	var (
		repo          string
		binary        string
		description   string
		author        string
		interactive   bool
		keepBootstrap bool
	)

	fsFlags.StringVar(&repo, "repo", "", "GitHub repository in <owner>/<repo> format (default: detected from git remote 'origin')")
	fsFlags.StringVar(&binary, "binary", "", "CLI binary name (default: repo name from -repo)")
	fsFlags.StringVar(&description, "description", "", "Short description of the tool for README.md and cmd/root.go (default: fetched from GitHub repository)")
	fsFlags.StringVar(&author, "author", "", "Copyright holder for LICENSE (default: repository owner)")
	fsFlags.BoolVar(&interactive, "interactive", false, "Prompt interactively for project settings")
	fsFlags.BoolVar(&keepBootstrap, "keep-bootstrap", false, "Keep tools/bootstrap, the 'bootstrap' task in Taskfile.yml, and the template usage section in README.md")
	fsFlags.Usage = func() {
		_, _ = fmt.Fprintf(stdout, `Bootstrap a new repository created from %s.

Usage:
  go run ./tools/bootstrap [flags]
  task bootstrap [-- flags]

Flags:
`, templateBinary)
		fsFlags.PrintDefaults()
		_, _ = fmt.Fprint(stdout, `
Examples:
  task bootstrap
  go run ./tools/bootstrap -interactive
  go run ./tools/bootstrap -repo acme/my-cli -description "Convert foo files to bar"
  task bootstrap -- -interactive=false -repo acme/my-cli-repo -binary mycli -description "My CLI tool" -author "Acme Corp"
`)
	}

	if err := fsFlags.Parse(args); err != nil {
		return err
	}
	if fsFlags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(fsFlags.Args(), " "))
	}

	visited := make(map[string]bool)
	fsFlags.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	reader := bufio.NewReader(stdin)

	if !visited["repo"] {
		detected, err := detectRepoFromOrigin(rootDir)
		if interactive {
			defaultRepo := ""
			if err == nil && detected != templateRepo {
				defaultRepo = detected
			}
			repo, err = promptLine(reader, stdout, "GitHub repository (owner/repo)", defaultRepo)
			if err != nil {
				return fmt.Errorf("read repository: %w", err)
			}
		} else {
			if err != nil {
				return fmt.Errorf("could not detect GitHub owner/repo from git remote 'origin' (%w); pass -repo <owner>/<repo>", err)
			}
			repo = detected
		}
	}

	if !validRepoRe.MatchString(repo) {
		return fmt.Errorf("invalid repository format %q: expected <owner>/<repo>", repo)
	}

	owner, repoName, _ := strings.Cut(repo, "/")
	if !visited["binary"] {
		if interactive {
			var err error
			binary, err = promptLine(reader, stdout, "Binary name", repoName)
			if err != nil {
				return fmt.Errorf("read binary name: %w", err)
			}
		} else {
			binary = repoName
		}
	}

	if !validBinaryRe.MatchString(binary) {
		return fmt.Errorf("invalid binary name %q: allowed characters are letters, digits, '.', '_', '-'", binary)
	}

	if repo == templateRepo && binary == templateBinary {
		return fmt.Errorf("target repo and binary match the template itself (%s); run in a repository created from the template or pass -repo", templateRepo)
	}

	if !visited["description"] {
		defaultDesc := fetchRepoDescription(repo)
		if interactive {
			var err error
			description, err = promptLine(reader, stdout, "Short description (optional)", defaultDesc)
			if err != nil {
				return fmt.Errorf("read description: %w", err)
			}
		} else {
			description = defaultDesc
		}
	}
	description = strings.TrimSpace(description)

	defaultAuthor := owner
	if strings.EqualFold(owner, templateOwner) {
		defaultAuthor = templateOwner
	}
	if !visited["author"] {
		if interactive {
			var err error
			author, err = promptLine(reader, stdout, "License copyright holder", defaultAuthor)
			if err != nil {
				return fmt.Errorf("read author: %w", err)
			}
		} else {
			author = defaultAuthor
		}
	}
	author = strings.TrimSpace(author)
	if author == "" {
		author = defaultAuthor
	}

	_, _ = fmt.Fprintf(stdout, "Bootstrapping repository:\n  Repository  : %s\n  Binary      : %s\n", repo, binary)
	if description != "" {
		_, _ = fmt.Fprintf(stdout, "  Description : %s\n", description)
	}
	_, _ = fmt.Fprintf(stdout, "  Author      : %s\n", author)

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return fmt.Errorf("open root %s: %w", rootDir, err)
	}
	defer func() { _ = root.Close() }()

	if description != "" {
		if err := updateDescription(root, description); err != nil {
			return err
		}
	}

	if err := replaceInRepo(root, repo, binary); err != nil {
		return err
	}

	if err := updateLicense(root, author, time.Now().Year()); err != nil {
		return err
	}

	if !keepBootstrap {
		if err := replaceMarkedBlock(root, "Taskfile.yml", taskBootstrapBegin, taskBootstrapEnd, ""); err != nil {
			return err
		}
		if err := replaceMarkedBlock(root, "README.md", readmeUsageBegin, readmeUsageEnd, ""); err != nil {
			return err
		}
		if description == "" {
			if err := stripLines(root, "README.md", readmeDescBegin, readmeDescEnd); err != nil {
				return err
			}
		}
		if err := os.RemoveAll(filepath.Join(rootDir, "tools", "bootstrap")); err != nil {
			return fmt.Errorf("remove tools/bootstrap: %w", err)
		}
		// Remove tools/ directory if it is now empty.
		_ = os.Remove(filepath.Join(rootDir, "tools"))
	}

	_, _ = fmt.Fprint(stdout, `
Done! Next steps:
  1. Run 'task check' to verify formatting, linting, build, and tests.
  2. Replace cmd/hello.go and internal/hello/ with your own CLI commands.
  3. Review changes with 'git diff' and commit.
`)
	return nil
}

func promptLine(r *bufio.Reader, w io.Writer, label, defaultVal string) (string, error) {
	if defaultVal != "" {
		_, _ = fmt.Fprintf(w, "%s [%s]: ", label, defaultVal)
	} else {
		_, _ = fmt.Fprintf(w, "%s: ", label)
	}
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	val := strings.TrimSpace(line)
	if val == "" {
		return defaultVal, nil
	}
	return val, nil
}

func defaultFetchRepoDescription(repo string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 5 * time.Second}
	desc, err := fetchGitHubDescription(ctx, client, "https://api.github.com", repo, resolveGitHubToken())
	if err != nil {
		return ""
	}
	return desc
}

func resolveGitHubToken() string {
	for _, envVar := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if token := strings.TrimSpace(os.Getenv(envVar)); token != "" {
			return token
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func fetchGitHubDescription(ctx context.Context, client *http.Client, apiBaseURL, repo, token string) (string, error) {
	if !validRepoRe.MatchString(repo) {
		return "", fmt.Errorf("invalid repo %q", repo)
	}
	endpoint := strings.TrimRight(apiBaseURL, "/") + "/repos/" + repo
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "go-cli-tool-template-bootstrap")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var payload struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.Description), nil
}

func detectRepoFromOrigin(dir string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return parseGitHubRepo(strings.TrimSpace(string(out)))
}

func parseGitHubRepo(remoteURL string) (string, error) {
	cleaned := strings.TrimSuffix(strings.TrimSpace(remoteURL), ".git")
	matches := githubRemoteRe.FindStringSubmatch(cleaned)
	if len(matches) != 2 {
		return "", fmt.Errorf("unrecognized GitHub remote URL %q", remoteURL)
	}
	return matches[1], nil
}

func updateDescription(root *os.Root, description string) error {
	if err := replaceMarkedBlock(root, "README.md", readmeDescBegin, readmeDescEnd, description); err != nil {
		return err
	}

	quoted := strconv.Quote(description)
	rootCmdPath := "cmd/root.go"
	if err := replaceExactInFile(root, rootCmdPath, templateRootShort, "Short: "+quoted+","); err != nil {
		return err
	}
	if err := replaceExactInFile(root, rootCmdPath, templateRootLong, "Long: "+quoted+","); err != nil {
		return err
	}
	return formatGoFile(root, rootCmdPath)
}

func formatGoFile(root *os.Root, name string) error {
	data, err := root.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", name, err)
	}
	formatted, err := format.Source(data)
	if err != nil {
		return fmt.Errorf("format %s: %w", name, err)
	}
	if bytes.Equal(data, formatted) {
		return nil
	}
	info, err := root.Stat(name)
	if err != nil {
		return fmt.Errorf("stat %s: %w", name, err)
	}
	if writeErr := root.WriteFile(name, formatted, info.Mode().Perm()); writeErr != nil {
		return fmt.Errorf("write %s: %w", name, writeErr)
	}
	return nil
}

func replaceExactInFile(root *os.Root, name, oldText, newText string) error {
	data, err := root.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", name, err)
	}

	content := string(data)
	updated := strings.ReplaceAll(content, oldText, newText)
	if updated == content {
		return nil
	}

	info, err := root.Stat(name)
	if err != nil {
		return fmt.Errorf("stat %s: %w", name, err)
	}
	if writeErr := root.WriteFile(name, []byte(updated), info.Mode().Perm()); writeErr != nil {
		return fmt.Errorf("write %s: %w", name, writeErr)
	}
	return nil
}

func updateLicense(root *os.Root, author string, year int) error {
	data, err := root.ReadFile("LICENSE")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read LICENSE: %w", err)
	}

	replacement := fmt.Sprintf("Copyright (c) %d %s", year, author)
	updated := licenseLineRe.ReplaceAllString(string(data), replacement)
	if updated == string(data) {
		return nil
	}

	info, err := root.Stat("LICENSE")
	if err != nil {
		return fmt.Errorf("stat LICENSE: %w", err)
	}
	if writeErr := root.WriteFile("LICENSE", []byte(updated), info.Mode().Perm()); writeErr != nil {
		return fmt.Errorf("write LICENSE: %w", writeErr)
	}
	return nil
}

func replaceInRepo(root *os.Root, repo, binary string) error {
	replacer := strings.NewReplacer(
		templateRepo, repo,
		templateBinary, binary,
	)

	return fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch path {
			case ".git", ".idea", ".vscode", "bin", "tools/bootstrap":
				return fs.SkipDir
			}
			return nil
		}

		if path == "go.sum" || !d.Type().IsRegular() {
			return nil
		}

		data, readErr := root.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		if !bytes.Contains(data, []byte(templateBinary)) {
			return nil
		}

		updated := replacer.Replace(string(data))
		if updated == string(data) {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return fmt.Errorf("stat %s: %w", path, infoErr)
		}

		if writeErr := root.WriteFile(path, []byte(updated), info.Mode().Perm()); writeErr != nil {
			return fmt.Errorf("write %s: %w", path, writeErr)
		}
		return nil
	})
}

func replaceMarkedBlock(root *os.Root, name, beginMarker, endMarker, replacement string) error {
	data, err := root.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", name, err)
	}

	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	skip := false
	for scanner.Scan() {
		line := scanner.Text()
		switch line {
		case beginMarker:
			skip = true
			if replacement != "" {
				out.WriteString(replacement)
				out.WriteByte('\n')
			}
			continue
		case endMarker:
			skip = false
			continue
		}
		if !skip {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return fmt.Errorf("scan %s: %w", name, scanErr)
	}

	if bytes.Equal(data, out.Bytes()) {
		return nil
	}

	info, err := root.Stat(name)
	if err != nil {
		return fmt.Errorf("stat %s: %w", name, err)
	}

	if writeErr := root.WriteFile(name, out.Bytes(), info.Mode().Perm()); writeErr != nil {
		return fmt.Errorf("write %s: %w", name, writeErr)
	}
	return nil
}

func stripLines(root *os.Root, name string, targets ...string) error {
	data, err := root.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", name, err)
	}

	drop := make(map[string]bool, len(targets))
	for _, t := range targets {
		drop[t] = true
	}

	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if drop[line] {
			continue
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return fmt.Errorf("scan %s: %w", name, scanErr)
	}

	if bytes.Equal(data, out.Bytes()) {
		return nil
	}

	info, err := root.Stat(name)
	if err != nil {
		return fmt.Errorf("stat %s: %w", name, err)
	}

	if writeErr := root.WriteFile(name, out.Bytes(), info.Mode().Perm()); writeErr != nil {
		return fmt.Errorf("write %s: %w", name, writeErr)
	}
	return nil
}
