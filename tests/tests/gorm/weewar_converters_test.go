package gorm

import (
	"testing"
	"time"

	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"github.com/panyam/protoc-gen-dal/tests/gen/gorm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestTileConversion_SimpleStruct verifies that Tile (a simple struct with multiple int32 fields)
// converts correctly.
func TestTileConversion_SimpleStruct(t *testing.T) {
	src := &api.Tile{
		Q:                1,
		R:                2,
		TileType:         3,
		Player:           1,
		Shortcut:         "A1",
		LastActedTurn:    10,
		LastToppedupTurn: 9,
	}

	// Convert to GORM
	gormTile, err := gorm.TileToTileGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("TileToTileGORM failed: %v", err)
	}

	// Verify all fields
	if gormTile.Q != src.Q {
		t.Errorf("Q mismatch: got %d, want %d", gormTile.Q, src.Q)
	}

	if gormTile.R != src.R {
		t.Errorf("R mismatch: got %d, want %d", gormTile.R, src.R)
	}

	if gormTile.TileType != src.TileType {
		t.Errorf("TileType mismatch: got %d, want %d", gormTile.TileType, src.TileType)
	}

	if gormTile.Player != src.Player {
		t.Errorf("Player mismatch: got %d, want %d", gormTile.Player, src.Player)
	}

	if gormTile.Shortcut != src.Shortcut {
		t.Errorf("Shortcut mismatch: got %s, want %s", gormTile.Shortcut, src.Shortcut)
	}

	if gormTile.LastActedTurn != src.LastActedTurn {
		t.Errorf("LastActedTurn mismatch: got %d, want %d", gormTile.LastActedTurn, src.LastActedTurn)
	}

	if gormTile.LastToppedupTurn != src.LastToppedupTurn {
		t.Errorf("LastToppedupTurn mismatch: got %d, want %d", gormTile.LastToppedupTurn, src.LastToppedupTurn)
	}

	// Convert back to API
	apiTile, err := gorm.TileFromTileGORM(nil, gormTile, nil)
	if err != nil {
		t.Fatalf("TileFromTileGORM failed: %v", err)
	}

	// Verify round-trip
	if apiTile.Q != src.Q {
		t.Errorf("Round-trip Q mismatch: got %d, want %d", apiTile.Q, src.Q)
	}

	if apiTile.R != src.R {
		t.Errorf("Round-trip R mismatch: got %d, want %d", apiTile.R, src.R)
	}

	if apiTile.TileType != src.TileType {
		t.Errorf("Round-trip TileType mismatch: got %d, want %d", apiTile.TileType, src.TileType)
	}

	if apiTile.Player != src.Player {
		t.Errorf("Round-trip Player mismatch: got %d, want %d", apiTile.Player, src.Player)
	}

	if apiTile.Shortcut != src.Shortcut {
		t.Errorf("Round-trip Shortcut mismatch: got %s, want %s", apiTile.Shortcut, src.Shortcut)
	}

	if apiTile.LastActedTurn != src.LastActedTurn {
		t.Errorf("Round-trip LastActedTurn mismatch: got %d, want %d", apiTile.LastActedTurn, src.LastActedTurn)
	}

	if apiTile.LastToppedupTurn != src.LastToppedupTurn {
		t.Errorf("Round-trip LastToppedupTurn mismatch: got %d, want %d", apiTile.LastToppedupTurn, src.LastToppedupTurn)
	}
}

