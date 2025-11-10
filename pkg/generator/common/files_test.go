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
			name:      "GORM file with path",
			protoPath: "gorm/user.proto",
			suffix:    "_gorm.go",
			expected:  "user_gorm.go",
		},
		{
			name:      "Datastore file with nested path",
			protoPath: "dal/v1/book.proto",
			suffix:    "_datastore.go",
			expected:  "book_datastore.go",
		},
		{
			name:      "Simple file with .go suffix",
			protoPath: "protos/author.proto",
			suffix:    ".go",
			expected:  "author.go",
		},
		{
			name:      "File without directory",
			protoPath: "message.proto",
			suffix:    "_gorm.go",
			expected:  "message_gorm.go",
		},
		{
			name:      "Deep nested path",
			protoPath: "a/b/c/d/entity.proto",
			suffix:    "_model.go",
			expected:  "entity_model.go",
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
			name:      "GORM converter",
			protoPath: "gorm/user.proto",
			expected:  "user_converters.go",
		},
		{
			name:      "Datastore converter",
			protoPath: "dal/v1/book.proto",
			expected:  "book_converters.go",
		},
		{
			name:      "Simple converter",
			protoPath: "message.proto",
			expected:  "message_converters.go",
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

// Note: GroupMessagesByFile testing requires actual proto descriptors
// which is complex to mock. This function is tested indirectly via
// integration tests in the GORM and Datastore generators.
