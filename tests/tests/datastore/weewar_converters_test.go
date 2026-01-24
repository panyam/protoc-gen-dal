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

package datastore

import (
	"reflect"
	"testing"
	"time"

	"github.com/panyam/protoc-gen-dal/tests/gen/datastore/datastore"
	v1 "github.com/panyam/protoc-gen-dal/tests/gen/go/weewar/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Tile Converter Tests
// =============================================================================

// TestTileConversion_AllFields verifies basic tile field conversion.
func TestTileConversion_AllFields(t *testing.T) {
	src := &v1.Tile{
		Q:                5,
		R:                10,
		TileType:         3,
		Player:           2,
		Shortcut:         "base",
		LastActedTurn:    15,
		LastToppedupTurn: 12,
	}

	// Convert to Datastore
	dsTile, err := datastore.TileToTileDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TileToTileDatastore failed: %v", err)
	}

	// Verify fields
	if dsTile.Q != src.Q {
		t.Errorf("Q mismatch: got %d, want %d", dsTile.Q, src.Q)
	}
	if dsTile.R != src.R {
		t.Errorf("R mismatch: got %d, want %d", dsTile.R, src.R)
	}
	if dsTile.TileType != src.TileType {
		t.Errorf("TileType mismatch: got %d, want %d", dsTile.TileType, src.TileType)
	}
	if dsTile.Player != src.Player {
		t.Errorf("Player mismatch: got %d, want %d", dsTile.Player, src.Player)
	}
	if dsTile.Shortcut != src.Shortcut {
		t.Errorf("Shortcut mismatch: got %s, want %s", dsTile.Shortcut, src.Shortcut)
	}
	if dsTile.LastActedTurn != src.LastActedTurn {
		t.Errorf("LastActedTurn mismatch: got %d, want %d", dsTile.LastActedTurn, src.LastActedTurn)
	}
	if dsTile.LastToppedupTurn != src.LastToppedupTurn {
		t.Errorf("LastToppedupTurn mismatch: got %d, want %d", dsTile.LastToppedupTurn, src.LastToppedupTurn)
	}

	// Convert back
	apiTile, err := datastore.TileFromTileDatastore(nil, dsTile, nil)
	if err != nil {
		t.Fatalf("TileFromTileDatastore failed: %v", err)
	}

	// Verify round-trip
	if apiTile.Q != src.Q || apiTile.R != src.R {
		t.Errorf("Round-trip coordinates mismatch: got (%d,%d), want (%d,%d)", apiTile.Q, apiTile.R, src.Q, src.R)
	}
	if apiTile.TileType != src.TileType {
		t.Errorf("Round-trip TileType mismatch: got %d, want %d", apiTile.TileType, src.TileType)
	}
}

// TestTileConversion_NilSource verifies nil handling.
func TestTileConversion_NilSource(t *testing.T) {
	dsTile, err := datastore.TileToTileDatastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("TileToTileDatastore with nil failed: %v", err)
	}
	if dsTile != nil {
		t.Errorf("Expected nil result for nil source, got %v", dsTile)
	}

	apiTile, err := datastore.TileFromTileDatastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("TileFromTileDatastore with nil failed: %v", err)
	}
	if apiTile != nil {
		t.Errorf("Expected nil result for nil source, got %v", apiTile)
	}
}

// =============================================================================
// AttackRecord Converter Tests
// =============================================================================

// TestAttackRecordConversion_AllFields verifies attack record conversion.
func TestAttackRecordConversion_AllFields(t *testing.T) {
	src := &v1.AttackRecord{
		Q:          3,
		R:          7,
		IsRanged:   true,
		TurnNumber: 25,
	}

	dsRecord, err := datastore.AttackRecordToAttackRecordDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("AttackRecordToAttackRecordDatastore failed: %v", err)
	}

	if dsRecord.Q != src.Q || dsRecord.R != src.R {
		t.Errorf("Coordinates mismatch: got (%d,%d), want (%d,%d)", dsRecord.Q, dsRecord.R, src.Q, src.R)
	}
	if dsRecord.IsRanged != src.IsRanged {
		t.Errorf("IsRanged mismatch: got %v, want %v", dsRecord.IsRanged, src.IsRanged)
	}
	if dsRecord.TurnNumber != src.TurnNumber {
		t.Errorf("TurnNumber mismatch: got %d, want %d", dsRecord.TurnNumber, src.TurnNumber)
	}

	// Round-trip
	apiRecord, err := datastore.AttackRecordFromAttackRecordDatastore(nil, dsRecord, nil)
	if err != nil {
		t.Fatalf("AttackRecordFromAttackRecordDatastore failed: %v", err)
	}

	if apiRecord.IsRanged != src.IsRanged {
		t.Errorf("Round-trip IsRanged mismatch: got %v, want %v", apiRecord.IsRanged, src.IsRanged)
	}
}

