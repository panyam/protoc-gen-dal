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

func TestProtoScalarToGo(t *testing.T) {
	tests := []struct {
		name      string
		protoType string
		expected  string
	}{
		{"string", "string", "string"},
		{"int32", "int32", "int32"},
		{"int64", "int64", "int64"},
		{"uint32", "uint32", "uint32"},
		{"uint64", "uint64", "uint64"},
		{"bool", "bool", "bool"},
		{"float", "float", "float32"},
		{"double", "double", "float64"},
		{"bytes", "bytes", "[]byte"},
		{"unknown", "unknown_type", "interface{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoScalarToGo(tt.protoType)
			if result != tt.expected {
				t.Errorf("ProtoScalarToGo(%q) = %q; want %q",
					tt.protoType, result, tt.expected)
			}
		})
	}
}

func TestIsNumericKind(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		expected bool
	}{
		{"int32", "int32", true},
		{"int64", "int64", true},
		{"uint32", "uint32", true},
		{"uint64", "uint64", true},
		{"sint32", "sint32", true},
		{"sint64", "sint64", true},
		{"fixed32", "fixed32", true},
		{"fixed64", "fixed64", true},
		{"sfixed32", "sfixed32", true},
		{"sfixed64", "sfixed64", true},
		{"float", "float", true},
		{"double", "double", true},
		{"string", "string", false},
		{"bool", "bool", false},
		{"bytes", "bytes", false},
		{"message", "message", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNumericKind(tt.kind)
			if result != tt.expected {
				t.Errorf("IsNumericKind(%q) = %v; want %v",
					tt.kind, result, tt.expected)
			}
		})
	}
}

func TestProtoKindToGoType(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		expected string
	}{
		{"int32", "int32", "int32"},
		{"sint32", "sint32", "int32"},
		{"sfixed32", "sfixed32", "int32"},
		{"int64", "int64", "int64"},
		{"sint64", "sint64", "int64"},
		{"sfixed64", "sfixed64", "int64"},
		{"uint32", "uint32", "uint32"},
		{"fixed32", "fixed32", "uint32"},
		{"uint64", "uint64", "uint64"},
		{"fixed64", "fixed64", "uint64"},
		{"float", "float", "float32"},
		{"double", "double", "float64"},
		{"string", "string", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProtoKindToGoType(tt.kind)
			if result != tt.expected {
				t.Errorf("ProtoKindToGoType(%q) = %q; want %q",
					tt.kind, result, tt.expected)
			}
		})
	}
}

// Note: ProtoFieldToGoType testing requires actual protogen.Field structures
// which depend on complex proto descriptors that are difficult to mock.
// This function is tested indirectly via integration tests in the GORM
// and Datastore generators, which use real proto files.
