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

func TestGenerateFilenameFromProto(t *testing.T) {
	tests := []struct {
		name      string
		protoPath string
		suffix    string
		expected  string
	}{
		{
			name:      "GORM file with path - preserves directory",
			protoPath: "gorm/user.proto",
			suffix:    "_gorm.go",
			expected:  "gorm/user_gorm.go",
		},
		{
			name:      "Datastore file with nested path - preserves directory",
			protoPath: "dal/v1/book.proto",
			suffix:    "_datastore.go",
			expected:  "dal/v1/book_datastore.go",
		},
		{
			name:      "Simple file with .go suffix - preserves directory",
			protoPath: "protos/author.proto",
			suffix:    ".go",
			expected:  "protos/author.go",
		},
		{
			name:      "File without directory",
			protoPath: "message.proto",
			suffix:    "_gorm.go",
			expected:  "message_gorm.go",
		},
		{
			name:      "Deep nested path - preserves full structure",
			protoPath: "a/b/c/d/entity.proto",
			suffix:    "_model.go",
			expected:  "a/b/c/d/entity_model.go",
		},
		{
			name:      "Multi-service collision case 1",
			protoPath: "likes/v1/gorm.proto",
			suffix:    "_gorm.go",
			expected:  "likes/v1/gorm_gorm.go",
		},
		{
			name:      "Multi-service collision case 2",
			protoPath: "tags/v1/gorm.proto",
			suffix:    "_gorm.go",
			expected:  "tags/v1/gorm_gorm.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFilenameFromProto(tt.protoPath, tt.suffix)
			if result != tt.expected {
				t.Errorf("GenerateFilenameFromProto(%q, %q) = %q; want %q",
					tt.protoPath, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestGenerateConverterFilename(t *testing.T) {
	tests := []struct {
		name      string
		protoPath string
		expected  string
	}{
		{
			name:      "GORM converter - preserves directory",
			protoPath: "gorm/user.proto",
			expected:  "gorm/user_converters.go",
		},
		{
			name:      "Datastore converter - preserves directory",
			protoPath: "dal/v1/book.proto",
			expected:  "dal/v1/book_converters.go",
		},
		{
			name:      "Simple converter - no directory",
			protoPath: "message.proto",
			expected:  "message_converters.go",
		},
		{
			name:      "Multi-service converter 1",
			protoPath: "likes/v1/gorm.proto",
			expected:  "likes/v1/gorm_converters.go",
		},
		{
			name:      "Multi-service converter 2",
			protoPath: "tags/v1/gorm.proto",
			expected:  "tags/v1/gorm_converters.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateConverterFilename(tt.protoPath)
			if result != tt.expected {
				t.Errorf("GenerateConverterFilename(%q) = %q; want %q",
					tt.protoPath, result, tt.expected)
			}
		})
	}
}

// TestGenerateFilenameFromProto_CollisionAvoidance tests that proto files with
// the same base name in different directories generate unique output filenames.
// This is critical for multi-service repos like goapplib where likes/v1/gorm.proto
// and tags/v1/gorm.proto would otherwise both generate "gorm_gorm.go".
func TestGenerateFilenameFromProto_CollisionAvoidance(t *testing.T) {
	// These proto paths should generate DIFFERENT output filenames
	collisionCases := []struct {
		protoPath1 string
		protoPath2 string
		suffix     string
	}{
		{
			protoPath1: "likes/v1/gorm.proto",
			protoPath2: "tags/v1/gorm.proto",
			suffix:     "_gorm.go",
		},
		{
			protoPath1: "service1/v1/models.proto",
			protoPath2: "service2/v1/models.proto",
			suffix:     "_gorm.go",
		},
		{
			protoPath1: "a/gae.proto",
			protoPath2: "b/gae.proto",
			suffix:     ".go",
		},
	}

	for _, tc := range collisionCases {
		result1 := GenerateFilenameFromProto(tc.protoPath1, tc.suffix)
		result2 := GenerateFilenameFromProto(tc.protoPath2, tc.suffix)

		if result1 == result2 {
			t.Errorf("COLLISION: %q and %q both generate %q - filenames must be unique",
				tc.protoPath1, tc.protoPath2, result1)
		}
	}
}

// Note: GroupMessagesByFile testing requires actual proto descriptors
// which is complex to mock. This function is tested indirectly via
// integration tests in the GORM and Datastore generators.
