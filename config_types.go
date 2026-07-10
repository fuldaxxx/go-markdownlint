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

package markdownlint

import "github.com/fuldaxxx/go-markdownlint/internal/types"

// Per-rule configuration types, re-exported so callers can construct a
// Configuration with typed struct fields (e.g. Configuration{MD013: MD013Config{...}}).
// RuleConfig is the generic entry used for rules without typed options and for
// Custom (custom-rule / tag) keys.
type (
	RuleConfig = types.RuleConfig

	MD001Config = types.MD001Config
	MD003Config = types.MD003Config
	MD004Config = types.MD004Config
	MD007Config = types.MD007Config
	MD009Config = types.MD009Config
	MD010Config = types.MD010Config
	MD012Config = types.MD012Config
	MD013Config = types.MD013Config
	MD022Config = types.MD022Config
	MD024Config = types.MD024Config
	MD025Config = types.MD025Config
	MD026Config = types.MD026Config
	MD027Config = types.MD027Config
	MD029Config = types.MD029Config
	MD030Config = types.MD030Config
	MD031Config = types.MD031Config
	MD033Config = types.MD033Config
	MD035Config = types.MD035Config
	MD036Config = types.MD036Config
	MD040Config = types.MD040Config
	MD041Config = types.MD041Config
	MD043Config = types.MD043Config
	MD044Config = types.MD044Config
	MD046Config = types.MD046Config
	MD048Config = types.MD048Config
	MD049Config = types.MD049Config
	MD050Config = types.MD050Config
	MD051Config = types.MD051Config
	MD052Config = types.MD052Config
	MD053Config = types.MD053Config
	MD054Config = types.MD054Config
	MD055Config = types.MD055Config
	MD059Config = types.MD059Config
	MD060Config = types.MD060Config
)
