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

// Package styles provides the bundled markdownlint configuration presets
// (all, relaxed, prettier, cirosantilli) as parsed Configuration values.
package styles

import (
	_ "embed"
	"encoding/json"

	"github.com/ldmonster/go-markdownlint/internal/types"
)

//go:embed all.json
var allJSON []byte

//go:embed relaxed.json
var relaxedJSON []byte

//go:embed prettier.json
var prettierJSON []byte

//go:embed cirosantilli.json
var cirosantilliJSON []byte

func parse(b []byte) types.Configuration {
	var m map[string]interface{}

	_ = json.Unmarshal(b, &m)

	return types.ConfigFromMap(m)
}

// All returns the "all" style (every rule enabled).
func All() types.Configuration { return parse(allJSON) }

// Relaxed returns the "relaxed" style.
func Relaxed() types.Configuration { return parse(relaxedJSON) }

// Prettier returns the "prettier" style.
func Prettier() types.Configuration { return parse(prettierJSON) }

// Cirosantilli returns the "cirosantilli" style.
func Cirosantilli() types.Configuration { return parse(cirosantilliJSON) }
