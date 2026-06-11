// Copyright 2026 go-markdownlint Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// End-to-end tests for the markdownlint CLI.
//
// These build the CLI binary once (in TestMain) and then invoke it as a
// subprocess against temporary Markdown files, asserting on stdout/stderr and
// the process exit code. Because the binary runs out-of-process, these tests
// do NOT contribute to `go test -cover` statement coverage of main.go; that is
// expected. The value here is behavioral end-to-end verification of the actual
// compiled CLI.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binPath is the path to the CLI binary built once by TestMain.
var binPath string

func TestMain(m *testing.M) {
	// Skip the whole suite gracefully if the Go toolchain is unavailable, since
	// we need it to build the binary under test.
	goTool, err := exec.LookPath("go")
	if err != nil {
		// Cannot t.Skip from TestMain; print a notice and exit 0 so the run is
		// not considered a failure when `go` is not on PATH.
		os.Stdout.WriteString("SKIP: go not found on PATH; skipping CLI e2e tests\n")
		os.Exit(0)
	}

	tmp, err := os.MkdirTemp("", "mdl-e2e-bin")
	if err != nil {
		panic("mkdtemp: " + err.Error())
	}

	binPath = filepath.Join(tmp, "mdl")
	if os.PathSeparator == '\\' { // Windows safety, though tests target linux.
		binPath += ".exe"
	}

	// Build the CLI once. "." refers to the package in the current working
	// directory, which is the directory of this test file (cmd/markdownlint).
	build := exec.Command(goTool, "build", "-o", binPath, ".")
	build.Env = os.Environ()
	if out, err := build.CombinedOutput(); err != nil {
		os.Stderr.WriteString("build failed: " + err.Error() + "\n" + string(out) + "\n")
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// runResult holds the outcome of a single CLI invocation.
type runResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runCLI runs the built binary with the given args in dir and returns the
// captured output and exit code.
func runCLI(t *testing.T, dir string, args ...string) runResult {
	t.Helper()
	if binPath == "" {
		t.Skip("CLI binary not built (go unavailable)")
	}
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	code := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else {
			t.Fatalf("running CLI %v: %v", args, err)
		}
	}
	return runResult{stdout: stdout.String(), stderr: stderr.String(), exitCode: code}
}

// writeFile writes content to name inside dir and returns the full path.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestCLI_FileWithViolations(t *testing.T) {
	dir := t.TempDir()
	// Line 1 has a single trailing space (MD009); the file does not end with a
	// trailing newline (MD047).
	writeFile(t, dir, "bad.md", "# Title \n\ntrailing  \nno final newline")

	res := runCLI(t, dir, "bad.md")

	if res.exitCode != 1 {
		t.Fatalf("exit code = %d, want 1 (stderr=%q stdout=%q)", res.exitCode, res.stderr, res.stdout)
	}
	for _, want := range []string{"MD009", "MD047"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("stdout missing %q; got:\n%s", want, res.stdout)
		}
	}
}

func TestCLI_CleanFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "clean.md", "# Title\n\nText.\n")
	// Disable all default rules so nothing fires regardless of formatting.
	writeFile(t, dir, "off.json", `{"default": false}`)

	res := runCLI(t, dir, "--config", "off.json", "clean.md")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected no stdout output, got:\n%s", res.stdout)
	}
	if res.stderr != "" {
		t.Errorf("expected no stderr output, got:\n%s", res.stderr)
	}
}

func TestCLI_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	badPath := writeFile(t, dir, "bad.md", "# Title\n\ntrailing \n")

	res := runCLI(t, dir, "--json", "bad.md")

	// JSON mode still exits 1 because there are violations.
	if res.exitCode != 1 {
		t.Fatalf("exit code = %d, want 1 (stderr=%q)", res.exitCode, res.stderr)
	}

	// Results is map[string][]Error; unmarshal into a generic map and verify
	// structure.
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(res.stdout), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, res.stdout)
	}

	// The key should be the file argument as passed ("bad.md"); fall back to any
	// key referencing the file path.
	var findings []interface{}
	if v, ok := parsed["bad.md"]; ok {
		findings, _ = v.([]interface{})
	} else {
		for k, v := range parsed {
			if strings.Contains(k, "bad.md") || k == badPath {
				findings, _ = v.([]interface{})
			}
		}
	}
	if len(findings) == 0 {
		t.Fatalf("expected at least one finding in JSON, got: %s", res.stdout)
	}

	first, ok := findings[0].(map[string]interface{})
	if !ok {
		t.Fatalf("first finding is not an object: %#v", findings[0])
	}
	for _, field := range []string{"lineNumber", "ruleNames", "ruleDescription"} {
		if _, ok := first[field]; !ok {
			t.Errorf("finding missing field %q: %#v", field, first)
		}
	}
	if names, ok := first["ruleNames"].([]interface{}); !ok || len(names) == 0 {
		t.Errorf("ruleNames is not a non-empty array: %#v", first["ruleNames"])
	}
}

