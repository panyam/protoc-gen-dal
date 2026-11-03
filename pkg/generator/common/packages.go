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
	"github.com/panyam/protokit/packages"
	"google.golang.org/protobuf/compiler/protogen"
)

// ExtractPackageName extracts the Go package name from a protogen message.
// Delegates to protokit/packages.ExtractGoPackageName.
func ExtractPackageName(msg *protogen.Message) string {
	return packages.ExtractGoPackageName(msg)
}

// GetPackageAlias returns the default alias for a package path.
// Delegates to protokit/packages.GetPackageAlias.
func GetPackageAlias(pkgPath string) string {
	return packages.GetPackageAlias(pkgPath)
}
