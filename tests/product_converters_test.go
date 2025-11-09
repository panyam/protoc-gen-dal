package tests

import (
	"reflect"
	"testing"

	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"github.com/panyam/protoc-gen-dal/tests/gen/gorm"
)

// TestProductConversion_RepeatedAndMapFields verifies that repeated primitives
// and map<K, primitive> fields are correctly converted using direct assignment.
func TestProductConversion_RepeatedAndMapFields(t *testing.T) {
	// Create a source Product with all fields populated
	src := &api.Product{
		Id:   1,
		Name: "Test Product",
		Tags: []string{"electronics", "smartphone", "5G"},
		Categories: []string{"tech", "gadgets", "mobile"},
		Metadata: map[string]string{
			"color":   "black",
			"storage": "128GB",
			"brand":   "TestBrand",
		},
		Ratings: []int32{5, 4, 5, 3, 4},
	}

	// Convert to GORM
	gormProduct, err := gorm.ProductToProductGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductGORM failed: %v", err)
	}

	// Verify all fields were converted correctly
	if gormProduct.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormProduct.Id, src.Id)
	}

	if gormProduct.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", gormProduct.Name, src.Name)
	}

	// Verify Tags (repeated string)
	if len(gormProduct.Tags) != len(src.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(gormProduct.Tags), len(src.Tags))
	}
	for i, tag := range src.Tags {
		if gormProduct.Tags[i] != tag {
			t.Errorf("Tags[%d] mismatch: got %s, want %s", i, gormProduct.Tags[i], tag)
		}
	}

	// Verify Categories (repeated string)
	if len(gormProduct.Categories) != len(src.Categories) {
		t.Errorf("Categories length mismatch: got %d, want %d", len(gormProduct.Categories), len(src.Categories))
	}
	for i, cat := range src.Categories {
		if gormProduct.Categories[i] != cat {
			t.Errorf("Categories[%d] mismatch: got %s, want %s", i, gormProduct.Categories[i], cat)
		}
	}

	// Verify Metadata (map<string, string>)
	if len(gormProduct.Metadata) != len(src.Metadata) {
		t.Errorf("Metadata length mismatch: got %d, want %d", len(gormProduct.Metadata), len(src.Metadata))
	}
	for key, value := range src.Metadata {
		if gormProduct.Metadata[key] != value {
			t.Errorf("Metadata[%s] mismatch: got %s, want %s", key, gormProduct.Metadata[key], value)
		}
	}

	// Verify Ratings (repeated int32)
	if len(gormProduct.Ratings) != len(src.Ratings) {
		t.Errorf("Ratings length mismatch: got %d, want %d", len(gormProduct.Ratings), len(src.Ratings))
	}
	for i, rating := range src.Ratings {
		if gormProduct.Ratings[i] != rating {
			t.Errorf("Ratings[%d] mismatch: got %d, want %d", i, gormProduct.Ratings[i], rating)
		}
	}

	// Convert back to API
	apiProduct, err := gorm.ProductFromProductGORM(nil, gormProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductGORM failed: %v", err)
	}

	// Verify round-trip conversion
	if apiProduct.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiProduct.Id, src.Id)
	}

	if apiProduct.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiProduct.Name, src.Name)
	}

	// Verify Tags round-trip
	if len(apiProduct.Tags) != len(src.Tags) {
		t.Errorf("Round-trip Tags length mismatch: got %d, want %d", len(apiProduct.Tags), len(src.Tags))
	}
	for i, tag := range src.Tags {
		if apiProduct.Tags[i] != tag {
			t.Errorf("Round-trip Tags[%d] mismatch: got %s, want %s", i, apiProduct.Tags[i], tag)
		}
	}

	// Verify Categories round-trip
	if len(apiProduct.Categories) != len(src.Categories) {
		t.Errorf("Round-trip Categories length mismatch: got %d, want %d", len(apiProduct.Categories), len(src.Categories))
	}
	for i, cat := range src.Categories {
		if apiProduct.Categories[i] != cat {
			t.Errorf("Round-trip Categories[%d] mismatch: got %s, want %s", i, apiProduct.Categories[i], cat)
		}
	}

	// Verify Metadata round-trip
	if len(apiProduct.Metadata) != len(src.Metadata) {
		t.Errorf("Round-trip Metadata length mismatch: got %d, want %d", len(apiProduct.Metadata), len(src.Metadata))
	}
	for key, value := range src.Metadata {
		if apiProduct.Metadata[key] != value {
			t.Errorf("Round-trip Metadata[%s] mismatch: got %s, want %s", key, apiProduct.Metadata[key], value)
		}
	}

	// Verify Ratings round-trip
	if len(apiProduct.Ratings) != len(src.Ratings) {
		t.Errorf("Round-trip Ratings length mismatch: got %d, want %d", len(apiProduct.Ratings), len(src.Ratings))
	}
	for i, rating := range src.Ratings {
		if apiProduct.Ratings[i] != rating {
			t.Errorf("Round-trip Ratings[%d] mismatch: got %d, want %d", i, apiProduct.Ratings[i], rating)
		}
	}
}

