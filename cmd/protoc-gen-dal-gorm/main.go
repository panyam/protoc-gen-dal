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
	"github.com/panyam/protoc-gen-dal/pkg/gorm"
)

func main() {
	// Parse flags (protogen requires this)
	var flags flag.FlagSet
	generateDAL := flags.Bool("generate_dal", false, "Generate DAL helper methods")
	dalFilenameSuffix := flags.String("dal_filename_suffix", "_dal", "Suffix for DAL helper filename (e.g., '_dal' -> 'world_gorm_dal.go')")
	dalFilenamePrefix := flags.String("dal_filename_prefix", "", "Prefix for DAL helper filename (e.g., 'dal_' -> 'dal_world_gorm.go')")
	dalOutputDir := flags.String("dal_output_dir", "", "Subdirectory for DAL files relative to main output (e.g., 'dal' -> 'gen/gorm/dal/')")

	// Run the plugin
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		// Phase 1: Collect all GORM messages
		messages, err := collector.CollectMessages(plugin, collector.TargetGorm)
		if err != nil {
			return fmt.Errorf("failed to collect GORM messages: %w", err)
		}

		if len(messages) == 0 {
			// No GORM messages found - this is not an error, just skip
			return nil
		}

		// Phase 2: Generate GORM struct code
		result, err := gorm.Generate(messages)
		if err != nil {
			return fmt.Errorf("failed to generate GORM code: %w", err)
		}

		// Phase 3: Generate converter code
		converterResult, err := gorm.GenerateConverters(messages)
		if err != nil {
			return fmt.Errorf("failed to generate converter code: %w", err)
		}

		// Phase 3.5: Generate DAL helper code (if enabled)
		var dalResult *gorm.GenerateResult
		if *generateDAL {
			dalResult, err = gorm.GenerateDALHelpers(messages, &gorm.DALOptions{
				FilenameSuffix: *dalFilenameSuffix,
				FilenamePrefix: *dalFilenamePrefix,
				OutputDir:      *dalOutputDir,
			})
			if err != nil {
				return fmt.Errorf("failed to generate DAL helper code: %w", err)
			}
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

		// Write DAL helper files (if generated)
		if dalResult != nil {
			for _, genFile := range dalResult.Files {
				f := plugin.NewGeneratedFile(genFile.Path, protogen.GoImportPath(genFile.Path))
				f.P(genFile.Content)
			}
		}

		return nil
	})
}
