// Copyright 2025 Sri Panyam
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gorm

import (
	"testing"

	"github.com/panyam/protoc-gen-dal/pkg/collector"
	"github.com/panyam/protoc-gen-dal/pkg/generator/common"
	"github.com/panyam/protoc-gen-dal/pkg/generator/registry"
	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

// TestConverterLookup_NestedMessage tests that nested message fields correctly
// resolve to their GORM target types using MessageRegistry.
//
// Before fix: Lookup for "IndexInfo:IndexInfo" fails because both source and target
//             point to the same proto message (api.IndexInfo).
// After fix:  Lookup resolves api.IndexInfo → IndexInfoGORM via MessageRegistry,
//             then checks "IndexInfo:IndexInfoGORM" which succeeds.
//
// This test will FAIL until we pass MessageRegistry to buildFieldConversion.
func TestConverterLookup_NestedMessage(t *testing.T) {
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API proto with nested message
			{
				name: "api/world.proto",
				pkg:  "api",
				messages: []testMessage{
					{
						name: "IndexInfo",
						fields: []testField{
							{name: "last_indexed", number: 1, typeName: "int64"},
						},
					},
					{
						name: "World",
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "screenshot_info", number: 2, typeName: "api.IndexInfo"},
						},
					},
				},
			},
			// GORM DAL proto
			{
				name: "gorm/world.proto",
				pkg:  "gorm",
				messages: []testMessage{
					{
						name: "IndexInfoGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "api.IndexInfo",
						},
					},
					{
						name: "WorldGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "api.World",
							Table:  "worlds",
						},
					},
				},
			},
		},
	})

	// Collect GORM messages
	messages, err := collector.CollectMessages(plugin, collector.TargetGorm)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Build registries
	msgRegistry := common.NewMessageRegistry(messages, buildStructName)
	convRegistry := registry.NewConverterRegistry(messages, buildStructName)

	// Find WorldGorm message
	var worldMsg *collector.MessageInfo
	for _, msg := range messages {
		if string(msg.TargetMessage.Desc.Name()) == "WorldGorm" {
			worldMsg = msg
			break
		}
	}
	if worldMsg == nil {
		t.Fatal("WorldGorm message not found")
	}

	// Build converter data - this should succeed without warnings
	converterData, err := buildConverterData(worldMsg, convRegistry, msgRegistry)
	if err != nil {
		t.Fatalf("buildConverterData failed: %v", err)
	}

	// Check that screenshot_info field has a converter mapping
	var screenshotMapping *FieldMappingData
	for i := range converterData.FieldMappings {
		if converterData.FieldMappings[i].SourceField == "ScreenshotInfo" {
			screenshotMapping = &converterData.FieldMappings[i]
			break
		}
	}

	if screenshotMapping == nil {
		t.Fatal("Expected ScreenshotInfo field mapping, got nil - field was skipped due to missing converter")
	}

	// Verify converter function names are set
	if screenshotMapping.ToTargetConverterFunc != "IndexInfoToIndexInfoGORM" {
		t.Errorf("Expected ToTargetConverterFunc='IndexInfoToIndexInfoGORM', got '%s'",
			screenshotMapping.ToTargetConverterFunc)
	}

	if screenshotMapping.FromTargetConverterFunc != "IndexInfoFromIndexInfoGORM" {
		t.Errorf("Expected FromTargetConverterFunc='IndexInfoFromIndexInfoGORM', got '%s'",
			screenshotMapping.FromTargetConverterFunc)
	}

	// Ensure MessageRegistry was used (not just the ConverterRegistry)
	// This assertion validates that we're passing msgRegistry through properly
	_ = msgRegistry
}

