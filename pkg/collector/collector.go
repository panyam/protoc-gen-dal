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
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	dalv1 "github.com/panyam/protoc-gen-go-dal/proto/gen/go/dal/v1"
)

// Target identifies a datastore target
type Target int

const (
	// TargetPostgres identifies PostgreSQL target
	TargetPostgres Target = iota
	// TargetFirestore identifies Google Cloud Firestore target
	TargetFirestore
	// TargetMongoDB identifies MongoDB target
	TargetMongoDB
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
func CollectMessages(gen *protogen.Plugin, target Target) []*MessageInfo {
	var collected []*MessageInfo

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
			info := extractMessageInfo(msg, target, messageIndex)
			if info != nil {
				collected = append(collected, info)
			}
		}
	}

	return collected
}

// buildMessageIndex creates a map of fully qualified message names to messages.
//
// Why build an index?
// DAL schema messages reference source messages via fully qualified names
// (e.g., "library.v1.Book"). We need fast lookup to link them together.
//
// Example:
//   message BookPostgres {
//     option (dal.v1.postgres) = {source: "library.v1.Book"};  // <- Need to find this
//   }
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
// Returns nil if:
// - Message has no options
// - Message doesn't have an annotation for this target
// - Source message cannot be found (broken reference)
func extractMessageInfo(msg *protogen.Message, target Target, index map[string]*protogen.Message) *MessageInfo {
	opts := msg.Desc.Options()
	if opts == nil {
		return nil
	}

	// Dispatch to target-specific extraction
	switch target {
	case TargetPostgres:
		return extractPostgresInfo(msg, opts, index)
	case TargetFirestore:
		return extractFirestoreInfo(msg, opts, index)
	case TargetMongoDB:
		return extractMongoDBInfo(msg, opts, index)
	}

	return nil
}

// extractPostgresInfo extracts PostgreSQL message info.
//
// Looks for the (dal.v1.postgres) annotation on the message.
//
// Example proto:
//   message BookPostgres {
//     option (dal.v1.postgres) = {
//       source: "library.v1.Book"    // API message to convert from
//       table: "books"                // PostgreSQL table name
//       schema: "library"             // PostgreSQL schema (optional)
//     };
//   }
//
// Why check if source exists?
// If the source message isn't found, this is a broken reference.
// We return nil to skip this message rather than crash.
// In production, we might want to log a warning.
func extractPostgresInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) *MessageInfo {
	// Check if message has postgres annotation
	if v := proto.GetExtension(opts, dalv1.E_Postgres); v != nil {
		if pgOpts, ok := v.(*dalv1.PostgresOptions); ok && pgOpts != nil {
			// Look up source message by fully qualified name
			sourceMsg := index[pgOpts.Source]
			if sourceMsg == nil {
				// Source message not found - skip this DAL schema
				// TODO: Consider logging warning for broken reference
				return nil
			}

			return &MessageInfo{
				SourceMessage: sourceMsg,
				TargetMessage: msg,
				SourceName:    pgOpts.Source,
				TableName:     pgOpts.Table,
				SchemaName:    pgOpts.Schema,
			}
		}
	}
	return nil
}

// extractFirestoreInfo extracts Firestore message info.
//
// Similar to extractPostgresInfo but for Firestore target.
//
// Example proto:
//   message BookFirestore {
//     option (dal.v1.firestore) = {
//       source: "library.v1.Book"
//       collection: "books"
//     };
//   }
//
// Note: TableName is used for "collection" name to keep MessageInfo generic.
func extractFirestoreInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) *MessageInfo {
	if v := proto.GetExtension(opts, dalv1.E_Firestore); v != nil {
		if fsOpts, ok := v.(*dalv1.FirestoreOptions); ok && fsOpts != nil {
			sourceMsg := index[fsOpts.Source]
			if sourceMsg == nil {
				return nil
			}

			return &MessageInfo{
				SourceMessage: sourceMsg,
				TargetMessage: msg,
				SourceName:    fsOpts.Source,
				TableName:     fsOpts.Collection, // Firestore uses "collection" instead of "table"
			}
		}
	}
	return nil
}

// extractMongoDBInfo extracts MongoDB message info.
//
// Similar to extractPostgresInfo but for MongoDB target.
//
// Example proto:
//   message BookMongo {
//     option (dal.v1.mongodb) = {
//       source: "library.v1.Book"
//       collection: "books"
//       database: "library_db"
//     };
//   }
//
// Note: SchemaName is used for "database" name to keep MessageInfo generic.
func extractMongoDBInfo(msg *protogen.Message, opts proto.Message, index map[string]*protogen.Message) *MessageInfo {
	if v := proto.GetExtension(opts, dalv1.E_Mongodb); v != nil {
		if mongoOpts, ok := v.(*dalv1.MongoDBOptions); ok && mongoOpts != nil {
			sourceMsg := index[mongoOpts.Source]
			if sourceMsg == nil {
				return nil
			}

			return &MessageInfo{
				SourceMessage: sourceMsg,
				TargetMessage: msg,
				SourceName:    mongoOpts.Source,
				TableName:     mongoOpts.Collection, // MongoDB uses "collection"
				SchemaName:    mongoOpts.Database,    // SchemaName repurposed for "database"
			}
		}
	}
	return nil
}
