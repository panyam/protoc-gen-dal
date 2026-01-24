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
// This preserves the directory structure from the proto file path to ensure
// that files from different packages don't collide. The directory structure
// is maintained and only the .proto extension is replaced with the suffix.
//
// Examples:
//   - GenerateFilenameFromProto("gorm/user.proto", "_gorm.go") -> "gorm/user_gorm.go"
//   - GenerateFilenameFromProto("likes/v1/gorm.proto", "_gorm.go") -> "likes/v1/gorm_gorm.go"
//   - GenerateFilenameFromProto("tags/v1/gorm.proto", "_gorm.go") -> "tags/v1/gorm_gorm.go"
//   - GenerateFilenameFromProto("dal/v1/book.proto", "_datastore.go") -> "dal/v1/book_datastore.go"
//
// Parameters:
//   - protoPath: Path to the proto file (e.g., "likes/v1/gorm.proto")
//   - suffix: Suffix to add to the base name (e.g., "_gorm.go", "_datastore.go")
//
// Returns:
//   - output filename with directory structure preserved and specified suffix
func GenerateFilenameFromProto(protoPath, suffix string) string {
	// Remove .proto extension and add suffix, preserving directory structure
	result := protoPath
	if idx := strings.LastIndex(result, ".proto"); idx != -1 {
		result = result[:idx]
	}
	return result + suffix
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
