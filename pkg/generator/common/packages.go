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

// ExtractPackageName extracts the Go package name from a protogen message.
//
// The Go package name is used for generating code in the correct package and
// for creating import statements. This function handles buf-managed packages
// that may have explicit package name overrides after a semicolon.
//
// Examples:
//   - "github.com/panyam/protoc-gen-dal/test/gen/dal;testdal" -> "testdal"
//   - "github.com/example/api/v1" -> "v1"
//   - "mypackage" -> "mypackage"
//
// Parameters:
//   - msg: A protogen message from which to extract the package name
//
// Returns:
//   - the Go package name
func ExtractPackageName(msg *protogen.Message) string {
	// GoImportPath format is "path/to/package" or "path/to/package;packagename"
	importPath := string(msg.GoIdent.GoImportPath)

	// Check if there's a package override (after semicolon)
	if idx := strings.LastIndex(importPath, ";"); idx != -1 {
		return importPath[idx+1:]
	}

	// Otherwise use the last part of the path
	if idx := strings.LastIndex(importPath, "/"); idx != -1 {
		return importPath[idx+1:]
	}

	return importPath
}

// GetPackageAlias returns the default alias for a package path.
//
// This is used when creating import statements with aliases. The alias is
// typically the last segment of the package path.
//
// Examples:
//   - GetPackageAlias("github.com/myapp/converters") -> "converters"
//   - GetPackageAlias("example.com/api/v1") -> "v1"
//   - GetPackageAlias("mypackage") -> "mypackage"
//
// Parameters:
//   - pkgPath: The full package import path
//
// Returns:
//   - the default alias (last segment of path)
func GetPackageAlias(pkgPath string) string {
	if idx := strings.LastIndex(pkgPath, "/"); idx != -1 {
		return pkgPath[idx+1:]
	}
	return pkgPath
}