func TestCLI_StyleRelaxed(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "doc.md", "# Title\n\nText.\n")

	res := runCLI(t, dir, "--style", "relaxed", "doc.md")

	// Should run without an internal error (exit code 0 or 1, not 2).
	if res.exitCode == 2 {
		t.Fatalf("--style relaxed errored: exit=2 stderr=%q", res.stderr)
	}
	if res.stderr != "" {
		t.Errorf("unexpected stderr: %q", res.stderr)
	}
}

func TestCLI_ConfigDisablesRule(t *testing.T) {
	dir := t.TempDir()
	// Single trailing space => MD009 violation.
	writeFile(t, dir, "td.md", "trailing \n")

	// With MD009 enabled the rule fires (exit 1).
	writeFile(t, dir, "on.json", `{"default": false, "MD009": true}`)
	on := runCLI(t, dir, "--config", "on.json", "td.md")
	if on.exitCode != 1 {
		t.Fatalf("with MD009 enabled: exit=%d want 1 (stderr=%q stdout=%q)", on.exitCode, on.stderr, on.stdout)
	}
	if !strings.Contains(on.stdout, "MD009") {
		t.Errorf("expected MD009 in output, got:\n%s", on.stdout)
	}

	// Disabling MD009 in config suppresses the finding (exit 0, no output).
	writeFile(t, dir, "off.json", `{"default": false, "MD009": false}`)
	off := runCLI(t, dir, "--config", "off.json", "td.md")
	if off.exitCode != 0 {
		t.Fatalf("with MD009 disabled: exit=%d want 0 (stderr=%q stdout=%q)", off.exitCode, off.stderr, off.stdout)
	}
	if strings.Contains(off.stdout, "MD009") {
		t.Errorf("MD009 should be suppressed, got:\n%s", off.stdout)
	}
}

func TestCLI_FixRewritesInPlace(t *testing.T) {
	dir := t.TempDir()
	// Trailing spaces on line 1 (MD009) and a missing final newline (MD047).
	path := writeFile(t, dir, "fix.md", "# Title \n\ntrailing  \nno newline")

	// First run with --fix: it reports the original violations and rewrites the
	// file; exit code reflects the violations found (1).
	first := runCLI(t, dir, "--fix", "fix.md")
	if first.exitCode != 1 {
		t.Fatalf("first --fix run: exit=%d want 1 (stderr=%q)", first.exitCode, first.stderr)
	}

	fixed, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("re-read fixed file: %v", err)
	}
	content := string(fixed)
	if !strings.HasSuffix(content, "\n") {
		t.Errorf("expected file to end with a newline after --fix, got %q", content)
	}
	if strings.HasSuffix(content, " \n") || strings.Contains(content, "Title \n") {
		t.Errorf("expected single trailing spaces to be fixed, got %q", content)
	}

	// Second run on the now-fixed file: nothing left to report, exit 0.
	second := runCLI(t, dir, "fix.md")
	if second.exitCode != 0 {
		t.Fatalf("second run after fix: exit=%d want 0 (stdout=%q stderr=%q)", second.exitCode, second.stdout, second.stderr)
	}
	if second.stdout != "" {
		t.Errorf("expected clean output after fix, got:\n%s", second.stdout)
	}
}

func TestCLI_UnknownStyle(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "doc.md", "# Title\n\nText.\n")

	res := runCLI(t, dir, "--style", "xyz", "doc.md")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2 (stderr=%q)", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "unknown style") {
		t.Errorf("expected 'unknown style' on stderr, got: %q", res.stderr)
	}
}

func TestCLI_NoInputFiles(t *testing.T) {
	dir := t.TempDir()

	res := runCLI(t, dir)

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2 (stderr=%q)", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "no input files") {
		t.Errorf("expected 'no input files' on stderr, got: %q", res.stderr)
	}
}
