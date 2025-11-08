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
	"google.golang.org/protobuf/compiler/protogen"
	dalv1 "github.com/panyam/protoc-gen-go-dal/proto/gen/go/dal/v1"
)

// MessageMapping represents the intermediate representation of a DAL message mapping
// This is language-agnostic and can be used by any renderer (Go, Python, TypeScript, etc.)
type MessageMapping struct {
	// SourceMessage is the original API proto message (e.g., library.v1.Book)
	SourceMessage *protogen.Message

	// TargetMessage is the DAL schema message (e.g., library.v1.dal.BookGORM)
	TargetMessage *protogen.Message

	// TableName is the database table name
	TableName string

	// SchemaName is the database schema/namespace
	SchemaName string

	// Target identifies the generation target (gorm, raw, datastore, etc.)
	Target string

	// Fields are the field mappings between source and target
	Fields []*FieldMapping

	// Relationships defines foreign keys and nested relationships
	Relationships []*Relationship

	// Indexes defines table indexes
	Indexes []*Index

	// Hooks configuration
	Hooks *HookConfig

	// Target-specific options
	GormOptions      *dalv1.GormOptions
	DatastoreOptions *dalv1.DatastoreOptions
}

// HookConfig defines which lifecycle hooks should be generated
type HookConfig struct {
	EnableBeforeCreate bool
	EnableAfterCreate  bool
	EnableBeforeUpdate bool
	EnableAfterUpdate  bool
	EnableBeforeDelete bool
	EnableAfterDelete  bool
	EnableAfterLoad    bool
}

// Index represents a database index
type Index struct {
	Name   string
	Fields []string
	Unique bool
	Type   string
	Where  string
}
