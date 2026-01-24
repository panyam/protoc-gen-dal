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
	"context"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	dsgen "github.com/panyam/protoc-gen-dal/tests/gen/datastore/datastore"
)

// Test kinds for Weewar entities
const (
	testWorldKind     = "TestWorld"
	testGameKind      = "TestGame"
	testGameStateKind = "TestGameState"
)

// WeewarDAL provides DAL methods for Weewar entities.
// Tests complex nested structures, arrays of structs, and embedded types.
type WeewarDAL struct {
	Kind      string
	Namespace string
}

func NewWeewarDAL(kind string) *WeewarDAL {
	return &WeewarDAL{Kind: kind}
}

func (d *WeewarDAL) newKey(id string) *datastore.Key {
	key := datastore.NameKey(d.Kind, id, nil)
	if d.Namespace != "" {
		key.Namespace = d.Namespace
	}
	return key
}

// TestWorldDAL_PutAndGet tests saving and retrieving a World entity with complex nested structures.
// World contains:
//   - Nested structs: WorldData, GameConfiguration, IndexInfo
//   - Arrays of strings: Tags, PreviewUrls
//   - Embedded structs with their own nested arrays
func TestWorldDAL_PutAndGet(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// Clean up before and after test
	cleanupKindInNamespace(ctx, client, testWorldKind, getTestNamespace())
	t.Cleanup(func() {
		cleanupKindInNamespace(ctx, client, testWorldKind, getTestNamespace())
	})

	dal := NewWeewarDAL(testWorldKind)
	if ns := getTestNamespace(); ns != "" {
		dal.Namespace = ns
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create a complex World entity
	world := &dsgen.WorldDatastore{
		Id:          "world-1",
		CreatorId:   "creator-123",
		Name:        "Test World",
		Description: "A test world with complex nested structures",
		Tags:        []string{"strategy", "multiplayer", "hex"},
		ImageUrl:    "https://example.com/world.png",
		Difficulty:  "medium",
		PreviewUrls: []string{"https://example.com/preview1.png", "https://example.com/preview2.png"},
		CreatedAt:   now,
		UpdatedAt:   now,
		WorldData: dsgen.WorldDataDatastore{
			Tiles: []dsgen.TileDatastore{
				{Q: 0, R: 0, TileType: 1, Player: 0, Shortcut: "A1"},
				{Q: 1, R: 0, TileType: 2, Player: 1, Shortcut: "B1"},
				{Q: 0, R: 1, TileType: 3, Player: 2, Shortcut: "A2"},
			},
			Units: []dsgen.UnitDatastore{
				{
					Q: 0, R: 0, Player: 1, UnitType: 1, Shortcut: "INF1",
					AvailableHealth: 10, DistanceLeft: 3.5,
					AttackHistory: []dsgen.AttackRecordDatastore{
						{Q: 1, R: 0, IsRanged: false, TurnNumber: 1},
						{Q: 2, R: 0, IsRanged: true, TurnNumber: 2},
					},
				},
				{
					Q: 1, R: 1, Player: 2, UnitType: 2, Shortcut: "TNK1",
					AvailableHealth: 8, DistanceLeft: 2.0,
				},
			},
		},
		DefaultGameConfig: dsgen.GameConfigurationDatastore{
			Players: []dsgen.GamePlayerDatastore{
				{PlayerId: 1, PlayerType: "human", Color: "red", TeamId: 1, Name: "Player 1", IsActive: true, StartingCoins: 100, Coins: 100},
				{PlayerId: 2, PlayerType: "ai", Color: "blue", TeamId: 2, Name: "Player 2", IsActive: true, StartingCoins: 100, Coins: 150},
			},
			Teams: []dsgen.GameTeamDatastore{
				{TeamId: 1, Name: "Team Red", Color: "red", IsActive: true},
				{TeamId: 2, Name: "Team Blue", Color: "blue", IsActive: true},
			},
			IncomeConfigs: dsgen.IncomeConfigDatastore{
				StartingCoins:     100,
				GameIncome:        10,
				LandbaseIncome:    5,
				NavalbaseIncome:   8,
				AirportbaseIncome: 12,
				MissilesiloIncome: 15,
				MinesIncome:       3,
			},
			Settings: dsgen.GameSettingsDatastore{
				AllowedUnits:  []int32{1, 2, 3, 4, 5},
				TurnTimeLimit: 300,
				TeamMode:      "ffa",
				MaxTurns:      50,
			},
		},
		ScreenshotIndexInfo: dsgen.IndexInfoDatastore{
			LastUpdatedAt: now,
			LastIndexedAt: now.Add(-time.Hour),
			NeedsIndexing: true,
		},
		SearchIndexInfo: dsgen.IndexInfoDatastore{
			LastUpdatedAt: now,
			LastIndexedAt: now,
			NeedsIndexing: false,
		},
	}

	// Put the entity
	key := dal.newKey(world.Id)
	_, err := client.Put(ctx, key, world)
	if err != nil {
		t.Fatalf("Failed to put World: %v", err)
	}

	// Get the entity back
	var retrieved dsgen.WorldDatastore
	err = client.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Failed to get World: %v", err)
	}

	// Verify basic fields
	if retrieved.Id != world.Id {
		t.Errorf("Id mismatch: got %q, want %q", retrieved.Id, world.Id)
	}
	if retrieved.Name != world.Name {
		t.Errorf("Name mismatch: got %q, want %q", retrieved.Name, world.Name)
	}

	// Verify tags array
	if len(retrieved.Tags) != len(world.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(retrieved.Tags), len(world.Tags))
	}

	// Verify nested WorldData.Tiles
	if len(retrieved.WorldData.Tiles) != len(world.WorldData.Tiles) {
		t.Errorf("Tiles count mismatch: got %d, want %d", len(retrieved.WorldData.Tiles), len(world.WorldData.Tiles))
	} else {
		for i, tile := range retrieved.WorldData.Tiles {
			if tile.Q != world.WorldData.Tiles[i].Q || tile.R != world.WorldData.Tiles[i].R {
				t.Errorf("Tile[%d] coordinates mismatch: got (%d,%d), want (%d,%d)",
					i, tile.Q, tile.R, world.WorldData.Tiles[i].Q, world.WorldData.Tiles[i].R)
			}
		}
	}

	// Verify nested WorldData.Units with AttackHistory
	if len(retrieved.WorldData.Units) != len(world.WorldData.Units) {
		t.Errorf("Units count mismatch: got %d, want %d", len(retrieved.WorldData.Units), len(world.WorldData.Units))
	} else {
		unit := retrieved.WorldData.Units[0]
		if len(unit.AttackHistory) != 2 {
			t.Errorf("AttackHistory count mismatch: got %d, want 2", len(unit.AttackHistory))
		}
	}

	// Verify Players in GameConfiguration
	if len(retrieved.DefaultGameConfig.Players) != 2 {
		t.Errorf("Players count mismatch: got %d, want 2", len(retrieved.DefaultGameConfig.Players))
	}

	// Verify Teams in GameConfiguration
	if len(retrieved.DefaultGameConfig.Teams) != 2 {
		t.Errorf("Teams count mismatch: got %d, want 2", len(retrieved.DefaultGameConfig.Teams))
	}

	// Verify AllowedUnits in Settings
	if len(retrieved.DefaultGameConfig.Settings.AllowedUnits) != 5 {
		t.Errorf("AllowedUnits count mismatch: got %d, want 5", len(retrieved.DefaultGameConfig.Settings.AllowedUnits))
	}

	t.Logf("Successfully saved and retrieved World with %d tiles, %d units, %d players",
		len(retrieved.WorldData.Tiles), len(retrieved.WorldData.Units), len(retrieved.DefaultGameConfig.Players))
}

