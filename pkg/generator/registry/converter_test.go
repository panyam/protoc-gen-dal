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

package registry

import (
	"testing"
)

func TestConverterRegistry_HasConverter(t *testing.T) {
	// Create a registry and manually populate it
	registry := &ConverterRegistry{
		converters: map[string]bool{
			"User:UserGORM":     true,
			"Book:BookGORM":     true,
			"Address:AddressGORM": true,
		},
	}

	tests := []struct {
		name       string
		sourceType string
		targetType string
		expected   bool
	}{
		{
			name:       "User converter exists",
			sourceType: "User",
			targetType: "UserGORM",
			expected:   true,
		},
		{
			name:       "Book converter exists",
			sourceType: "Book",
			targetType: "BookGORM",
			expected:   true,
		},
		{
			name:       "Address converter exists",
			sourceType: "Address",
			targetType: "AddressGORM",
			expected:   true,
		},
		{
			name:       "Non-existent converter",
			sourceType: "NonExistent",
			targetType: "NonExistentGORM",
			expected:   false,
		},
		{
			name:       "Wrong target type",
			sourceType: "User",
			targetType: "UserDatastore",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.HasConverter(tt.sourceType, tt.targetType)
			if result != tt.expected {
				t.Errorf("HasConverter(%q, %q) = %v; want %v",
					tt.sourceType, tt.targetType, result, tt.expected)
			}
		})
	}
}

func TestConverterRegistry_Count(t *testing.T) {
	tests := []struct {
		name       string
		converters map[string]bool
		expected   int
	}{
		{
			name:       "Empty registry",
			converters: map[string]bool{},
			expected:   0,
		},
		{
			name: "Single converter",
			converters: map[string]bool{
				"User:UserGORM": true,
			},
			expected: 1,
		},
		{
			name: "Multiple converters",
			converters: map[string]bool{
				"User:UserGORM":     true,
				"Book:BookGORM":     true,
				"Address:AddressGORM": true,
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &ConverterRegistry{
				converters: tt.converters,
			}

			result := registry.Count()
			if result != tt.expected {
				t.Errorf("Count() = %d; want %d", result, tt.expected)
			}
		})
	}
}

// Note: NewConverterRegistry testing requires actual protogen.Message structures
// which depend on complex proto descriptors that are difficult to mock.
// This function is tested indirectly via integration tests in the GORM
// and Datastore generators, which use real proto files.