// =============================================================================
// IndexInfo Converter Tests
// =============================================================================

// TestIndexInfoConversion_WithTimestamps verifies timestamp handling.
func TestIndexInfoConversion_WithTimestamps(t *testing.T) {
	lastUpdated := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	lastIndexed := time.Date(2024, 6, 2, 15, 30, 0, 0, time.UTC)

	src := &v1.IndexInfo{
		LastUpdatedAt: timestamppb.New(lastUpdated),
		LastIndexedAt: timestamppb.New(lastIndexed),
		NeedsIndexing: true,
	}

	dsInfo, err := datastore.IndexInfoToIndexInfoDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("IndexInfoToIndexInfoDatastore failed: %v", err)
	}

	if !dsInfo.LastUpdatedAt.Equal(lastUpdated) {
		t.Errorf("LastUpdatedAt mismatch: got %v, want %v", dsInfo.LastUpdatedAt, lastUpdated)
	}
	if !dsInfo.LastIndexedAt.Equal(lastIndexed) {
		t.Errorf("LastIndexedAt mismatch: got %v, want %v", dsInfo.LastIndexedAt, lastIndexed)
	}
	if dsInfo.NeedsIndexing != src.NeedsIndexing {
		t.Errorf("NeedsIndexing mismatch: got %v, want %v", dsInfo.NeedsIndexing, src.NeedsIndexing)
	}

	// Round-trip
	apiInfo, err := datastore.IndexInfoFromIndexInfoDatastore(nil, dsInfo, nil)
	if err != nil {
		t.Fatalf("IndexInfoFromIndexInfoDatastore failed: %v", err)
	}

	if !apiInfo.LastUpdatedAt.AsTime().Equal(lastUpdated) {
		t.Errorf("Round-trip LastUpdatedAt mismatch")
	}
	if apiInfo.NeedsIndexing != src.NeedsIndexing {
		t.Errorf("Round-trip NeedsIndexing mismatch")
	}
}

// TestIndexInfoConversion_NilTimestamps verifies nil timestamp handling.
func TestIndexInfoConversion_NilTimestamps(t *testing.T) {
	src := &v1.IndexInfo{
		LastUpdatedAt: nil,
		LastIndexedAt: nil,
		NeedsIndexing: false,
	}

	dsInfo, err := datastore.IndexInfoToIndexInfoDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("IndexInfoToIndexInfoDatastore failed: %v", err)
	}

	if !dsInfo.LastUpdatedAt.IsZero() {
		t.Errorf("LastUpdatedAt should be zero, got %v", dsInfo.LastUpdatedAt)
	}
	if !dsInfo.LastIndexedAt.IsZero() {
		t.Errorf("LastIndexedAt should be zero, got %v", dsInfo.LastIndexedAt)
	}
}

// =============================================================================
// Unit Converter Tests (with nested AttackHistory)
// =============================================================================

// TestUnitConversion_WithAttackHistory verifies nested slice conversion.
func TestUnitConversion_WithAttackHistory(t *testing.T) {
	src := &v1.Unit{
		Q:               5,
		R:               8,
		Player:          1,
		UnitType:        4,
		Shortcut:        "tank",
		AvailableHealth: 80,
		DistanceLeft:    3.5,
		LastActedTurn:   10,
		AttackHistory: []*v1.AttackRecord{
			{Q: 4, R: 8, IsRanged: false, TurnNumber: 8},
			{Q: 6, R: 7, IsRanged: true, TurnNumber: 9},
		},
		ProgressionStep:   2,
		ChosenAlternative: "heavy",
	}

	dsUnit, err := datastore.UnitToUnitDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("UnitToUnitDatastore failed: %v", err)
	}

	// Verify basic fields
	if dsUnit.Q != src.Q || dsUnit.R != src.R {
		t.Errorf("Coordinates mismatch")
	}
	if dsUnit.AvailableHealth != src.AvailableHealth {
		t.Errorf("AvailableHealth mismatch: got %d, want %d", dsUnit.AvailableHealth, src.AvailableHealth)
	}
	if dsUnit.DistanceLeft != src.DistanceLeft {
		t.Errorf("DistanceLeft mismatch: got %f, want %f", dsUnit.DistanceLeft, src.DistanceLeft)
	}

	// Verify nested AttackHistory
	if len(dsUnit.AttackHistory) != 2 {
		t.Fatalf("AttackHistory length mismatch: got %d, want 2", len(dsUnit.AttackHistory))
	}
	if dsUnit.AttackHistory[0].Q != 4 {
		t.Errorf("AttackHistory[0].Q mismatch: got %d, want 4", dsUnit.AttackHistory[0].Q)
	}
	if !dsUnit.AttackHistory[1].IsRanged {
		t.Errorf("AttackHistory[1].IsRanged should be true")
	}

	// Round-trip
	apiUnit, err := datastore.UnitFromUnitDatastore(nil, dsUnit, nil)
	if err != nil {
		t.Fatalf("UnitFromUnitDatastore failed: %v", err)
	}

	if len(apiUnit.AttackHistory) != 2 {
		t.Fatalf("Round-trip AttackHistory length mismatch: got %d, want 2", len(apiUnit.AttackHistory))
	}
	if apiUnit.AttackHistory[1].IsRanged != true {
		t.Errorf("Round-trip AttackHistory[1].IsRanged mismatch")
	}
}

