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

	"github.com/panyam/protoc-gen-dal/pkg/collector"
)

// GroupMessagesByFile groups messages by their source proto file path.
//
// This function organizes messages based on which proto file they came from,
// allowing generators to create one output file per proto file. This maintains
// a logical organization where all messages from "dal/user.proto" go into a
// single output file.
//
// Parameters:
//   - messages: Collected messages from the collector
//
// Returns:
//   - map from proto file path to list of messages in that file
func GroupMessagesByFile(messages []*collector.MessageInfo) map[string][]*collector.MessageInfo {
	groups := make(map[string][]*collector.MessageInfo)
	for _, msg := range messages {
		// Get the proto file path from the target message
		protoFile := msg.TargetMessage.Desc.ParentFile().Path()
		groups[protoFile] = append(groups[protoFile], msg)
	}
	return groups
}

// GenerateFilenameFromProto creates an output filename from a proto file path.
//
// This extracts the base name from the proto file path and adds the specified
// suffix to create the output filename.
//
// Examples:
//   - GenerateFilenameFromProto("gorm/user.proto", "_gorm.go") -> "user_gorm.go"
//   - GenerateFilenameFromProto("dal/v1/book.proto", "_datastore.go") -> "book_datastore.go"
//   - GenerateFilenameFromProto("protos/author.proto", ".go") -> "author.go"
//
// Parameters:
//   - protoPath: Path to the proto file (e.g., "gorm/user.proto")
//   - suffix: Suffix to add to the base name (e.g., "_gorm.go", "_datastore.go")
//
// Returns:
//   - output filename with the specified suffix
func GenerateFilenameFromProto(protoPath, suffix string) string {
	// Extract base name without extension
	baseName := protoPath
	if idx := strings.LastIndex(baseName, "/"); idx != -1 {
		baseName = baseName[idx+1:]
	}
	if idx := strings.LastIndex(baseName, ".proto"); idx != -1 {
		baseName = baseName[:idx]
	}
	return baseName + suffix
}

// GenerateConverterFilename creates a converter filename from a proto file path.
//
// This is a convenience wrapper around GenerateFilenameFromProto that always
// uses "_converters.go" as the suffix.
//
// Examples:
//   - GenerateConverterFilename("gorm/user.proto") -> "user_converters.go"
//   - GenerateConverterFilename("dal/v1/book.proto") -> "book_converters.go"
//
// Parameters:
//   - protoPath: Path to the proto file
//
// Returns:
//   - converter filename with "_converters.go" suffix
func GenerateConverterFilename(protoPath string) string {
	return GenerateFilenameFromProto(protoPath, "_converters.go")
}
