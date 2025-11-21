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

package collector

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// Target identifies a datastore target
type Target int

const (
	// TargetGorm identifies GORM target (database-agnostic ORM)
	TargetGorm Target = iota
	// TargetPostgres identifies PostgreSQL target (raw SQL)
	TargetPostgres
	// TargetFirestore identifies Google Cloud Firestore target
	TargetFirestore
	// TargetMongoDB identifies MongoDB target
	TargetMongoDB
	// TargetDatastore identifies Google Cloud Datastore target
	TargetDatastore
)

// MessageInfo contains a DAL message and its metadata
type MessageInfo struct {
	// SourceMessage is the original API proto message (e.g., library.v1.Book)
	SourceMessage *protogen.Message

	// TargetMessage is the DAL schema message (e.g., library.v1.dal.BookPostgres)
	TargetMessage *protogen.Message

	// SourceName is the fully qualified source message name (e.g., "library.v1.Book")
	SourceName string

	// TableName is the database table/collection/kind name
	TableName string

	// SchemaName is the database schema/namespace (optional)
	SchemaName string

	// ImplementScanner indicates whether to generate driver.Valuer/sql.Scanner methods
	ImplementScanner bool
}

// CollectMessages finds all messages for a target across all proto files.
//
// This is the main entry point for the collection phase. It performs a two-pass scan:
//  1. First pass: Build an index of ALL messages by fully qualified name
//     (needed to resolve source message references)
//  2. Second pass: Find messages with target annotations and link to source messages
//
// Why collect all at once?
// - Foreign keys may reference messages in different files
// - Relationships (one-to-many, many-to-many) need cross-file resolution
// - Generators need the complete picture to analyze relationships
//
// Parameters:
//   - gen: The protogen plugin containing all proto files
//   - target: Which datastore target to collect (postgres, firestore, etc.)
//
// Returns:
//   - Slice of MessageInfo, one per DAL schema message found
//   - Error if any messages reference missing source messages
func CollectMessages(gen *protogen.Plugin, target Target) ([]*MessageInfo, error) {
	var collected []*MessageInfo
	var errors []string

	// Build index of all messages for source lookup
	// This allows us to resolve "source: library.v1.Book" references quickly
	messageIndex := buildMessageIndex(gen)

	// Scan all files looking for messages with target annotations
	for _, file := range gen.Files {
		// Skip files not marked for generation (e.g., imported dependencies)
		if !file.Generate {
			continue
		}

		for _, msg := range file.Messages {
			info, err := extractMessageInfo(msg, target, messageIndex)
			if err != nil {
				errors = append(errors, err.Error())
			}
			if info != nil {
				collected = append(collected, info)
			}
		}
	}

	// If we encountered any errors, fail
	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to collect messages:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return collected, nil
}

// buildMessageIndex creates a map of fully qualified message names to messages.
//
// Why build an index?
// DAL schema messages reference source messages via fully qualified names
// (e.g., "library.v1.Book"). We need fast lookup to link them together.
//
// Example:
//
//	message BookPostgres {
//	  option (dal.v1.postgres) = {source: "library.v1.Book"};  // <- Need to find this
//	}
//
// The index maps "library.v1.Book" -> *protogen.Message for Book
//
// Note: We index ALL messages (not just those being generated) because
// source messages might be in imported files.
func buildMessageIndex(gen *protogen.Plugin) map[string]*protogen.Message {
	index := make(map[string]*protogen.Message)
	for _, file := range gen.Files {
		for _, msg := range file.Messages {
			fqn := string(msg.Desc.FullName())
			index[fqn] = msg
		}
	}
	return index
}