// TestUnitConversion_EmptyAttackHistory verifies empty slice handling.
func TestUnitConversion_EmptyAttackHistory(t *testing.T) {
	src := &v1.Unit{
		Q:             1,
		R:             1,
		Player:        1,
		AttackHistory: []*v1.AttackRecord{},
	}

	dsUnit, err := datastore.UnitToUnitDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("UnitToUnitDatastore failed: %v", err)
	}

	// Empty slice should result in nil (not converted when nil check fails)
	// Actually the check is for != nil, and empty slice is not nil
	// So it creates an empty slice
	if dsUnit.AttackHistory != nil && len(dsUnit.AttackHistory) != 0 {
		t.Errorf("AttackHistory should be empty, got %v", dsUnit.AttackHistory)
	}
}

// TestUnitConversion_NilAttackHistory verifies nil slice handling.
func TestUnitConversion_NilAttackHistory(t *testing.T) {
	src := &v1.Unit{
		Q:             1,
		R:             1,
		Player:        1,
		AttackHistory: nil,
	}

	dsUnit, err := datastore.UnitToUnitDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("UnitToUnitDatastore failed: %v", err)
	}

	if dsUnit.AttackHistory != nil {
		t.Errorf("AttackHistory should be nil, got %v", dsUnit.AttackHistory)
	}
}

// =============================================================================
// WorldData Converter Tests (nested tiles and units)
// =============================================================================

// TestWorldDataConversion_WithTilesAndUnits verifies complex nested structure.
func TestWorldDataConversion_WithTilesAndUnits(t *testing.T) {
	src := &v1.WorldData{
		Tiles: []*v1.Tile{
			{Q: 0, R: 0, TileType: 1},
			{Q: 1, R: 0, TileType: 2},
			{Q: 0, R: 1, TileType: 1},
		},
		Units: []*v1.Unit{
			{Q: 0, R: 0, Player: 1, UnitType: 1, AvailableHealth: 100},
			{Q: 1, R: 0, Player: 2, UnitType: 2, AvailableHealth: 80},
		},
	}

	dsWorld, err := datastore.WorldDataToWorldDataDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("WorldDataToWorldDataDatastore failed: %v", err)
	}

	if len(dsWorld.Tiles) != 3 {
		t.Fatalf("Tiles length mismatch: got %d, want 3", len(dsWorld.Tiles))
	}
	if len(dsWorld.Units) != 2 {
		t.Fatalf("Units length mismatch: got %d, want 2", len(dsWorld.Units))
	}

	// Verify tile data
	if dsWorld.Tiles[1].TileType != 2 {
		t.Errorf("Tiles[1].TileType mismatch: got %d, want 2", dsWorld.Tiles[1].TileType)
	}

	// Verify unit data
	if dsWorld.Units[0].AvailableHealth != 100 {
		t.Errorf("Units[0].AvailableHealth mismatch: got %d, want 100", dsWorld.Units[0].AvailableHealth)
	}

	// Round-trip
	apiWorld, err := datastore.WorldDataFromWorldDataDatastore(nil, dsWorld, nil)
	if err != nil {
		t.Fatalf("WorldDataFromWorldDataDatastore failed: %v", err)
	}

	if len(apiWorld.Tiles) != 3 {
		t.Fatalf("Round-trip Tiles length mismatch: got %d, want 3", len(apiWorld.Tiles))
	}
	if len(apiWorld.Units) != 2 {
		t.Fatalf("Round-trip Units length mismatch: got %d, want 2", len(apiWorld.Units))
	}
}

// =============================================================================
// GameConfiguration Converter Tests
// =============================================================================

