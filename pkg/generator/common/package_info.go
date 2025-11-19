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
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

// PackageInfo contains extracted information about a proto message's package.
type PackageInfo struct {
	// ImportPath is the clean import path without ;packagename suffix
	// Example: "github.com/test/gen/go/api"
	ImportPath string

	// Alias is the package alias for import statements
	// Example: "api" (from ".../go/api")
	Alias string
}

// ExtractPackageInfo extracts clean import path and package alias from a proto message.
// This is used when generating converters to properly import and reference message types.
//
// The function handles:
// - Stripping ;packagename suffix from GoImportPath
// - Extracting the last path component as the alias
//
// Example:
//   Input: message with GoImportPath = "github.com/test/gen/go/api;apipb"
//   Output: PackageInfo{ImportPath: "github.com/test/gen/go/api", Alias: "api"}
func ExtractPackageInfo(msg *protogen.Message) PackageInfo {
	if msg == nil {
		return PackageInfo{}
	}

	importPath := string(msg.GoIdent.GoImportPath)

	// Strip ;packagename suffix if present
	// protogen sometimes includes ";packagename" suffix which is not valid for imports
	if idx := strings.LastIndex(importPath, ";"); idx != -1 {
		importPath = importPath[:idx]
	}

	// Extract package alias from last path component
	alias := GetPackageAlias(importPath)

	return PackageInfo{
		ImportPath: importPath,
		Alias:      alias,
	}
}