// TestGameDAL_PutAndGet tests saving and retrieving a Game entity.
func TestGameDAL_PutAndGet(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKindInNamespace(ctx, client, testGameKind, getTestNamespace())
	t.Cleanup(func() {
		cleanupKindInNamespace(ctx, client, testGameKind, getTestNamespace())
	})

	dal := NewWeewarDAL(testGameKind)
	if ns := getTestNamespace(); ns != "" {
		dal.Namespace = ns
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	game := &dsgen.GameDatastore{
		Id:          "game-1",
		CreatorId:   "user-456",
		WorldId:     "world-1",
		Name:        "Epic Battle",
		Description: "A heated battle between two factions",
		Tags:        []string{"pvp", "ranked", "tournament"},
		ImageUrl:    "https://example.com/game.png",
		Difficulty:  "hard",
		PreviewUrls: []string{"https://example.com/game-preview.png"},
		CreatedAt:   now,
		UpdatedAt:   now,
		Config: dsgen.GameConfigurationDatastore{
			Players: []dsgen.GamePlayerDatastore{
				{PlayerId: 1, PlayerType: "human", Color: "green", Name: "Alice", IsActive: true, Coins: 200},
				{PlayerId: 2, PlayerType: "human", Color: "purple", Name: "Bob", IsActive: true, Coins: 180},
				{PlayerId: 3, PlayerType: "ai", Color: "orange", Name: "CPU", IsActive: false, Coins: 0},
			},
			Teams: []dsgen.GameTeamDatastore{
				{TeamId: 1, Name: "Alliance", Color: "green", IsActive: true},
				{TeamId: 2, Name: "Horde", Color: "purple", IsActive: true},
			},
			Settings: dsgen.GameSettingsDatastore{
				AllowedUnits:  []int32{1, 2, 3},
				TurnTimeLimit: 600,
				TeamMode:      "teams",
				MaxTurns:      100,
			},
		},
	}

	key := dal.newKey(game.Id)
	_, err := client.Put(ctx, key, game)
	if err != nil {
		t.Fatalf("Failed to put Game: %v", err)
	}

	var retrieved dsgen.GameDatastore
	err = client.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Failed to get Game: %v", err)
	}

	if retrieved.Name != game.Name {
		t.Errorf("Name mismatch: got %q, want %q", retrieved.Name, game.Name)
	}

	if len(retrieved.Config.Players) != 3 {
		t.Errorf("Players count mismatch: got %d, want 3", len(retrieved.Config.Players))
	}

	// Verify player details
	if retrieved.Config.Players[0].Name != "Alice" {
		t.Errorf("Player[0] name mismatch: got %q, want %q", retrieved.Config.Players[0].Name, "Alice")
	}

	t.Logf("Successfully saved and retrieved Game with %d players, %d teams",
		len(retrieved.Config.Players), len(retrieved.Config.Teams))
}

