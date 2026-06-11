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

// Command markdownlint is a reference CLI for the go-markdownlint library.
//
// Usage:
//
//	markdownlint [flags] <file-or-glob>...
//
// Flags:
//
//	--config <file>   Configuration file (JSON/JSONC/YAML/TOML).
//	--fix             Apply fixes in place where possible.
//	--json            Emit results as JSON.
//	--style <name>    Built-in style: all|relaxed|prettier|cirosantilli.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	markdownlint "github.com/ldmonster/go-markdownlint"
	"github.com/ldmonster/go-markdownlint/configparse"
	"github.com/ldmonster/go-markdownlint/styles"
)

func main() {
	os.Exit(run())
}

func run() int {
	configFile := flag.String("config", "", "configuration file (JSON/JSONC/YAML/TOML)")
	style := flag.String("style", "", "built-in style: all|relaxed|prettier|cirosantilli")
	fix := flag.Bool("fix", false, "apply fixes in place where possible")
	asJSON := flag.Bool("json", false, "emit results as JSON")

	flag.Parse()

	cfg, err := resolveConfig(*configFile, *style)
	if err != nil {
		fmt.Fprintln(os.Stderr, "markdownlint:", err)
		return 2
	}

	files, err := expandGlobs(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "markdownlint:", err)
		return 2
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "markdownlint: no input files")
		return 2
	}

	results, err := markdownlint.Lint(context.Background(), markdownlint.Options{
		Config:        cfg,
		ConfigParsers: configparse.Common(),
		Files:         files,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "markdownlint:", err)
		return 2
	}

	if *fix {
		applyFixes(results)
	}

	count := 0
	for _, errs := range results {
		count += len(errs)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
	} else {
		fmt.Print(markdownlint.ResultsString(results))
	}

	if count > 0 {
		return 1
	}

	return 0
}

func resolveConfig(configFile, style string) (markdownlint.Configuration, error) {
	switch style {
	case "":
	case "all":
		return styles.All(), nil
	case "relaxed":
		return styles.Relaxed(), nil
	case "prettier":
		return styles.Prettier(), nil
	case "cirosantilli":
		return styles.Cirosantilli(), nil
	default:
		return markdownlint.Configuration{}, fmt.Errorf("unknown style %q", style)
	}

	if configFile != "" {
		return markdownlint.ReadConfig(configFile, configparse.Common())
	}

	deflt := true

	return markdownlint.Configuration{Default: &deflt}, nil
}

func expandGlobs(args []string) ([]string, error) {
	var files []string

	seen := map[string]bool{}

	for _, a := range args {
		matches, err := filepath.Glob(a)
		if err != nil {
			return nil, err
		}

		if matches == nil {
			matches = []string{a}
		}

		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				files = append(files, m)
			}
		}
	}

	sort.Strings(files)

	return files, nil
}

func applyFixes(results markdownlint.Results) {
	for file, errs := range results {
		hasFix := false

		for _, e := range errs {
			if e.FixInfo != nil {
				hasFix = true
				break
			}
		}

		if !hasFix {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		fixed := markdownlint.ApplyFixes(string(content), errs)
		_ = os.WriteFile(file, []byte(fixed), 0o644)
	}
}