// TestUnitConversion_WithNestedMessages verifies that Unit (with repeated AttackRecord)
// converts correctly.
func TestUnitConversion_WithNestedMessages(t *testing.T) {
	src := &api.Unit{
		Q:                       5,
		R:                       6,
		Player:                  2,
		UnitType:                100,
		Shortcut:                "U1",
		AvailableHealth:         50,
		DistanceLeft:            3.5,
		LastActedTurn:           15,
		LastToppedupTurn:        14,
		AttacksReceivedThisTurn: 2,
		AttackHistory: []*api.AttackRecord{
			{Q: 4, R: 5, IsRanged: true, TurnNumber: 14},
			{Q: 6, R: 7, IsRanged: false, TurnNumber: 15},
		},
		ProgressionStep:    1,
		ChosenAlternative: "attack",
	}

	// Convert to GORM
	gormUnit, err := gorm.UnitToUnitGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("UnitToUnitGORM failed: %v", err)
	}

	// Verify basic fields
	if gormUnit.Q != src.Q {
		t.Errorf("Q mismatch: got %d, want %d", gormUnit.Q, src.Q)
	}

	if gormUnit.R != src.R {
		t.Errorf("R mismatch: got %d, want %d", gormUnit.R, src.R)
	}

	if gormUnit.Player != src.Player {
		t.Errorf("Player mismatch: got %d, want %d", gormUnit.Player, src.Player)
	}

	if gormUnit.UnitType != src.UnitType {
		t.Errorf("UnitType mismatch: got %d, want %d", gormUnit.UnitType, src.UnitType)
	}

	if gormUnit.AvailableHealth != src.AvailableHealth {
		t.Errorf("AvailableHealth mismatch: got %d, want %d", gormUnit.AvailableHealth, src.AvailableHealth)
	}

	if gormUnit.DistanceLeft != src.DistanceLeft {
		t.Errorf("DistanceLeft mismatch: got %f, want %f", gormUnit.DistanceLeft, src.DistanceLeft)
	}

	// Verify AttackHistory (repeated nested message)
	if len(gormUnit.AttackHistory) != len(src.AttackHistory) {
		t.Fatalf("AttackHistory length mismatch: got %d, want %d", len(gormUnit.AttackHistory), len(src.AttackHistory))
	}

	for i, srcRecord := range src.AttackHistory {
		gormRecord := gormUnit.AttackHistory[i]
		if gormRecord.Q != srcRecord.Q {
			t.Errorf("AttackHistory[%d].Q mismatch: got %d, want %d", i, gormRecord.Q, srcRecord.Q)
		}
		if gormRecord.R != srcRecord.R {
			t.Errorf("AttackHistory[%d].R mismatch: got %d, want %d", i, gormRecord.R, srcRecord.R)
		}
		if gormRecord.IsRanged != srcRecord.IsRanged {
			t.Errorf("AttackHistory[%d].IsRanged mismatch: got %v, want %v", i, gormRecord.IsRanged, srcRecord.IsRanged)
		}
		if gormRecord.TurnNumber != srcRecord.TurnNumber {
			t.Errorf("AttackHistory[%d].TurnNumber mismatch: got %d, want %d", i, gormRecord.TurnNumber, srcRecord.TurnNumber)
		}
	}

	// Convert back to API
	apiUnit, err := gorm.UnitFromUnitGORM(nil, gormUnit, nil)
	if err != nil {
		t.Fatalf("UnitFromUnitGORM failed: %v", err)
	}

	// Verify round-trip
	if apiUnit.Q != src.Q {
		t.Errorf("Round-trip Q mismatch: got %d, want %d", apiUnit.Q, src.Q)
	}

	if apiUnit.Player != src.Player {
		t.Errorf("Round-trip Player mismatch: got %d, want %d", apiUnit.Player, src.Player)
	}

	// Verify AttackHistory round-trip
	if len(apiUnit.AttackHistory) != len(src.AttackHistory) {
		t.Fatalf("Round-trip AttackHistory length mismatch: got %d, want %d", len(apiUnit.AttackHistory), len(src.AttackHistory))
	}

	for i, srcRecord := range src.AttackHistory {
		apiRecord := apiUnit.AttackHistory[i]
		if apiRecord.Q != srcRecord.Q {
			t.Errorf("Round-trip AttackHistory[%d].Q mismatch: got %d, want %d", i, apiRecord.Q, srcRecord.Q)
		}
		if apiRecord.IsRanged != srcRecord.IsRanged {
			t.Errorf("Round-trip AttackHistory[%d].IsRanged mismatch: got %v, want %v", i, apiRecord.IsRanged, srcRecord.IsRanged)
		}
	}
}

