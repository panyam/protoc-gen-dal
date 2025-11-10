// Copyright 2025 Sri Panyam
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

// ImportSpec represents a Go import with optional alias.
//
// Import specs are used when generating import statements in Go files.
// The alias is optional and used when the package name conflicts with
// local identifiers or when a shorter name is desired.
//
// Examples:
//   - ImportSpec{Path: "fmt"} -> import "fmt"
//   - ImportSpec{Alias: "pb", Path: "google.golang.org/protobuf"} -> import pb "google.golang.org/protobuf"
type ImportSpec struct {
	// Alias is the optional import alias (e.g., "models", "pb", "api")
	Alias string

	// Path is the full import path (e.g., "github.com/example/models")
	Path string
}

// ImportMap is a helper type for collecting unique imports.
//
// Use a map with the import path as key to automatically deduplicate imports.
// Convert to []ImportSpec when ready to render in templates.
type ImportMap map[string]ImportSpec

// Add adds an import to the map.
//
// If an import with the same path already exists, it keeps the existing one.
// This prevents duplicate imports.
//
// Parameters:
//   - spec: The import specification to add
func (m ImportMap) Add(spec ImportSpec) {
	if _, exists := m[spec.Path]; !exists {
		m[spec.Path] = spec
	}
}

// ToSlice converts the import map to a slice of import specs.
//
// This is useful for template rendering which typically expects slices.
//
// Returns:
//   - slice of import specs
func (m ImportMap) ToSlice() []ImportSpec {
	var result []ImportSpec
	for _, spec := range m {
		result = append(result, spec)
	}
	return result
}