// TestGameConfigurationConversion_WithPlayersAndTeams verifies game config conversion.
func TestGameConfigurationConversion_WithPlayersAndTeams(t *testing.T) {
	src := &v1.GameConfiguration{
		Players: []*v1.GamePlayer{
			{PlayerId: 1, Name: "Player 1", Color: "red", TeamId: 1, IsActive: true, Coins: 1000},
			{PlayerId: 2, Name: "Player 2", Color: "blue", TeamId: 2, IsActive: true, Coins: 1000},
		},
		Teams: []*v1.GameTeam{
			{TeamId: 1, Name: "Red Team", Color: "red", IsActive: true},
			{TeamId: 2, Name: "Blue Team", Color: "blue", IsActive: true},
		},
		IncomeConfigs: &v1.IncomeConfig{
			StartingCoins:  1000,
			GameIncome:     100,
			LandbaseIncome: 50,
		},
		Settings: &v1.GameSettings{
			AllowedUnits:  []int32{1, 2, 3, 4, 5},
			TurnTimeLimit: 300,
			TeamMode:      "versus",
			MaxTurns:      100,
		},
	}

	dsConfig, err := datastore.GameConfigurationToGameConfigurationDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GameConfigurationToGameConfigurationDatastore failed: %v", err)
	}

	// Verify players
	if len(dsConfig.Players) != 2 {
		t.Fatalf("Players length mismatch: got %d, want 2", len(dsConfig.Players))
	}
	if dsConfig.Players[0].Name != "Player 1" {
		t.Errorf("Players[0].Name mismatch: got %s, want 'Player 1'", dsConfig.Players[0].Name)
	}

	// Verify teams
	if len(dsConfig.Teams) != 2 {
		t.Fatalf("Teams length mismatch: got %d, want 2", len(dsConfig.Teams))
	}
	if dsConfig.Teams[1].Color != "blue" {
		t.Errorf("Teams[1].Color mismatch: got %s, want 'blue'", dsConfig.Teams[1].Color)
	}

	// Verify income config
	if dsConfig.IncomeConfigs.StartingCoins != 1000 {
		t.Errorf("IncomeConfigs.StartingCoins mismatch: got %d, want 1000", dsConfig.IncomeConfigs.StartingCoins)
	}

	// Verify settings
	if !reflect.DeepEqual(dsConfig.Settings.AllowedUnits, src.Settings.AllowedUnits) {
		t.Errorf("Settings.AllowedUnits mismatch: got %v, want %v", dsConfig.Settings.AllowedUnits, src.Settings.AllowedUnits)
	}

	// Round-trip
	apiConfig, err := datastore.GameConfigurationFromGameConfigurationDatastore(nil, dsConfig, nil)
	if err != nil {
		t.Fatalf("GameConfigurationFromGameConfigurationDatastore failed: %v", err)
	}

	if len(apiConfig.Players) != 2 {
		t.Fatalf("Round-trip Players length mismatch")
	}
	if apiConfig.Settings.TeamMode != "versus" {
		t.Errorf("Round-trip Settings.TeamMode mismatch: got %s, want 'versus'", apiConfig.Settings.TeamMode)
	}
}

// =============================================================================
// GameState Converter Tests (with enum)
// =============================================================================

// TestGameStateConversion_WithEnum verifies enum field handling.
func TestGameStateConversion_WithEnum(t *testing.T) {
	updatedAt := time.Date(2024, 8, 15, 12, 0, 0, 0, time.UTC)

	src := &v1.GameState{
		UpdatedAt:     timestamppb.New(updatedAt),
		GameId:        "game-123",
		TurnCounter:   10,
		CurrentPlayer: 1,
		StateHash:     "abc123hash",
		Version:       5,
		Status:        v1.GameStatus_GAME_STATUS_PLAYING,
		Finished:      false,
		WinningPlayer: 0,
		WinningTeam:   0,
		WorldData: &v1.WorldData{
			Tiles: []*v1.Tile{{Q: 0, R: 0, TileType: 1}},
			Units: []*v1.Unit{{Q: 0, R: 0, Player: 1, UnitType: 1}},
		},
	}

	dsState, err := datastore.GameStateToGameStateDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GameStateToGameStateDatastore failed: %v", err)
	}

	// Verify enum
	if dsState.Status != v1.GameStatus_GAME_STATUS_PLAYING {
		t.Errorf("Status mismatch: got %v, want %v", dsState.Status, v1.GameStatus_GAME_STATUS_PLAYING)
	}

	// Verify other fields
	if dsState.GameId != src.GameId {
		t.Errorf("GameId mismatch: got %s, want %s", dsState.GameId, src.GameId)
	}
	if dsState.TurnCounter != src.TurnCounter {
		t.Errorf("TurnCounter mismatch: got %d, want %d", dsState.TurnCounter, src.TurnCounter)
	}
	if dsState.Version != src.Version {
		t.Errorf("Version mismatch: got %d, want %d", dsState.Version, src.Version)
	}

	// Verify nested WorldData
	if len(dsState.WorldData.Tiles) != 1 {
		t.Errorf("WorldData.Tiles length mismatch: got %d, want 1", len(dsState.WorldData.Tiles))
	}

	// Round-trip
	apiState, err := datastore.GameStateFromGameStateDatastore(nil, dsState, nil)
	if err != nil {
		t.Fatalf("GameStateFromGameStateDatastore failed: %v", err)
	}

	if apiState.Status != v1.GameStatus_GAME_STATUS_PLAYING {
		t.Errorf("Round-trip Status mismatch: got %v, want %v", apiState.Status, v1.GameStatus_GAME_STATUS_PLAYING)
	}
	if apiState.GameId != src.GameId {
		t.Errorf("Round-trip GameId mismatch")
	}
}