// TestWorldDataConversion_DeepNestedStructures verifies that WorldData with
// repeated Tiles and Units converts correctly (2-level nesting).
func TestWorldDataConversion_DeepNestedStructures(t *testing.T) {
	src := &api.WorldData{
		Tiles: []*api.Tile{
			{Q: 0, R: 0, TileType: 1, Player: 0, Shortcut: "T1"},
			{Q: 1, R: 0, TileType: 2, Player: 1, Shortcut: "T2"},
			{Q: 0, R: 1, TileType: 1, Player: 0, Shortcut: "T3"},
		},
		Units: []*api.Unit{
			{Q: 0, R: 0, Player: 1, UnitType: 100, AvailableHealth: 100, DistanceLeft: 5.0},
			{Q: 1, R: 0, Player: 2, UnitType: 101, AvailableHealth: 80, DistanceLeft: 4.5},
		},
	}

	// Convert to GORM
	gormData, err := gorm.WorldDataToWorldDataGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("WorldDataToWorldDataGORM failed: %v", err)
	}

	// Verify Tiles
	if len(gormData.Tiles) != len(src.Tiles) {
		t.Fatalf("Tiles length mismatch: got %d, want %d", len(gormData.Tiles), len(src.Tiles))
	}

	for i, srcTile := range src.Tiles {
		gormTile := gormData.Tiles[i]
		if gormTile.Q != srcTile.Q || gormTile.R != srcTile.R {
			t.Errorf("Tiles[%d] coordinates mismatch: got (%d,%d), want (%d,%d)",
				i, gormTile.Q, gormTile.R, srcTile.Q, srcTile.R)
		}
		if gormTile.TileType != srcTile.TileType {
			t.Errorf("Tiles[%d].TileType mismatch: got %d, want %d", i, gormTile.TileType, srcTile.TileType)
		}
	}

	// Verify Units
	if len(gormData.Units) != len(src.Units) {
		t.Fatalf("Units length mismatch: got %d, want %d", len(gormData.Units), len(src.Units))
	}

	for i, srcUnit := range src.Units {
		gormUnit := gormData.Units[i]
		if gormUnit.Q != srcUnit.Q || gormUnit.R != srcUnit.R {
			t.Errorf("Units[%d] coordinates mismatch: got (%d,%d), want (%d,%d)",
				i, gormUnit.Q, gormUnit.R, srcUnit.Q, srcUnit.R)
		}
		if gormUnit.Player != srcUnit.Player {
			t.Errorf("Units[%d].Player mismatch: got %d, want %d", i, gormUnit.Player, srcUnit.Player)
		}
	}

	// Convert back to API
	apiData, err := gorm.WorldDataFromWorldDataGORM(nil, gormData, nil)
	if err != nil {
		t.Fatalf("WorldDataFromWorldDataGORM failed: %v", err)
	}

	// Verify round-trip
	if len(apiData.Tiles) != len(src.Tiles) {
		t.Fatalf("Round-trip Tiles length mismatch: got %d, want %d", len(apiData.Tiles), len(src.Tiles))
	}

	if len(apiData.Units) != len(src.Units) {
		t.Fatalf("Round-trip Units length mismatch: got %d, want %d", len(apiData.Units), len(src.Units))
	}

	for i, srcTile := range src.Tiles {
		apiTile := apiData.Tiles[i]
		if apiTile.Q != srcTile.Q || apiTile.R != srcTile.R {
			t.Errorf("Round-trip Tiles[%d] coordinates mismatch: got (%d,%d), want (%d,%d)",
				i, apiTile.Q, apiTile.R, srcTile.Q, srcTile.R)
		}
	}
}

