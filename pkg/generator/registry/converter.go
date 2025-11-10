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
	"fmt"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"google.golang.org/protobuf/compiler/protogen"
)

// ConverterRegistry tracks which converter functions are being generated.
//
// This registry is used to determine if nested converter calls are available
// when generating field mappings. For example, if a User message contains an
// Address message field, the registry helps determine if an AddressToAddressGORM
// converter exists so we can generate the appropriate conversion code.
//
// The registry prevents generating broken code that calls non-existent converters.
type ConverterRegistry struct {
	converters map[string]bool // key: "SourceType:TargetType"
}

// StructNameFunc is a function that converts a proto message to a target struct name.
//
// Different targets use different naming conventions:
//   - GORM: buildStructName() transforms "UserGorm" -> "UserGORM"
//   - Datastore: uses raw message name "UserDatastore" as-is
//
// This function type allows the registry to work with any target generator.
type StructNameFunc func(*protogen.Message) string

// NewConverterRegistry creates a new converter registry from collected messages.
//
// The registry is built by examining which message pairs will have converters
// generated. For each message with a source, we create an entry mapping the
// source type name to the target type name.
//
// Parameters:
//   - messages: Collected messages from the collector (must have SourceMessage set)
//   - structNameFunc: Function to convert target messages to struct names
//
// Returns:
//   - A registry that can answer "does converter X->Y exist?" queries
//
// Example:
//   registry := NewConverterRegistry(messages, buildStructName)
//   if registry.HasConverter("User", "UserGORM") {
//       // Generate call to UserToUserGORM()
//   }
func NewConverterRegistry(messages []*collector.MessageInfo, structNameFunc StructNameFunc) *ConverterRegistry {
	reg := &ConverterRegistry{
		converters: make(map[string]bool),
	}

	for _, msg := range messages {
		// Skip messages without a source (embedded types don't have converters)
		if msg.SourceMessage == nil {
			continue
		}

		// Extract source type name (e.g., "User")
		sourceType := string(msg.SourceMessage.Desc.Name())

		// Extract target type name using provided function (e.g., "UserGORM", "UserDatastore")
		targetType := structNameFunc(msg.TargetMessage)

		// Store the converter mapping
		key := fmt.Sprintf("%s:%s", sourceType, targetType)
		reg.converters[key] = true
	}

	return reg
}

// HasConverter checks if a converter exists for the given source and target types.
//
// This is used during field mapping to determine if nested message conversions
// can be generated automatically.
//
// Parameters:
//   - sourceType: The source message type name (e.g., "User", "Address")
//   - targetType: The target struct name (e.g., "UserGORM", "AddressDatastore")
//
// Returns:
//   - true if a converter will be generated for this type pair
//
// Example:
//   if registry.HasConverter("Address", "AddressGORM") {
//       // Safe to generate: AddressToAddressGORM(src.Address, &dest.Address, nil)
//   }
func (r *ConverterRegistry) HasConverter(sourceType, targetType string) bool {
	key := fmt.Sprintf("%s:%s", sourceType, targetType)
	return r.converters[key]
}

// Count returns the total number of converters registered.
//
// This is useful for debugging and testing.
//
// Returns:
//   - the number of converter pairs in the registry
func (r *ConverterRegistry) Count() int {
	return len(r.converters)
}