// TestGameStateDAL_PutAndGet tests saving and retrieving a GameState entity.
// GameState has WorldData with Tiles and Units arrays.
func TestGameStateDAL_PutAndGet(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	cleanupKindInNamespace(ctx, client, testGameStateKind, getTestNamespace())
	t.Cleanup(func() {
		cleanupKindInNamespace(ctx, client, testGameStateKind, getTestNamespace())
	})

	dal := NewWeewarDAL(testGameStateKind)
	if ns := getTestNamespace(); ns != "" {
		dal.Namespace = ns
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	gameState := &dsgen.GameStateDatastore{
		GameId:        "game-1",
		TurnCounter:   15,
		CurrentPlayer: 2,
		StateHash:     "abc123def456",
		Version:       42,
		Finished:      false,
		WinningPlayer: 0,
		WinningTeam:   0,
		UpdatedAt:     now,
		WorldData: dsgen.WorldDataDatastore{
			Tiles: []dsgen.TileDatastore{
				{Q: 0, R: 0, TileType: 1, Player: 1},
				{Q: 1, R: 0, TileType: 1, Player: 2},
				{Q: 2, R: 0, TileType: 2, Player: 0},
				{Q: 0, R: 1, TileType: 3, Player: 1},
				{Q: 1, R: 1, TileType: 1, Player: 2},
			},
			Units: []dsgen.UnitDatastore{
				{
					Q: 0, R: 0, Player: 1, UnitType: 1,
					AvailableHealth: 7, DistanceLeft: 0,
					LastActedTurn: 14,
					AttackHistory: []dsgen.AttackRecordDatastore{
						{Q: 1, R: 0, IsRanged: false, TurnNumber: 10},
						{Q: 1, R: 0, IsRanged: false, TurnNumber: 12},
						{Q: 1, R: 0, IsRanged: false, TurnNumber: 14},
					},
				},
				{
					Q: 1, R: 1, Player: 2, UnitType: 2,
					AvailableHealth: 5, DistanceLeft: 1.5,
					LastActedTurn: 15,
					AttackHistory: []dsgen.AttackRecordDatastore{
						{Q: 0, R: 0, IsRanged: true, TurnNumber: 11},
						{Q: 0, R: 0, IsRanged: true, TurnNumber: 13},
					},
				},
			},
		},
	}

	key := dal.newKey(gameState.GameId)
	_, err := client.Put(ctx, key, gameState)
	if err != nil {
		t.Fatalf("Failed to put GameState: %v", err)
	}

	var retrieved dsgen.GameStateDatastore
	err = client.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Failed to get GameState: %v", err)
	}

	if retrieved.TurnCounter != 15 {
		t.Errorf("TurnCounter mismatch: got %d, want 15", retrieved.TurnCounter)
	}

	if len(retrieved.WorldData.Tiles) != 5 {
		t.Errorf("Tiles count mismatch: got %d, want 5", len(retrieved.WorldData.Tiles))
	}

	if len(retrieved.WorldData.Units) != 2 {
		t.Errorf("Units count mismatch: got %d, want 2", len(retrieved.WorldData.Units))
	}

	// Verify attack history is preserved
	unit1 := retrieved.WorldData.Units[0]
	if len(unit1.AttackHistory) != 3 {
		t.Errorf("Unit[0] AttackHistory count mismatch: got %d, want 3", len(unit1.AttackHistory))
	}

	unit2 := retrieved.WorldData.Units[1]
	if len(unit2.AttackHistory) != 2 {
		t.Errorf("Unit[1] AttackHistory count mismatch: got %d, want 2", len(unit2.AttackHistory))
	}

	t.Logf("Successfully saved and retrieved GameState at turn %d with %d tiles, %d units",
		retrieved.TurnCounter, len(retrieved.WorldData.Tiles), len(retrieved.WorldData.Units))
}

