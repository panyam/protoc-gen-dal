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
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// CollectCustomConverterImports scans a message's fields for custom converter functions
// (to_func/from_func in ColumnOptions) and adds their import paths to the imports map.
//
// This is generic Go import management - nothing target-specific. Used by both
// GORM and Datastore generators to ensure custom converter packages are imported.
//
// Example:
//   Field with: to_func: {package: "github.com/myapp/converters", function: "ToMillis"}
//   Adds: ImportSpec{Path: "github.com/myapp/converters", Alias: "converters"}
func CollectCustomConverterImports(msg *protogen.Message, imports ImportMap) {
	if msg == nil {
		return
	}

	for _, field := range msg.Fields {
		opts := field.Desc.Options()
		if opts == nil {
			continue
		}

		v := proto.GetExtension(opts, dalv1.E_Column)
		if v == nil {
			continue
		}

		colOpts, ok := v.(*dalv1.ColumnOptions)
		if !ok || colOpts == nil {
			continue
		}

		// Add to_func package import
		if colOpts.ToFunc != nil && colOpts.ToFunc.Package != "" {
			pkgPath := colOpts.ToFunc.Package
			pkgAlias := colOpts.ToFunc.Alias
			if pkgAlias == "" {
				pkgAlias = GetPackageAlias(pkgPath)
			}
			imports.Add(ImportSpec{Alias: pkgAlias, Path: pkgPath})
		}

		// Add from_func package import
		if colOpts.FromFunc != nil && colOpts.FromFunc.Package != "" {
			pkgPath := colOpts.FromFunc.Package
			pkgAlias := colOpts.FromFunc.Alias
			if pkgAlias == "" {
				pkgAlias = GetPackageAlias(pkgPath)
			}
			imports.Add(ImportSpec{Alias: pkgAlias, Path: pkgPath})
		}
	}
}

// ExtractCustomConverters extracts custom converter code from a field's column options.
// Returns the to_func and from_func conversion code if specified.
//
// This is used during field mapping to apply user-defined converter functions.
// Generic utility - works for any target (GORM, Datastore, etc.).
//
// Example:
//   Field with: to_func: {package: "conv", alias: "c", function: "ToMillis"}
//   Returns: toCode = "c.ToMillis(src.FieldName)", fromCode = ""
func ExtractCustomConverters(field *protogen.Field, fieldName string) (toTargetCode, fromTargetCode string) {
	opts := field.Desc.Options()
	if opts == nil {
		return "", ""
	}

	// Get column options
	v := proto.GetExtension(opts, dalv1.E_Column)
	if v == nil {
		return "", ""
	}

	colOpts, ok := v.(*dalv1.ColumnOptions)
	if !ok || colOpts == nil {
		return "", ""
	}

	// Extract to_func
	if colOpts.ToFunc != nil && colOpts.ToFunc.Function != "" {
		pkgAlias := colOpts.ToFunc.Alias
		if pkgAlias == "" {
			// Use last segment of package path as alias
			pkgAlias = GetPackageAlias(colOpts.ToFunc.Package)
		}
		toTargetCode = pkgAlias + "." + colOpts.ToFunc.Function + "(src." + fieldName + ")"
	}

	// Extract from_func
	if colOpts.FromFunc != nil && colOpts.FromFunc.Function != "" {
		pkgAlias := colOpts.FromFunc.Alias
		if pkgAlias == "" {
			pkgAlias = GetPackageAlias(colOpts.FromFunc.Package)
		}
		fromTargetCode = pkgAlias + "." + colOpts.FromFunc.Function + "(src." + fieldName + ")"
	}

	return toTargetCode, fromTargetCode
}