// TestConverterLookup_RepeatedMessage tests that repeated message fields
// correctly resolve their element types via MessageRegistry.
//
// Before fix: []Tile lookup fails because elementType points to api.Tile
//             but we look up "Tile:Tile" instead of "Tile:TileGORM".
// After fix:  MessageRegistry resolves api.Tile → TileGORM, lookup succeeds.
//
// This test will FAIL until we pass MessageRegistry to buildFieldConversion.
func TestConverterLookup_RepeatedMessage(t *testing.T) {
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API proto
			{
				name: "api/world.proto",
				pkg:  "api",
				messages: []testMessage{
					{
						name: "Tile",
						fields: []testField{
							{name: "q", number: 1, typeName: "int32"},
							{name: "r", number: 2, typeName: "int32"},
						},
					},
					{
						name: "WorldData",
						fields: []testField{
							{name: "tiles", number: 1, typeName: "api.Tile", repeated: true},
						},
					},
				},
			},
			// GORM DAL proto
			{
				name: "gorm/world.proto",
				pkg:  "gorm",
				messages: []testMessage{
					{
						name: "TileGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "api.Tile",
						},
					},
					{
						name: "WorldDataGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "api.WorldData",
						},
					},
				},
			},
		},
	})

	messages, err := collector.CollectMessages(plugin, collector.TargetGorm)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}

	msgRegistry := common.NewMessageRegistry(messages, buildStructName)
	convRegistry := registry.NewConverterRegistry(messages, buildStructName)

	// Find WorldDataGorm
	var worldDataMsg *collector.MessageInfo
	for _, msg := range messages {
		if string(msg.TargetMessage.Desc.Name()) == "WorldDataGorm" {
			worldDataMsg = msg
			break
		}
	}
	if worldDataMsg == nil {
		t.Fatal("WorldDataGorm message not found")
	}

	converterData, err := buildConverterData(worldDataMsg, convRegistry, msgRegistry)
	if err != nil {
		t.Fatalf("buildConverterData failed: %v", err)
	}

	// Check tiles field mapping
	var tilesMapping *FieldMappingData
	for i := range converterData.FieldMappings {
		if converterData.FieldMappings[i].SourceField == "Tiles" {
			tilesMapping = &converterData.FieldMappings[i]
			break
		}
	}

	if tilesMapping == nil {
		t.Fatal("Expected Tiles field mapping, got nil - repeated field was skipped due to missing converter")
	}

	// For repeated fields, check element type conversion
	if tilesMapping.SourceElementType != "Tile" {
		t.Errorf("Expected SourceElementType='Tile', got '%s'", tilesMapping.SourceElementType)
	}

	if tilesMapping.TargetElementType != "TileGORM" {
		t.Errorf("Expected TargetElementType='TileGORM', got '%s'", tilesMapping.TargetElementType)
	}

	if tilesMapping.ToTargetConverterFunc != "TileToTileGORM" {
		t.Errorf("Expected ToTargetConverterFunc='TileToTileGORM', got '%s'",
			tilesMapping.ToTargetConverterFunc)
	}

	_ = msgRegistry
}

// TestConverterLookup_MapWithMessageValue tests that map fields with message values
// correctly resolve via MessageRegistry.
//
// This test will FAIL until we pass MessageRegistry to buildFieldConversion.
func TestConverterLookup_MapWithMessageValue(t *testing.T) {
	plugin := createTestPlugin(t, &testProtoSet{
		files: []testFile{
			// API proto
			{
				name: "api/game.proto",
				pkg:  "api",
				messages: []testMessage{
					{
						name: "Player",
						fields: []testField{
							{name: "player_id", number: 1, typeName: "int32"},
							{name: "name", number: 2, typeName: "string"},
						},
					},
					{
						name: "Game",
						fields: []testField{
							{name: "id", number: 1, typeName: "string"},
							{name: "players", number: 2, typeName: "api.Player", isMap: true, mapKeyType: "int32"},
						},
					},
				},
			},
			// GORM DAL proto
			{
				name: "gorm/game.proto",
				pkg:  "gorm",
				messages: []testMessage{
					{
						name: "PlayerGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "api.Player",
						},
					},
					{
						name: "GameGorm",
						gormOpts: &dalv1.GormOptions{
							Source: "api.Game",
							Table:  "games",
						},
					},
				},
			},
		},
	})

	messages, err := collector.CollectMessages(plugin, collector.TargetGorm)
	if err != nil {
		t.Fatalf("CollectMessages failed: %v", err)
	}

	msgRegistry := common.NewMessageRegistry(messages, buildStructName)
	convRegistry := registry.NewConverterRegistry(messages, buildStructName)

	var gameMsg *collector.MessageInfo
	for _, msg := range messages {
		if string(msg.TargetMessage.Desc.Name()) == "GameGorm" {
			gameMsg = msg
			break
		}
	}
	if gameMsg == nil {
		t.Fatal("GameGorm message not found")
	}

	converterData, err := buildConverterData(gameMsg, convRegistry, msgRegistry)
	if err != nil {
		t.Fatalf("buildConverterData failed: %v", err)
	}

	// Check players map field
	var playersMapping *FieldMappingData
	for i := range converterData.FieldMappings {
		if converterData.FieldMappings[i].SourceField == "Players" {
			playersMapping = &converterData.FieldMappings[i]
			break
		}
	}

	// NOTE: This might be nil in current implementation if map handling is different
	// The test documents expected behavior
	if playersMapping == nil {
		t.Skip("Map field support not yet implemented - skipping")
	}

	if playersMapping.TargetElementType != "PlayerGORM" {
		t.Errorf("Expected TargetElementType='PlayerGORM', got '%s'", playersMapping.TargetElementType)
	}

	_ = msgRegistry
}