// extractMessageInfo extracts message info for a specific target.
//
// This function checks if a message has an annotation for the requested target.
// If yes, it extracts the metadata and links to the source message.
//
// Why switch on target here?
// Each target has different annotation types (PostgresOptions, FirestoreOptions, etc.).
// By dispatching here, we keep target-specific logic isolated in separate functions.
//
// Returns:
//   - MessageInfo if message has annotation for this target and source is found
//   - nil, nil if message doesn't have annotation for this target
//   - nil, error if source message cannot be found (broken reference)
func extractMessageInfo(msg *protogen.Message, target Target, index map[string]*protogen.Message) (*MessageInfo, error) {
	opts := msg.Desc.Options()
	if opts == nil {
		return nil, nil
	}

	// Dispatch to target-specific extraction
	switch target {
	case TargetGorm:
		return extractGormInfo(msg, opts, index)
	case TargetPostgres:
		return extractPostgresInfo(msg, opts, index)
	case TargetFirestore:
		return extractFirestoreInfo(msg, opts, index)
	case TargetMongoDB:
		return extractMongoDBInfo(msg, opts, index)
	case TargetDatastore:
		return extractDatastoreInfo(msg, opts, index)
	}

	return nil, nil
}

// extractGormInfo extracts GORM message info.
//
// Looks for the (dal.v1.gorm) annotation on the message.
//
// Example proto:
//
//	message BookGorm {
//	  option (dal.v1.gorm) = {
//	    source: "library.v1.Book"    // API message to convert from
//	    table: "books"                // Table name
//	  };
//	}
//
// Returns error if source message is not found.
func extractGormInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) (*MessageInfo, error) {
	// Check if message has gorm annotation
	v := proto.GetExtension(opts, dalv1.E_Gorm)
	if v == nil {
		return nil, nil
	}

	gormOpts, ok := v.(*dalv1.GormOptions)
	if !ok || gormOpts == nil {
		return nil, nil
	}

	// Look up source message by fully qualified name
	sourceMsg := index[gormOpts.Source]
	if sourceMsg == nil {
		// Source message not found - return error
		return nil, fmt.Errorf("message '%s' references source '%s' which does not exist. Please ensure the source proto file is imported and the message name is correct",
			msg.Desc.FullName(), gormOpts.Source)
	}

	return &MessageInfo{
		SourceMessage:    sourceMsg,
		TargetMessage:    msg,
		SourceName:       gormOpts.Source,
		TableName:        gormOpts.Table,
		SchemaName:       "", // GORM doesn't use schema
		ImplementScanner: gormOpts.ImplementScanner,
	}, nil
}

// extractPostgresInfo extracts PostgreSQL message info.
//
// Looks for the (dal.v1.postgres) annotation on the message.
//
// Example proto:
//
//	message BookPostgres {
//	  option (dal.v1.postgres) = {
//	    source: "library.v1.Book"    // API message to convert from
//	    table: "books"                // PostgreSQL table name
//	    schema: "library"             // PostgreSQL schema (optional)
//	  };
//	}
//
// Returns error if source message is not found.
func extractPostgresInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) (*MessageInfo, error) {
	// Check if message has postgres annotation
	if v := proto.GetExtension(opts, dalv1.E_Postgres); v != nil {
		if pgOpts, ok := v.(*dalv1.PostgresOptions); ok && pgOpts != nil {
			// Look up source message by fully qualified name
			sourceMsg := index[pgOpts.Source]
			if sourceMsg == nil {
				// Source message not found - return error
				return nil, fmt.Errorf("message '%s' references source '%s' which does not exist. Please ensure the source proto file is imported and the message name is correct",
					msg.Desc.FullName(), pgOpts.Source)
			}

			return &MessageInfo{
				SourceMessage: sourceMsg,
				TargetMessage: msg,
				SourceName:    pgOpts.Source,
				TableName:     pgOpts.Table,
				SchemaName:    pgOpts.Schema,
			}, nil
		}
	}
	return nil, nil
}

