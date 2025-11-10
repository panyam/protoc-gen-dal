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

	"google.golang.org/protobuf/compiler/protogen"
)

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		expected   string
	}{
		{
			name:       "Package with explicit override",
			importPath: "github.com/panyam/protoc-gen-dal/test/gen/dal;testdal",
			expected:   "testdal",
		},
		{
			name:       "Package without override",
			importPath: "github.com/example/api/v1",
			expected:   "v1",
		},
		{
			name:       "Simple package",
			importPath: "mypackage",
			expected:   "mypackage",
		},
		{
			name:       "Package with multiple segments",
			importPath: "example.com/foo/bar/baz",
			expected:   "baz",
		},
		{
			name:       "Package with override and deep path",
			importPath: "github.com/org/proj/pkg/subpkg;custom",
			expected:   "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock message with the specified import path
			msg := &protogen.Message{
				GoIdent: protogen.GoIdent{
					GoImportPath: protogen.GoImportPath(tt.importPath),
				},
			}

			result := ExtractPackageName(msg)
			if result != tt.expected {
				t.Errorf("ExtractPackageName(%q) = %q; want %q",
					tt.importPath, result, tt.expected)
			}
		})
	}
}

func TestGetPackageAlias(t *testing.T) {
	tests := []struct {
		name     string
		pkgPath  string
		expected string
	}{
		{
			name:     "Package with multiple segments",
			pkgPath:  "github.com/myapp/converters",
			expected: "converters",
		},
		{
			name:     "Package with version",
			pkgPath:  "example.com/api/v1",
			expected: "v1",
		},
		{
			name:     "Simple package",
			pkgPath:  "mypackage",
			expected: "mypackage",
		},
		{
			name:     "Deep nested package",
			pkgPath:  "domain.com/org/project/internal/utils",
			expected: "utils",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPackageAlias(tt.pkgPath)
			if result != tt.expected {
				t.Errorf("GetPackageAlias(%q) = %q; want %q",
					tt.pkgPath, result, tt.expected)
			}
		})
	}
}
