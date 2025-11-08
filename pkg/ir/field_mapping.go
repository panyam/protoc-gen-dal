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

package ir

import (
	dalv1 "github.com/panyam/protoc-gen-dal/proto/gen/dal/v1"
	"google.golang.org/protobuf/compiler/protogen"
)

// FieldMapping maps a source proto field to a target database column
type FieldMapping struct {
	// SourceField is the field from the API proto message
	SourceField *protogen.Field

	// TargetField is the field from the DAL schema message
	TargetField *protogen.Field

	// ColumnName is the database column name
	ColumnName string

	// ColumnType is the database-specific column type (e.g., "UUID", "JSONB")
	ColumnType string

	// IsPrimaryKey indicates if this is a primary key column
	IsPrimaryKey bool

	// IsNullable indicates if the column accepts NULL values
	IsNullable bool

	// DefaultValue is the default value expression
	DefaultValue string

	// IsUnique indicates if this column has a unique constraint
	IsUnique bool

	// AutoIncrement indicates if this column is auto-incremented
	AutoIncrement bool

	// AutoCreateTime indicates if this field should be set on creation
	AutoCreateTime bool

	// AutoUpdateTime indicates if this field should be updated on modification
	AutoUpdateTime bool

	// StorageStrategy defines how repeated/map fields are stored
	StorageStrategy StorageStrategy

	// NeedsTransformation indicates if this field needs custom transformation
	NeedsTransformation bool

	// TransformationComment provides a hint about what transformation is needed
	TransformationComment string
}

// StorageStrategy defines how complex types are stored in the database
type StorageStrategy int

const (
	// StorageAuto lets the generator decide based on field type and target
	StorageAuto StorageStrategy = iota

	// StorageJSON stores as JSON/JSONB column
	StorageJSON

	// StorageArray stores as native array (if supported by database)
	StorageArray

	// StorageSeparateTable stores in a separate table with foreign key
	StorageSeparateTable

	// StorageSerialized stores as binary/protobuf serialization
	StorageSerialized
)

// Relationship defines a foreign key or nested relationship
type Relationship struct {
	// Field is the field containing the relationship
	Field *protogen.Field

	// Type is the relationship type
	Type RelationshipType

	// ReferencedMessage is the target message of the relationship
	ReferencedMessage *MessageMapping

	// ForeignKey is the foreign key column name
	ForeignKey string

	// ReferenceKey is the referenced column name
	ReferenceKey string

	// JoinTable is the join table name for many-to-many relationships
	JoinTable string

	// LazyLoad indicates if this relationship should be lazy-loaded
	LazyLoad bool

	// CascadeDelete indicates if deletes should cascade
	CascadeDelete bool
}

// RelationshipType defines the type of relationship
type RelationshipType int

const (
	OneToOne RelationshipType = iota
	OneToMany
	ManyToOne
	ManyToMany
)

func (r RelationshipType) String() string {
	switch r {
	case OneToOne:
		return "OneToOne"
	case OneToMany:
		return "OneToMany"
	case ManyToOne:
		return "ManyToOne"
	case ManyToMany:
		return "ManyToMany"
	default:
		return "Unknown"
	}
}

// GetStorageStrategy determines the storage strategy from column options
func GetStorageStrategy(colOpts *dalv1.ColumnOptions, field *protogen.Field) StorageStrategy {
	if colOpts == nil {
		return StorageAuto
	}

	// Check if column type indicates a specific strategy
	switch colOpts.Type {
	case "JSON", "JSONB":
		return StorageJSON
	case "TEXT[]", "VARCHAR[]", "INTEGER[]":
		return StorageArray
	default:
		return StorageAuto
	}
}