// TestWorldConversion_ThreeLevelNesting verifies that World with nested WorldData
// (which contains Tiles and Units) converts correctly (3-level nesting).
func TestWorldConversion_ThreeLevelNesting(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)

	src := &api.World{
		CreatedAt:   timestamppb.New(createdAt),
		UpdatedAt:   timestamppb.New(updatedAt),
		Id:          "world-123",
		CreatorId:   "user-456",
		Name:        "Test World",
		Description: "A test world for converter testing",
		Tags:        []string{"test", "weewar", "game"},
		ImageUrl:    "https://example.com/world.png",
		Difficulty:  "medium",
		WorldData: &api.WorldData{
			Tiles: []*api.Tile{
				{Q: 0, R: 0, TileType: 1, Player: 0},
				{Q: 1, R: 0, TileType: 2, Player: 1},
			},
			Units: []*api.Unit{
				{Q: 0, R: 0, Player: 1, UnitType: 100, AvailableHealth: 100},
			},
		},
		PreviewUrls: []string{
			"https://example.com/preview1.png",
			"https://example.com/preview2.png",
		},
	}

	// Convert to GORM
	gormWorld, err := gorm.WorldToWorldGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("WorldToWorldGORM failed: %v", err)
	}

	// Verify top-level fields
	if gormWorld.Id != src.Id {
		t.Errorf("Id mismatch: got %s, want %s", gormWorld.Id, src.Id)
	}

	if gormWorld.CreatorId != src.CreatorId {
		t.Errorf("CreatorId mismatch: got %s, want %s", gormWorld.CreatorId, src.CreatorId)
	}

	if gormWorld.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", gormWorld.Name, src.Name)
	}

	if gormWorld.Description != src.Description {
		t.Errorf("Description mismatch: got %s, want %s", gormWorld.Description, src.Description)
	}

	if !gormWorld.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", gormWorld.CreatedAt, createdAt)
	}

	if !gormWorld.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", gormWorld.UpdatedAt, updatedAt)
	}

	// Verify Tags
	if len(gormWorld.Tags) != len(src.Tags) {
		t.Fatalf("Tags length mismatch: got %d, want %d", len(gormWorld.Tags), len(src.Tags))
	}

	for i, tag := range src.Tags {
		if gormWorld.Tags[i] != tag {
			t.Errorf("Tags[%d] mismatch: got %s, want %s", i, gormWorld.Tags[i], tag)
		}
	}

	// Verify PreviewUrls
	if len(gormWorld.PreviewUrls) != len(src.PreviewUrls) {
		t.Fatalf("PreviewUrls length mismatch: got %d, want %d", len(gormWorld.PreviewUrls), len(src.PreviewUrls))
	}

	// Verify nested WorldData (2nd level)
	if len(gormWorld.WorldData.Tiles) != len(src.WorldData.Tiles) {
		t.Fatalf("WorldData.Tiles length mismatch: got %d, want %d",
			len(gormWorld.WorldData.Tiles), len(src.WorldData.Tiles))
	}

	if len(gormWorld.WorldData.Units) != len(src.WorldData.Units) {
		t.Fatalf("WorldData.Units length mismatch: got %d, want %d",
			len(gormWorld.WorldData.Units), len(src.WorldData.Units))
	}

	// Verify a sample Tile (3rd level)
	if gormWorld.WorldData.Tiles[0].Q != src.WorldData.Tiles[0].Q {
		t.Errorf("WorldData.Tiles[0].Q mismatch: got %d, want %d",
			gormWorld.WorldData.Tiles[0].Q, src.WorldData.Tiles[0].Q)
	}

	// Verify a sample Unit (3rd level)
	if gormWorld.WorldData.Units[0].Player != src.WorldData.Units[0].Player {
		t.Errorf("WorldData.Units[0].Player mismatch: got %d, want %d",
			gormWorld.WorldData.Units[0].Player, src.WorldData.Units[0].Player)
	}

	// Convert back to API
	apiWorld, err := gorm.WorldFromWorldGORM(nil, gormWorld, nil)
	if err != nil {
		t.Fatalf("WorldFromWorldGORM failed: %v", err)
	}

	// Verify round-trip
	if apiWorld.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %s, want %s", apiWorld.Id, src.Id)
	}

	if apiWorld.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiWorld.Name, src.Name)
	}

	if !apiWorld.CreatedAt.AsTime().Equal(createdAt) {
		t.Errorf("Round-trip CreatedAt mismatch: got %v, want %v", apiWorld.CreatedAt.AsTime(), createdAt)
	}

	// Verify nested WorldData round-trip
	if apiWorld.WorldData == nil {
		t.Fatal("Round-trip: WorldData is nil")
	}

	if len(apiWorld.WorldData.Tiles) != len(src.WorldData.Tiles) {
		t.Fatalf("Round-trip WorldData.Tiles length mismatch: got %d, want %d",
			len(apiWorld.WorldData.Tiles), len(src.WorldData.Tiles))
	}

	if len(apiWorld.WorldData.Units) != len(src.WorldData.Units) {
		t.Fatalf("Round-trip WorldData.Units length mismatch: got %d, want %d",
			len(apiWorld.WorldData.Units), len(src.WorldData.Units))
	}
}

