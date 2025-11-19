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

func TestExtractPackageInfo(t *testing.T) {
	tests := []struct {
		name       string
		importPath protogen.GoImportPath
		want       PackageInfo
	}{
		{
			name:       "path with ;packagename suffix",
			importPath: "github.com/test/gen/go/api;apipb",
			want: PackageInfo{
				ImportPath: "github.com/test/gen/go/api",
				Alias:      "api",
			},
		},
		{
			name:       "path without suffix",
			importPath: "github.com/test/gen/go/api",
			want: PackageInfo{
				ImportPath: "github.com/test/gen/go/api",
				Alias:      "api",
			},
		},
		{
			name:       "short path",
			importPath: "api",
			want: PackageInfo{
				ImportPath: "api",
				Alias:      "api",
			},
		},
		{
			name:       "nested path",
			importPath: "github.com/company/project/gen/go/service/v1;servicev1",
			want: PackageInfo{
				ImportPath: "github.com/company/project/gen/go/service/v1",
				Alias:      "v1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock message with the test import path
			msg := &protogen.Message{
				GoIdent: protogen.GoIdent{
					GoImportPath: tt.importPath,
				},
			}

			got := ExtractPackageInfo(msg)

			if got.ImportPath != tt.want.ImportPath {
				t.Errorf("ImportPath = %q, want %q", got.ImportPath, tt.want.ImportPath)
			}
			if got.Alias != tt.want.Alias {
				t.Errorf("Alias = %q, want %q", got.Alias, tt.want.Alias)
			}
		})
	}
}

func TestExtractPackageInfo_NilMessage(t *testing.T) {
	got := ExtractPackageInfo(nil)

	want := PackageInfo{}
	if got != want {
		t.Errorf("ExtractPackageInfo(nil) = %v, want %v", got, want)
	}
}