// TestGameStateConversion_DifferentStatuses verifies different enum values.
func TestGameStateConversion_DifferentStatuses(t *testing.T) {
	statuses := []v1.GameStatus{
		v1.GameStatus_GAME_STATUS_UNSPECIFIED,
		v1.GameStatus_GAME_STATUS_PLAYING,
		v1.GameStatus_GAME_STATUS_PAUSED,
		v1.GameStatus_GAME_STATUS_ENDED,
	}

	for _, status := range statuses {
		src := &v1.GameState{
			GameId: "test-game",
			Status: status,
		}

		dsState, err := datastore.GameStateToGameStateDatastore(src, nil, nil)
		if err != nil {
			t.Fatalf("GameStateToGameStateDatastore failed for status %v: %v", status, err)
		}

		if dsState.Status != status {
			t.Errorf("Status mismatch for %v: got %v", status, dsState.Status)
		}

		apiState, err := datastore.GameStateFromGameStateDatastore(nil, dsState, nil)
		if err != nil {
			t.Fatalf("GameStateFromGameStateDatastore failed for status %v: %v", status, err)
		}

		if apiState.Status != status {
			t.Errorf("Round-trip Status mismatch for %v: got %v", status, apiState.Status)
		}
	}
}

// =============================================================================
// GameMoveGroup Converter Tests
// =============================================================================

// TestGameMoveGroupConversion_WithTimestampsAndMoves verifies move group conversion.
func TestGameMoveGroupConversion_WithTimestampsAndMoves(t *testing.T) {
	startedAt := time.Date(2024, 8, 15, 10, 0, 0, 0, time.UTC)
	endedAt := time.Date(2024, 8, 15, 10, 5, 0, 0, time.UTC)
	moveTime := time.Date(2024, 8, 15, 10, 2, 0, 0, time.UTC)

	src := &v1.GameMoveGroup{
		StartedAt: timestamppb.New(startedAt),
		EndedAt:   timestamppb.New(endedAt),
		Moves: []*v1.GameMove{
			{
				Player:      1,
				Timestamp:   timestamppb.New(moveTime),
				SequenceNum: 1,
				IsPermanent: true,
			},
		},
	}

	dsGroup, err := datastore.GameMoveGroupToGameMoveGroupDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GameMoveGroupToGameMoveGroupDatastore failed: %v", err)
	}

	if !dsGroup.StartedAt.Equal(startedAt) {
		t.Errorf("StartedAt mismatch: got %v, want %v", dsGroup.StartedAt, startedAt)
	}
	if !dsGroup.EndedAt.Equal(endedAt) {
		t.Errorf("EndedAt mismatch: got %v, want %v", dsGroup.EndedAt, endedAt)
	}
	if len(dsGroup.Moves) != 1 {
		t.Fatalf("Moves length mismatch: got %d, want 1", len(dsGroup.Moves))
	}
	if dsGroup.Moves[0].Player != 1 {
		t.Errorf("Moves[0].Player mismatch: got %d, want 1", dsGroup.Moves[0].Player)
	}

	// Round-trip
	apiGroup, err := datastore.GameMoveGroupFromGameMoveGroupDatastore(nil, dsGroup, nil)
	if err != nil {
		t.Fatalf("GameMoveGroupFromGameMoveGroupDatastore failed: %v", err)
	}

	if !apiGroup.StartedAt.AsTime().Equal(startedAt) {
		t.Errorf("Round-trip StartedAt mismatch")
	}
	if len(apiGroup.Moves) != 1 {
		t.Fatalf("Round-trip Moves length mismatch")
	}
}

// =============================================================================
// GameMoveHistory Converter Tests
// =============================================================================

