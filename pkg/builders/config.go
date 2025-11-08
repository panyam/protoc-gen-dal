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

package builders

// GenerationConfig holds configuration for DAL converter generation
type GenerationConfig struct {
	// DALExportPath is the path where DAL converter code should be generated
	DALExportPath string

	// Datastores is a comma-separated list of datastores to generate converters for
	// Supported: postgres, mysql, sqlite, mongodb, redis, dynamodb, firestore
	Datastores string

	// PackageName is the Go package name for generated code
	PackageName string

	// Target specifies the generation target
	// Supported: raw (direct DB), gorm (GORM ORM), datastore (Google Cloud Datastore), firestore (Google Cloud Firestore)
	Target string
}
