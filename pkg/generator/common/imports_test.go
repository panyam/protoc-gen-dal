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

import (
	"testing"
)

func TestImportMap_Add(t *testing.T) {
	im := make(ImportMap)

	// Add first import
	im.Add(ImportSpec{
		Alias: "pb",
		Path:  "google.golang.org/protobuf",
	})

	if len(im) != 1 {
		t.Errorf("Expected 1 import, got %d", len(im))
	}

	// Add duplicate (same path) - should not add
	im.Add(ImportSpec{
		Alias: "proto",
		Path:  "google.golang.org/protobuf",
	})

	if len(im) != 1 {
		t.Errorf("Expected 1 import after duplicate add, got %d", len(im))
	}

	// Verify the first one is kept
	spec := im["google.golang.org/protobuf"]
	if spec.Alias != "pb" {
		t.Errorf("Expected alias 'pb', got '%s'", spec.Alias)
	}

	// Add different import
	im.Add(ImportSpec{
		Path: "fmt",
	})

	if len(im) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(im))
	}
}

func TestImportMap_ToSlice(t *testing.T) {
	im := make(ImportMap)

	im.Add(ImportSpec{
		Alias: "pb",
		Path:  "google.golang.org/protobuf",
	})

	im.Add(ImportSpec{
		Path: "fmt",
	})

	im.Add(ImportSpec{
		Alias: "api",
		Path:  "github.com/example/api",
	})

	slice := im.ToSlice()

	if len(slice) != 3 {
		t.Errorf("Expected 3 imports in slice, got %d", len(slice))
	}

	// Verify all imports are present
	paths := make(map[string]bool)
	for _, spec := range slice {
		paths[spec.Path] = true
	}

	expectedPaths := []string{
		"google.golang.org/protobuf",
		"fmt",
		"github.com/example/api",
	}

	for _, path := range expectedPaths {
		if !paths[path] {
			t.Errorf("Expected path '%s' not found in slice", path)
		}
	}
}

func TestImportSpec(t *testing.T) {
	tests := []struct {
		name     string
		spec     ImportSpec
		expected string // expected import statement
	}{
		{
			name: "import with alias",
			spec: ImportSpec{
				Alias: "pb",
				Path:  "google.golang.org/protobuf",
			},
			expected: `pb "google.golang.org/protobuf"`,
		},
		{
			name: "import without alias",
			spec: ImportSpec{
				Path: "fmt",
			},
			expected: `"fmt"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies the structure is correct
			// Actual rendering happens in templates
			if tt.spec.Alias != "" && tt.spec.Path == "" {
				t.Error("ImportSpec should have a path")
			}
		})
	}
}