// TestGameMoveConversion_RepeatedAnyField verifies that GameMove with repeated
// google.protobuf.Any (Changes field) converts correctly to [][]byte.
func TestGameMoveConversion_RepeatedAnyField(t *testing.T) {
	timestamp := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	// Create some WorldChange messages with oneof fields
	// WorldChange uses oneof, so we create UnitMovedChange and UnitDamagedChange
	previousUnit := &api.Unit{
		Q: 1, R: 2, Player: 1, UnitType: 100, AvailableHealth: 100, DistanceLeft: 5.0,
	}

	movedUnit := &api.Unit{
		Q: 2, R: 2, Player: 1, UnitType: 100, AvailableHealth: 100, DistanceLeft: 4.0,
	}

	damagedUnit := &api.Unit{
		Q: 1, R: 2, Player: 1, UnitType: 100, AvailableHealth: 80, DistanceLeft: 5.0,
	}

	change1 := &api.WorldChange{
		ChangeType: &api.WorldChange_UnitMoved{
			UnitMoved: &api.UnitMovedChange{
				PreviousUnit: previousUnit,
				UpdatedUnit:  movedUnit,
			},
		},
	}

	change2 := &api.WorldChange{
		ChangeType: &api.WorldChange_UnitDamaged{
			UnitDamaged: &api.UnitDamagedChange{
				PreviousUnit: previousUnit,
				UpdatedUnit:  damagedUnit,
			},
		},
	}

	src := &api.GameMove{
		Player:      1,
		Timestamp:   timestamppb.New(timestamp),
		SequenceNum: 42,
		IsPermanent: true,
		Changes:     []*api.WorldChange{change1, change2},
		// Note: We're not testing the oneof move_type field here as it requires
		// special handling that may need decorator functions
	}

	// Convert to GORM
	gormMove, err := gorm.GameMoveToGameMoveGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("GameMoveToGameMoveGORM failed: %v", err)
	}

	// Verify basic fields
	if gormMove.Player != src.Player {
		t.Errorf("Player mismatch: got %d, want %d", gormMove.Player, src.Player)
	}

	if !gormMove.Timestamp.Equal(timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", gormMove.Timestamp, timestamp)
	}

	if gormMove.SequenceNum != src.SequenceNum {
		t.Errorf("SequenceNum mismatch: got %d, want %d", gormMove.SequenceNum, src.SequenceNum)
	}

	if gormMove.IsPermanent != src.IsPermanent {
		t.Errorf("IsPermanent mismatch: got %v, want %v", gormMove.IsPermanent, src.IsPermanent)
	}

	// Verify Changes (repeated *WorldChange → [][]byte)
	if len(gormMove.Changes) != len(src.Changes) {
		t.Fatalf("Changes length mismatch: got %d, want %d", len(gormMove.Changes), len(src.Changes))
	}

	for i, changeBytes := range gormMove.Changes {
		if len(changeBytes) == 0 {
			t.Errorf("Changes[%d] is empty", i)
		}
	}

	// Convert back to API
	apiMove, err := gorm.GameMoveFromGameMoveGORM(nil, gormMove, nil)
	if err != nil {
		t.Fatalf("GameMoveFromGameMoveGORM failed: %v", err)
	}

	// Verify round-trip
	if apiMove.Player != src.Player {
		t.Errorf("Round-trip Player mismatch: got %d, want %d", apiMove.Player, src.Player)
	}

	if !apiMove.Timestamp.AsTime().Equal(timestamp) {
		t.Errorf("Round-trip Timestamp mismatch: got %v, want %v", apiMove.Timestamp.AsTime(), timestamp)
	}

	// Verify Changes round-trip
	if len(apiMove.Changes) != len(src.Changes) {
		t.Fatalf("Round-trip Changes length mismatch: got %d, want %d", len(apiMove.Changes), len(src.Changes))
	}

	// Verify the unpacked WorldChange messages (check the oneof is present)
	for i, change := range apiMove.Changes {
		if change.GetChangeType() == nil {
			t.Errorf("Round-trip Changes[%d] has nil ChangeType", i)
		}
	}

	// Verify first change is UnitMoved
	if unitMoved := apiMove.Changes[0].GetUnitMoved(); unitMoved != nil {
		if unitMoved.UpdatedUnit.Q != movedUnit.Q || unitMoved.UpdatedUnit.R != movedUnit.R {
			t.Errorf("Round-trip Changes[0] UnitMoved position mismatch: got (%d,%d), want (%d,%d)",
				unitMoved.UpdatedUnit.Q, unitMoved.UpdatedUnit.R, movedUnit.Q, movedUnit.R)
		}
	} else {
		t.Error("Round-trip Changes[0] is not UnitMoved")
	}

	// Verify second change is UnitDamaged
	if unitDamaged := apiMove.Changes[1].GetUnitDamaged(); unitDamaged != nil {
		if unitDamaged.UpdatedUnit.AvailableHealth != damagedUnit.AvailableHealth {
			t.Errorf("Round-trip Changes[1] UnitDamaged health mismatch: got %d, want %d",
				unitDamaged.UpdatedUnit.AvailableHealth, damagedUnit.AvailableHealth)
		}
	} else {
		t.Error("Round-trip Changes[1] is not UnitDamaged")
	}
}

