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

/*
protoc-gen-dal is a Protocol Buffers compiler plugin that generates Data Access Layer (DAL) converters for protobuf messages.

# Overview

This generator creates converter functions for transforming between protobuf messages and various database/datastore representations.
It supports generating converters for SQL databases, NoSQL stores, and other persistence backends.

# Installation

	go install github.com/panyam/protoc-gen-dal/cmd/protoc-gen-dal@latest

Verify installation:

	which protoc-gen-dal
	# Should output: /path/to/go/bin/protoc-gen-dal

# Usage with buf

Add to your buf.gen.yaml:

	version: v2
	plugins:
	  # Generate standard Go protobuf types
	  - remote: buf.build/protocolbuffers/go
	    out: ./gen/go
	    opt: paths=source_relative

	  # Generate DAL converters (raw database)
	  - local: protoc-gen-dal
	    out: ./gen/dal
	    opt:
	      - datastores=postgres
	      - target=raw

	  # OR Generate GORM models
	  - local: protoc-gen-dal
	    out: ./gen/dal
	    opt:
	      - datastores=postgres
	      - target=gorm

	  # OR Generate Google Cloud Datastore bindings
	  - local: protoc-gen-dal
	    out: ./gen/dal
	    opt:
	      - target=datastore

Generate code:

	buf generate

# Configuration Options

Core Generation:

  - dal_export_path: Path where DAL converter code should be generated (default: ".")
  - datastores: Comma-separated list of datastores to generate converters for (postgres|mysql|sqlite|mongodb|redis|dynamodb|firestore)
  - target: Generation target - raw (direct DB), gorm (GORM ORM), datastore (Google Cloud Datastore), firestore (Google Cloud Firestore)
  - package_name: Go package name for generated code (default: "dal")

Message Selection:

  - messages: Comma-separated list of messages to generate converters for (default: all)
  - message_include: Comma-separated glob patterns for messages to include
  - message_exclude: Comma-separated glob patterns for messages to exclude

# Generated Files

The generator produces converter files per datastore:

  - {package}_postgres.go: PostgreSQL converters (ToRow/FromRow)
  - {package}_mongodb.go: MongoDB converters (ToBSON/FromBSON)
  - {package}_redis.go: Redis converters (ToHash/FromHash)

# Usage Example

Define your message:

	syntax = "proto3";
	package library.v1;

	message Book {
	  string id = 1;
	  string title = 2;
	  string author = 3;
	  int64 published_year = 4;
	}

Generate DAL converters:

	buf generate

Use the generated converters:

	package main

	import (
	    "database/sql"
	    libraryv1 "your-project/gen/go/library/v1"
	    "your-project/gen/dal/library/v1"
	)

	func main() {
	    db, _ := sql.Open("postgres", "...")

	    // Convert protobuf to SQL row values
	    book := &libraryv1.Book{
	        Id: "1",
	        Title: "Go Programming",
	        Author: "John Doe",
	        PublishedYear: 2024,
	    }

	    // ToPostgresRow returns column names and values
	    columns, values := library_v1.BookToPostgresRow(book)
	    _, err := db.Exec("INSERT INTO books ...") // use columns/values

	    // FromPostgresRow populates protobuf from SQL row
	    rows, _ := db.Query("SELECT * FROM books WHERE id = $1", "1")
	    defer rows.Close()
	    if rows.Next() {
	        foundBook := library_v1.BookFromPostgresRow(rows)
	    }
	}

# Architecture

The generator uses a simple architecture:

 1. DALGenerator: Orchestrates converter generation
 2. Filters: Message filtering
 3. Templates: Converter code templates per datastore type

# Links

Documentation:

  - GitHub: https://github.com/panyam/protoc-gen-dal
*/
package main

import (
	"flag"
	"fmt"
	"log"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/panyam/protoc-gen-dal/pkg/builders"
	"github.com/panyam/protoc-gen-dal/pkg/filters"
	"github.com/panyam/protoc-gen-dal/pkg/generators"
)

func main() {
	var flagSet flag.FlagSet

	// Core generation options
	dalExportPath := flagSet.String("dal_export_path", ".", "Path where DAL converter code should be generated")
	datastores := flagSet.String("datastores", "postgres", "Comma-separated list of datastores (postgres|mysql|sqlite|mongodb|redis|dynamodb|firestore)")
	packageName := flagSet.String("package_name", "dal", "Go package name for generated code")

	// Target selection: raw database or ORM/API
	target := flagSet.String("target", "raw", "Generation target (raw|gorm|datastore|firestore)")

	// Message selection
	messages := flagSet.String("messages", "", "Comma-separated list of messages to generate converters for (default: all)")
	messageInclude := flagSet.String("message_include", "", "Comma-separated glob patterns for messages to include")
	messageExclude := flagSet.String("message_exclude", "", "Comma-separated glob patterns for messages to exclude")

	protogen.Options{
		ParamFunc: flagSet.Set,
	}.Run(func(gen *protogen.Plugin) error {
		log.Printf("DAL GENERATOR: Plugin callback started")
		log.Printf("DAL GENERATOR: Request has %d files", len(gen.Files))
		log.Printf("DAL GENERATOR: Request parameters: %+v", gen.Request.GetParameter())

		defer func() {
			log.Printf("DAL GENERATOR: Plugin callback ending")
			log.Printf("DAL GENERATOR: Response has %d files", len(gen.Response().File))
			for i, file := range gen.Response().File {
				log.Printf("DAL GENERATOR: Response file %d: %s (%d bytes)",
					i, file.GetName(), len(file.GetContent()))
			}
		}()

		// Create generation configuration
		config := &builders.GenerationConfig{
			DALExportPath: *dalExportPath,
			Datastores:    *datastores,
			PackageName:   *packageName,
			Target:        *target,
		}

		// Create filter criteria from configuration
		filterCriteria, err := filters.ParseFromConfig(*messages, *messageInclude, *messageExclude)
		if err != nil {
			return fmt.Errorf("invalid filter configuration: %w", err)
		}

		// Create DAL generator
		dalGenerator := generators.NewDALGenerator(gen)

		// Validate configuration
		if err := dalGenerator.ValidateConfig(config); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Perform generation with detailed error handling
		if err := dalGenerator.Generate(config, filterCriteria); err != nil {
			log.Printf("protoc-gen-dal: Generation failed: %v", err)
			return fmt.Errorf("DAL generation failed: %w", err)
		}

		log.Printf("protoc-gen-dal: Generation completed successfully")
		return nil
	})
}