// TestGameMoveHistoryConversion_WithGroups verifies move history with nested groups.
func TestGameMoveHistoryConversion_WithGroups(t *testing.T) {
	src := &v1.GameMoveHistory{
		GameId: "game-456",
		Groups: []*v1.GameMoveGroup{
			{
				StartedAt: timestamppb.Now(),
				Moves: []*v1.GameMove{
					{Player: 1, SequenceNum: 1},
					{Player: 2, SequenceNum: 2},
				},
			},
			{
				StartedAt: timestamppb.Now(),
				Moves: []*v1.GameMove{
					{Player: 1, SequenceNum: 3},
				},
			},
		},
	}

	dsHistory, err := datastore.GameMoveHistoryToGameMoveHistoryDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GameMoveHistoryToGameMoveHistoryDatastore failed: %v", err)
	}

	if dsHistory.GameId != src.GameId {
		t.Errorf("GameId mismatch: got %s, want %s", dsHistory.GameId, src.GameId)
	}
	if len(dsHistory.Groups) != 2 {
		t.Fatalf("Groups length mismatch: got %d, want 2", len(dsHistory.Groups))
	}
	if len(dsHistory.Groups[0].Moves) != 2 {
		t.Errorf("Groups[0].Moves length mismatch: got %d, want 2", len(dsHistory.Groups[0].Moves))
	}

	// Round-trip
	apiHistory, err := datastore.GameMoveHistoryFromGameMoveHistoryDatastore(nil, dsHistory, nil)
	if err != nil {
		t.Fatalf("GameMoveHistoryFromGameMoveHistoryDatastore failed: %v", err)
	}

	if apiHistory.GameId != src.GameId {
		t.Errorf("Round-trip GameId mismatch")
	}
	if len(apiHistory.Groups) != 2 {
		t.Fatalf("Round-trip Groups length mismatch")
	}
}

// =============================================================================
// GamePlayer and GameTeam Converter Tests
// =============================================================================

// TestGamePlayerConversion_AllFields verifies player conversion.
func TestGamePlayerConversion_AllFields(t *testing.T) {
	src := &v1.GamePlayer{
		PlayerId:      1,
		PlayerType:    "human",
		Color:         "red",
		TeamId:        1,
		Name:          "Test Player",
		IsActive:      true,
		StartingCoins: 1500,
		Coins:         1200,
	}

	dsPlayer, err := datastore.GamePlayerToGamePlayerDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GamePlayerToGamePlayerDatastore failed: %v", err)
	}

	if dsPlayer.PlayerId != src.PlayerId {
		t.Errorf("PlayerId mismatch")
	}
	if dsPlayer.PlayerType != src.PlayerType {
		t.Errorf("PlayerType mismatch")
	}
	if dsPlayer.Coins != src.Coins {
		t.Errorf("Coins mismatch: got %d, want %d", dsPlayer.Coins, src.Coins)
	}

	// Round-trip
	apiPlayer, err := datastore.GamePlayerFromGamePlayerDatastore(nil, dsPlayer, nil)
	if err != nil {
		t.Fatalf("GamePlayerFromGamePlayerDatastore failed: %v", err)
	}

	if apiPlayer.Name != src.Name {
		t.Errorf("Round-trip Name mismatch")
	}
}

// TestGameTeamConversion_AllFields verifies team conversion.
func TestGameTeamConversion_AllFields(t *testing.T) {
	src := &v1.GameTeam{
		TeamId:   1,
		Name:     "Alpha Team",
		Color:    "green",
		IsActive: true,
	}

	dsTeam, err := datastore.GameTeamToGameTeamDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GameTeamToGameTeamDatastore failed: %v", err)
	}

	if dsTeam.TeamId != src.TeamId {
		t.Errorf("TeamId mismatch")
	}
	if dsTeam.Name != src.Name {
		t.Errorf("Name mismatch")
	}
	if dsTeam.IsActive != src.IsActive {
		t.Errorf("IsActive mismatch")
	}

	// Round-trip
	apiTeam, err := datastore.GameTeamFromGameTeamDatastore(nil, dsTeam, nil)
	if err != nil {
		t.Fatalf("GameTeamFromGameTeamDatastore failed: %v", err)
	}

	if apiTeam.Color != src.Color {
		t.Errorf("Round-trip Color mismatch")
	}
}

// =============================================================================
// IncomeConfig Converter Tests
// =============================================================================