// extractFirestoreInfo extracts Firestore message info.
//
// Similar to extractPostgresInfo but for Firestore target.
//
// Example proto:
//
//	message BookFirestore {
//	  option (dal.v1.firestore) = {
//	    source: "library.v1.Book"
//	    collection: "books"
//	  };
//	}
//
// Note: TableName is used for "collection" name to keep MessageInfo generic.
// Returns error if source message is not found.
func extractFirestoreInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) (*MessageInfo, error) {
	if v := proto.GetExtension(opts, dalv1.E_Firestore); v != nil {
		if fsOpts, ok := v.(*dalv1.FirestoreOptions); ok && fsOpts != nil {
			sourceMsg := index[fsOpts.Source]
			if sourceMsg == nil {
				return nil, fmt.Errorf("message '%s' references source '%s' which does not exist. Please ensure the source proto file is imported and the message name is correct",
					msg.Desc.FullName(), fsOpts.Source)
			}

			return &MessageInfo{
				SourceMessage: sourceMsg,
				TargetMessage: msg,
				SourceName:    fsOpts.Source,
				TableName:     fsOpts.Collection, // Firestore uses "collection" instead of "table"
			}, nil
		}
	}
	return nil, nil
}

// extractMongoDBInfo extracts MongoDB message info.
//
// Similar to extractPostgresInfo but for MongoDB target.
//
// Example proto:
//
//	message BookMongo {
//	  option (dal.v1.mongodb) = {
//	    source: "library.v1.Book"
//	    collection: "books"
//	    database: "library_db"
//	  };
//	}
//
// Note: SchemaName is used for "database" name to keep MessageInfo generic.
// Returns error if source message is not found.
func extractMongoDBInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) (*MessageInfo, error) {
	if v := proto.GetExtension(opts, dalv1.E_Mongodb); v != nil {
		if mongoOpts, ok := v.(*dalv1.MongoDBOptions); ok && mongoOpts != nil {
			sourceMsg := index[mongoOpts.Source]
			if sourceMsg == nil {
				return nil, fmt.Errorf("message '%s' references source '%s' which does not exist. Please ensure the source proto file is imported and the message name is correct",
					msg.Desc.FullName(), mongoOpts.Source)
			}

			return &MessageInfo{
				SourceMessage: sourceMsg,
				TargetMessage: msg,
				SourceName:    mongoOpts.Source,
				TableName:     mongoOpts.Collection, // MongoDB uses "collection"
				SchemaName:    mongoOpts.Database,   // SchemaName repurposed for "database"
			}, nil
		}
	}
	return nil, nil
}

// extractDatastoreInfo extracts Google Cloud Datastore message info.
//
// Looks for the (dal.v1.datastore_options) annotation on the message.
//
// Example proto:
//
//	message UserDatastore {
//	  option (dal.v1.datastore_options) = {
//	    source: "api.v1.User"    // API message to convert from
//	    kind: "User"              // Datastore kind name
//	    namespace: "prod"         // Datastore namespace (optional)
//	  };
//	}
//
// Note: TableName is used for "kind" name and SchemaName for "namespace"
// to keep MessageInfo generic across all targets.
// Returns error if source message is not found.
func extractDatastoreInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) (*MessageInfo, error) {
	// Check if message has datastore_options annotation
	if v := proto.GetExtension(opts, dalv1.E_DatastoreOptions); v != nil {
		if dsOpts, ok := v.(*dalv1.DatastoreOptions); ok && dsOpts != nil {
			// Look up source message by fully qualified name
			sourceMsg := index[dsOpts.Source]
			if sourceMsg == nil {
				// Source message not found - return error
				return nil, fmt.Errorf("message '%s' references source '%s' which does not exist. Please ensure the source proto file is imported and the message name is correct",
					msg.Desc.FullName(), dsOpts.Source)
			}

			return &MessageInfo{
				SourceMessage: sourceMsg,
				TargetMessage: msg,
				SourceName:    dsOpts.Source,
				TableName:     dsOpts.Kind,      // Datastore uses "kind" instead of "table"
				SchemaName:    dsOpts.Namespace, // SchemaName repurposed for "namespace"
			}, nil
		}
	}
	return nil, nil
}