// TestProductConversion_NilMetadata verifies that nil maps are handled correctly.
func TestProductConversion_NilMetadata(t *testing.T) {
	src := &api.Product{
		Id:       2,
		Name:     "Product without metadata",
		Tags:     []string{"test"},
		Metadata: nil, // nil map
	}

	gormProduct, err := gorm.ProductToProductGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductGORM failed: %v", err)
	}

	// Nil map should result in nil map (not panic)
	if gormProduct.Metadata != nil {
		t.Errorf("Expected nil Metadata, got %v", gormProduct.Metadata)
	}

	// Convert back
	apiProduct, err := gorm.ProductFromProductGORM(nil, gormProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductGORM failed: %v", err)
	}

	if apiProduct.Metadata != nil {
		t.Errorf("Round-trip: Expected nil Metadata, got %v", apiProduct.Metadata)
	}
}

// TestProductConversion_EmptyCollections verifies empty slices and maps are handled correctly.
func TestProductConversion_EmptyCollections(t *testing.T) {
	src := &api.Product{
		Id:         3,
		Name:       "Product with empty collections",
		Tags:       []string{},           // empty slice
		Categories: []string{},           // empty slice
		Metadata:   map[string]string{}, // empty map
		Ratings:    []int32{},            // empty slice
	}

	gormProduct, err := gorm.ProductToProductGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductGORM failed: %v", err)
	}

	// Empty collections should remain empty
	if len(gormProduct.Tags) != 0 {
		t.Errorf("Expected empty Tags, got %v", gormProduct.Tags)
	}
	if len(gormProduct.Categories) != 0 {
		t.Errorf("Expected empty Categories, got %v", gormProduct.Categories)
	}
	if len(gormProduct.Metadata) != 0 {
		t.Errorf("Expected empty Metadata, got %v", gormProduct.Metadata)
	}
	if len(gormProduct.Ratings) != 0 {
		t.Errorf("Expected empty Ratings, got %v", gormProduct.Ratings)
	}

	// Convert back and verify
	apiProduct, err := gorm.ProductFromProductGORM(nil, gormProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductGORM failed: %v", err)
	}

	if len(apiProduct.Tags) != 0 {
		t.Errorf("Round-trip: Expected empty Tags, got %v", apiProduct.Tags)
	}
	if len(apiProduct.Categories) != 0 {
		t.Errorf("Round-trip: Expected empty Categories, got %v", apiProduct.Categories)
	}
	if len(apiProduct.Metadata) != 0 {
		t.Errorf("Round-trip: Expected empty Metadata, got %v", apiProduct.Metadata)
	}
	if len(apiProduct.Ratings) != 0 {
		t.Errorf("Round-trip: Expected empty Ratings, got %v", apiProduct.Ratings)
	}
}