// TestUnitWithAttackHistory tests saving a unit with a deep nested AttackHistory array.
func TestUnitWithAttackHistory(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	kind := "TestUnit"
	cleanupKindInNamespace(ctx, client, kind, getTestNamespace())
	t.Cleanup(func() {
		cleanupKindInNamespace(ctx, client, kind, getTestNamespace())
	})

	dal := NewWeewarDAL(kind)
	if ns := getTestNamespace(); ns != "" {
		dal.Namespace = ns
	}

	// Create a unit with extensive attack history
	unit := &dsgen.UnitDatastore{
		Q:                       5,
		R:                       3,
		Player:                  1,
		UnitType:                3,
		Shortcut:                "ARTY1",
		AvailableHealth:         6,
		DistanceLeft:            0,
		LastActedTurn:           20,
		LastToppedupTurn:        15,
		AttacksReceivedThisTurn: 2,
		ProgressionStep:         3,
		ChosenAlternative:       "upgrade_range",
		AttackHistory: []dsgen.AttackRecordDatastore{
			{Q: 4, R: 3, IsRanged: true, TurnNumber: 5},
			{Q: 4, R: 4, IsRanged: true, TurnNumber: 8},
			{Q: 5, R: 4, IsRanged: true, TurnNumber: 10},
			{Q: 6, R: 3, IsRanged: true, TurnNumber: 12},
			{Q: 6, R: 2, IsRanged: true, TurnNumber: 15},
			{Q: 5, R: 2, IsRanged: true, TurnNumber: 18},
			{Q: 4, R: 2, IsRanged: true, TurnNumber: 20},
		},
	}

	key := dal.newKey("unit-arty-1")
	_, err := client.Put(ctx, key, unit)
	if err != nil {
		t.Fatalf("Failed to put Unit: %v", err)
	}

	var retrieved dsgen.UnitDatastore
	err = client.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Failed to get Unit: %v", err)
	}

	if retrieved.Shortcut != "ARTY1" {
		t.Errorf("Shortcut mismatch: got %q, want %q", retrieved.Shortcut, "ARTY1")
	}

	if len(retrieved.AttackHistory) != 7 {
		t.Errorf("AttackHistory count mismatch: got %d, want 7", len(retrieved.AttackHistory))
	}

	// Verify attack history order is preserved
	if retrieved.AttackHistory[0].TurnNumber != 5 {
		t.Errorf("First attack turn mismatch: got %d, want 5", retrieved.AttackHistory[0].TurnNumber)
	}
	if retrieved.AttackHistory[6].TurnNumber != 20 {
		t.Errorf("Last attack turn mismatch: got %d, want 20", retrieved.AttackHistory[6].TurnNumber)
	}

	t.Logf("Successfully saved and retrieved Unit with %d attack records", len(retrieved.AttackHistory))
}