// TestIncomeConfigConversion_AllFields verifies all income fields.
func TestIncomeConfigConversion_AllFields(t *testing.T) {
	src := &v1.IncomeConfig{
		StartingCoins:     1000,
		GameIncome:        100,
		LandbaseIncome:    50,
		NavalbaseIncome:   75,
		AirportbaseIncome: 100,
		MissilesiloIncome: 25,
		MinesIncome:       30,
	}

	dsConfig, err := datastore.IncomeConfigToIncomeConfigDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("IncomeConfigToIncomeConfigDatastore failed: %v", err)
	}

	if dsConfig.StartingCoins != src.StartingCoins {
		t.Errorf("StartingCoins mismatch")
	}
	if dsConfig.NavalbaseIncome != src.NavalbaseIncome {
		t.Errorf("NavalbaseIncome mismatch: got %d, want %d", dsConfig.NavalbaseIncome, src.NavalbaseIncome)
	}
	if dsConfig.MinesIncome != src.MinesIncome {
		t.Errorf("MinesIncome mismatch")
	}

	// Round-trip
	apiConfig, err := datastore.IncomeConfigFromIncomeConfigDatastore(nil, dsConfig, nil)
	if err != nil {
		t.Fatalf("IncomeConfigFromIncomeConfigDatastore failed: %v", err)
	}

	if apiConfig.AirportbaseIncome != src.AirportbaseIncome {
		t.Errorf("Round-trip AirportbaseIncome mismatch")
	}
}

// =============================================================================
// GameSettings Converter Tests
// =============================================================================

// TestGameSettingsConversion_WithAllowedUnits verifies int32 slice handling.
func TestGameSettingsConversion_WithAllowedUnits(t *testing.T) {
	src := &v1.GameSettings{
		AllowedUnits:  []int32{1, 2, 3, 4, 5, 6, 7, 8},
		TurnTimeLimit: 600,
		TeamMode:      "coop",
		MaxTurns:      50,
	}

	dsSettings, err := datastore.GameSettingsToGameSettingsDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GameSettingsToGameSettingsDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(dsSettings.AllowedUnits, src.AllowedUnits) {
		t.Errorf("AllowedUnits mismatch: got %v, want %v", dsSettings.AllowedUnits, src.AllowedUnits)
	}
	if dsSettings.TurnTimeLimit != src.TurnTimeLimit {
		t.Errorf("TurnTimeLimit mismatch")
	}
	if dsSettings.TeamMode != src.TeamMode {
		t.Errorf("TeamMode mismatch")
	}
	if dsSettings.MaxTurns != src.MaxTurns {
		t.Errorf("MaxTurns mismatch")
	}

	// Round-trip
	apiSettings, err := datastore.GameSettingsFromGameSettingsDatastore(nil, dsSettings, nil)
	if err != nil {
		t.Fatalf("GameSettingsFromGameSettingsDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(apiSettings.AllowedUnits, src.AllowedUnits) {
		t.Errorf("Round-trip AllowedUnits mismatch")
	}
}

// =============================================================================
// World Converter Tests (deeply nested)
// =============================================================================

// TestWorldConversion_CompleteStructure verifies deeply nested World conversion.
func TestWorldConversion_CompleteStructure(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	src := &v1.World{
		CreatedAt:   timestamppb.New(createdAt),
		UpdatedAt:   timestamppb.New(updatedAt),
		Id:          "world-001",
		CreatorId:   "user-123",
		Name:        "Test World",
		Description: "A test world for unit testing",
		Tags:        []string{"test", "sample"},
		ImageUrl:    "https://example.com/world.png",
		Difficulty:  "medium",
		PreviewUrls: []string{"https://example.com/preview1.png", "https://example.com/preview2.png"},
		WorldData: &v1.WorldData{
			Tiles: []*v1.Tile{
				{Q: 0, R: 0, TileType: 1},
				{Q: 1, R: 0, TileType: 2},
			},
			Units: []*v1.Unit{
				{Q: 0, R: 0, Player: 1, UnitType: 1},
			},
		},
		DefaultGameConfig: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, Name: "Player 1"},
			},
			Settings: &v1.GameSettings{
				MaxTurns: 100,
			},
		},
		ScreenshotIndexInfo: &v1.IndexInfo{
			NeedsIndexing: true,
		},
		SearchIndexInfo: &v1.IndexInfo{
			NeedsIndexing: false,
		},
	}

	dsWorld, err := datastore.WorldToWorldDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("WorldToWorldDatastore failed: %v", err)
	}

	// Verify basic fields
	if dsWorld.Id != src.Id {
		t.Errorf("Id mismatch")
	}
	if dsWorld.Name != src.Name {
		t.Errorf("Name mismatch")
	}
	if !reflect.DeepEqual(dsWorld.Tags, src.Tags) {
		t.Errorf("Tags mismatch")
	}
	if !reflect.DeepEqual(dsWorld.PreviewUrls, src.PreviewUrls) {
		t.Errorf("PreviewUrls mismatch")
	}

	// Verify timestamps
	if !dsWorld.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt mismatch")
	}

	// Verify nested WorldData
	if len(dsWorld.WorldData.Tiles) != 2 {
		t.Errorf("WorldData.Tiles length mismatch")
	}

	// Verify nested DefaultGameConfig
	if len(dsWorld.DefaultGameConfig.Players) != 1 {
		t.Errorf("DefaultGameConfig.Players length mismatch")
	}

	// Verify nested IndexInfo
	if dsWorld.ScreenshotIndexInfo.NeedsIndexing != true {
		t.Errorf("ScreenshotIndexInfo.NeedsIndexing mismatch")
	}

	// Round-trip
	apiWorld, err := datastore.WorldFromWorldDatastore(nil, dsWorld, nil)
	if err != nil {
		t.Fatalf("WorldFromWorldDatastore failed: %v", err)
	}

	if apiWorld.Id != src.Id {
		t.Errorf("Round-trip Id mismatch")
	}
	if apiWorld.Name != src.Name {
		t.Errorf("Round-trip Name mismatch")
	}
	if len(apiWorld.WorldData.Tiles) != 2 {
		t.Errorf("Round-trip WorldData.Tiles length mismatch")
	}
}