// TestLibraryConversion_RepeatedMessageType verifies that repeated message types
// are correctly converted using loop-based converter application.
func TestLibraryConversion_RepeatedMessageType(t *testing.T) {
	// Create a source Library with multiple contributors
	src := &api.Library{
		Id:   1,
		Name: "Tech Library",
		Contributors: []*api.Author{
			{Name: "Alice Smith", Email: "alice@example.com"},
			{Name: "Bob Jones", Email: "bob@example.com"},
			{Name: "Carol White", Email: "carol@example.com"},
		},
	}

	// Convert to GORM
	gormLibrary, err := gorm.LibraryToLibraryGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("LibraryToLibraryGORM failed: %v", err)
	}

	// Verify basic fields
	if gormLibrary.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormLibrary.Id, src.Id)
	}

	if gormLibrary.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", gormLibrary.Name, src.Name)
	}

	// Verify Contributors (repeated message type)
	if len(gormLibrary.Contributors) != len(src.Contributors) {
		t.Fatalf("Contributors length mismatch: got %d, want %d", len(gormLibrary.Contributors), len(src.Contributors))
	}

	// Check each contributor was converted correctly
	for i, srcAuthor := range src.Contributors {
		gormAuthor := gormLibrary.Contributors[i]
		if gormAuthor.Name != srcAuthor.Name {
			t.Errorf("Contributors[%d].Name mismatch: got %s, want %s", i, gormAuthor.Name, srcAuthor.Name)
		}
		if gormAuthor.Email != srcAuthor.Email {
			t.Errorf("Contributors[%d].Email mismatch: got %s, want %s", i, gormAuthor.Email, srcAuthor.Email)
		}
	}

	// Convert back to API
	apiLibrary, err := gorm.LibraryFromLibraryGORM(nil, gormLibrary, nil)
	if err != nil {
		t.Fatalf("LibraryFromLibraryGORM failed: %v", err)
	}

	// Verify round-trip conversion
	if apiLibrary.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiLibrary.Id, src.Id)
	}

	if apiLibrary.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiLibrary.Name, src.Name)
	}

	// Verify Contributors round-trip
	if len(apiLibrary.Contributors) != len(src.Contributors) {
		t.Fatalf("Round-trip Contributors length mismatch: got %d, want %d", len(apiLibrary.Contributors), len(src.Contributors))
	}

	for i, srcAuthor := range src.Contributors {
		apiAuthor := apiLibrary.Contributors[i]
		if apiAuthor.Name != srcAuthor.Name {
			t.Errorf("Round-trip Contributors[%d].Name mismatch: got %s, want %s", i, apiAuthor.Name, srcAuthor.Name)
		}
		if apiAuthor.Email != srcAuthor.Email {
			t.Errorf("Round-trip Contributors[%d].Email mismatch: got %s, want %s", i, apiAuthor.Email, srcAuthor.Email)
		}
	}
}

// TestLibraryConversion_EmptyContributors verifies empty repeated message slices work.
func TestLibraryConversion_EmptyContributors(t *testing.T) {
	src := &api.Library{
		Id:           2,
		Name:         "Empty Library",
		Contributors: []*api.Author{}, // empty slice
	}

	gormLibrary, err := gorm.LibraryToLibraryGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("LibraryToLibraryGORM failed: %v", err)
	}

	if len(gormLibrary.Contributors) != 0 {
		t.Errorf("Expected empty Contributors, got %v", gormLibrary.Contributors)
	}

	// Convert back
	apiLibrary, err := gorm.LibraryFromLibraryGORM(nil, gormLibrary, nil)
	if err != nil {
		t.Fatalf("LibraryFromLibraryGORM failed: %v", err)
	}

	if !reflect.DeepEqual(apiLibrary.Contributors, src.Contributors) {
		t.Errorf("Round-trip: Expected empty Contributors, got %v", apiLibrary.Contributors)
	}
}

// TestLibraryConversion_NilContributors verifies nil repeated message slices work.
func TestLibraryConversion_NilContributors(t *testing.T) {
	src := &api.Library{
		Id:           3,
		Name:         "Nil Library",
		Contributors: nil, // nil slice
	}

	gormLibrary, err := gorm.LibraryToLibraryGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("LibraryToLibraryGORM failed: %v", err)
	}

	if gormLibrary.Contributors != nil {
		t.Errorf("Expected nil Contributors, got %v", gormLibrary.Contributors)
	}

	// Convert back
	apiLibrary, err := gorm.LibraryFromLibraryGORM(nil, gormLibrary, nil)
	if err != nil {
		t.Fatalf("LibraryFromLibraryGORM failed: %v", err)
	}

	if apiLibrary.Contributors != nil {
		t.Errorf("Round-trip: Expected nil Contributors, got %v", apiLibrary.Contributors)
	}
}