// TestIndexInfoConversion_NestedTimestamps verifies that IndexInfo with multiple
// timestamp fields converts correctly.
func TestIndexInfoConversion_NestedTimestamps(t *testing.T) {
	lastUpdatedAt := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	lastIndexedAt := time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)

	src := &api.IndexInfo{
		LastUpdatedAt:  timestamppb.New(lastUpdatedAt),
		LastIndexedAt:  timestamppb.New(lastIndexedAt),
		NeedsIndexing: true,
	}

	// Convert to GORM
	gormInfo, err := gorm.IndexInfoToIndexInfoGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("IndexInfoToIndexInfoGORM failed: %v", err)
	}

	// Verify fields
	if !gormInfo.LastUpdatedAt.Equal(lastUpdatedAt) {
		t.Errorf("LastUpdatedAt mismatch: got %v, want %v", gormInfo.LastUpdatedAt, lastUpdatedAt)
	}

	if !gormInfo.LastIndexedAt.Equal(lastIndexedAt) {
		t.Errorf("LastIndexedAt mismatch: got %v, want %v", gormInfo.LastIndexedAt, lastIndexedAt)
	}

	if gormInfo.NeedsIndexing != src.NeedsIndexing {
		t.Errorf("NeedsIndexing mismatch: got %v, want %v", gormInfo.NeedsIndexing, src.NeedsIndexing)
	}

	// Convert back to API
	apiInfo, err := gorm.IndexInfoFromIndexInfoGORM(nil, gormInfo, nil)
	if err != nil {
		t.Fatalf("IndexInfoFromIndexInfoGORM failed: %v", err)
	}

	// Verify round-trip
	if !apiInfo.LastUpdatedAt.AsTime().Equal(lastUpdatedAt) {
		t.Errorf("Round-trip LastUpdatedAt mismatch: got %v, want %v", apiInfo.LastUpdatedAt.AsTime(), lastUpdatedAt)
	}

	if !apiInfo.LastIndexedAt.AsTime().Equal(lastIndexedAt) {
		t.Errorf("Round-trip LastIndexedAt mismatch: got %v, want %v", apiInfo.LastIndexedAt.AsTime(), lastIndexedAt)
	}

	if apiInfo.NeedsIndexing != src.NeedsIndexing {
		t.Errorf("Round-trip NeedsIndexing mismatch: got %v, want %v", apiInfo.NeedsIndexing, src.NeedsIndexing)
	}
}

// TestEmptyWorldData_NilSlices verifies that empty/nil slices are handled correctly
// in nested structures.
func TestEmptyWorldData_NilSlices(t *testing.T) {
	testCases := []struct {
		name string
		data *api.WorldData
	}{
		{"NilSlices", &api.WorldData{Tiles: nil, Units: nil}},
		{"EmptySlices", &api.WorldData{Tiles: []*api.Tile{}, Units: []*api.Unit{}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to GORM
			gormData, err := gorm.WorldDataToWorldDataGORM(tc.data, nil, nil)
			if err != nil {
				t.Fatalf("WorldDataToWorldDataGORM failed: %v", err)
			}

			// Check slices
			if tc.data.Tiles == nil && gormData.Tiles != nil {
				t.Errorf("Expected nil Tiles, got %v", gormData.Tiles)
			}

			if tc.data.Tiles != nil && len(gormData.Tiles) != 0 {
				t.Errorf("Expected empty Tiles, got length %d", len(gormData.Tiles))
			}

			// Convert back
			apiData, err := gorm.WorldDataFromWorldDataGORM(nil, gormData, nil)
			if err != nil {
				t.Fatalf("WorldDataFromWorldDataGORM failed: %v", err)
			}

			// Verify round-trip
			if tc.data.Tiles == nil && apiData.Tiles != nil {
				t.Errorf("Round-trip: Expected nil Tiles, got %v", apiData.Tiles)
			}

			if tc.data.Units == nil && apiData.Units != nil {
				t.Errorf("Round-trip: Expected nil Units, got %v", apiData.Units)
			}
		})
	}
}
