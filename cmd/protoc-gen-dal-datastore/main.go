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

package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/datastore"
)

func main() {
	// Parse flags (protogen requires this)
	var flags flag.FlagSet

	// Run the plugin
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		// Phase 1: Collect all Datastore messages
		messages, err := collector.CollectMessages(plugin, collector.TargetDatastore)
		if err != nil {
			return fmt.Errorf("failed to collect Datastore messages: %w", err)
		}

		if len(messages) == 0 {
			// No Datastore messages found - this is not an error, just skip
			return nil
		}

		// Phase 2: Generate Datastore entity struct code
		result, err := datastore.Generate(messages)
		if err != nil {
			return fmt.Errorf("failed to generate Datastore code: %w", err)
		}

		// Phase 3: Generate converter code
		converterResult, err := datastore.GenerateConverters(messages)
		if err != nil {
			return fmt.Errorf("failed to generate converter code: %w", err)
		}

		// Phase 4: Write generated files to plugin response
		for _, genFile := range result.Files {
			// Create a new file in the plugin response
			// The second parameter is the Go import path for this generated file
			f := plugin.NewGeneratedFile(genFile.Path, protogen.GoImportPath(genFile.Path))

			// Write the generated content
			f.P(genFile.Content)
		}

		// Write converter files
		for _, genFile := range converterResult.Files {
			f := plugin.NewGeneratedFile(genFile.Path, protogen.GoImportPath(genFile.Path))
			f.P(genFile.Content)
		}

		return nil
	})
}