// =============================================================================
// Game Converter Tests
// =============================================================================

// TestGameConversion_WithConfig verifies Game conversion with nested config.
func TestGameConversion_WithConfig(t *testing.T) {
	createdAt := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	src := &v1.Game{
		CreatedAt:   timestamppb.New(createdAt),
		Id:          "game-001",
		CreatorId:   "user-456",
		WorldId:     "world-001",
		Name:        "Test Game",
		Description: "A test game",
		Tags:        []string{"multiplayer"},
		Difficulty:  "hard",
		Config: &v1.GameConfiguration{
			Players: []*v1.GamePlayer{
				{PlayerId: 1, Name: "Alice"},
				{PlayerId: 2, Name: "Bob"},
			},
			Teams: []*v1.GameTeam{
				{TeamId: 1, Name: "Team A"},
			},
		},
		ScreenshotIndexInfo: &v1.IndexInfo{NeedsIndexing: true},
		SearchIndexInfo:     &v1.IndexInfo{NeedsIndexing: false},
	}

	dsGame, err := datastore.GameToGameDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("GameToGameDatastore failed: %v", err)
	}

	if dsGame.Id != src.Id {
		t.Errorf("Id mismatch")
	}
	if dsGame.WorldId != src.WorldId {
		t.Errorf("WorldId mismatch")
	}
	if len(dsGame.Config.Players) != 2 {
		t.Errorf("Config.Players length mismatch")
	}

	// Round-trip
	apiGame, err := datastore.GameFromGameDatastore(nil, dsGame, nil)
	if err != nil {
		t.Fatalf("GameFromGameDatastore failed: %v", err)
	}

	if apiGame.Name != src.Name {
		t.Errorf("Round-trip Name mismatch")
	}
	if len(apiGame.Config.Players) != 2 {
		t.Errorf("Round-trip Config.Players length mismatch")
	}
}

// =============================================================================
// Decorator Tests
// =============================================================================

// TestTileConversion_WithDecorator verifies decorator function is called.
func TestTileConversion_WithDecorator(t *testing.T) {
	src := &v1.Tile{Q: 5, R: 10, TileType: 1}

	decoratorCalled := false
	decorator := func(src *v1.Tile, dest *datastore.TileDatastore) error {
		decoratorCalled = true
		dest.Shortcut = "modified"
		return nil
	}

	dsTile, err := datastore.TileToTileDatastore(src, nil, decorator)
	if err != nil {
		t.Fatalf("TileToTileDatastore failed: %v", err)
	}

	if !decoratorCalled {
		t.Error("Decorator was not called")
	}
	if dsTile.Shortcut != "modified" {
		t.Errorf("Decorator modification not applied: got %s, want 'modified'", dsTile.Shortcut)
	}
}

// TestUnitConversion_WithDecorator verifies decorator on nested conversion.
func TestUnitConversion_WithDecorator(t *testing.T) {
	src := &v1.Unit{
		Q:      1,
		R:      1,
		Player: 1,
	}

	decorator := func(src *v1.Unit, dest *datastore.UnitDatastore) error {
		dest.ChosenAlternative = "decorated"
		return nil
	}

	dsUnit, err := datastore.UnitToUnitDatastore(src, nil, decorator)
	if err != nil {
		t.Fatalf("UnitToUnitDatastore failed: %v", err)
	}

	if dsUnit.ChosenAlternative != "decorated" {
		t.Errorf("Decorator modification not applied")
	}
}
