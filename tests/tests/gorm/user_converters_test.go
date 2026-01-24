package gorm

import (
	"reflect"
	"testing"
	"time"

	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"github.com/panyam/protoc-gen-dal/tests/gen/gorm/gorm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestUserConversion_BasicFieldsAndTimestamps verifies that basic fields and
// google.protobuf.Timestamp fields are correctly converted to GORM time.Time fields.
func TestUserConversion_BasicFieldsAndTimestamps(t *testing.T) {
	// Create realistic timestamps
	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	activatedAt := time.Date(2020, 1, 10, 14, 30, 0, 0, time.UTC)
	createdAt := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 16, 45, 30, 0, time.UTC)

	src := &api.User{
		Id:           1,
		Name:         "Alice Smith",
		Email:        "alice@example.com",
		Age:          34,
		Birthday:     timestamppb.New(birthday),
		MemberNumber: "MEM-12345",
		ActivatedAt:  timestamppb.New(activatedAt),
		CreatedAt:    timestamppb.New(createdAt),
		UpdatedAt:    timestamppb.New(updatedAt),
	}

	// Convert to GORM
	gormUser, err := gorm.UserToUserGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserGORM failed: %v", err)
	}

	// Verify basic fields
	if gormUser.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormUser.Id, src.Id)
	}

	if gormUser.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", gormUser.Name, src.Name)
	}

	if gormUser.Email != src.Email {
		t.Errorf("Email mismatch: got %s, want %s", gormUser.Email, src.Email)
	}

	if gormUser.Age != src.Age {
		t.Errorf("Age mismatch: got %d, want %d", gormUser.Age, src.Age)
	}

	if gormUser.MemberNumber != src.MemberNumber {
		t.Errorf("MemberNumber mismatch: got %s, want %s", gormUser.MemberNumber, src.MemberNumber)
	}

	// Verify timestamp conversions
	if !gormUser.Birthday.Equal(birthday) {
		t.Errorf("Birthday mismatch: got %v, want %v", gormUser.Birthday, birthday)
	}

	if !gormUser.ActivatedAt.Equal(activatedAt) {
		t.Errorf("ActivatedAt mismatch: got %v, want %v", gormUser.ActivatedAt, activatedAt)
	}

	if !gormUser.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", gormUser.CreatedAt, createdAt)
	}

	if !gormUser.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", gormUser.UpdatedAt, updatedAt)
	}

	// Convert back to API
	apiUser, err := gorm.UserFromUserGORM(nil, gormUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserGORM failed: %v", err)
	}

	// Verify round-trip conversion for basic fields
	if apiUser.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiUser.Id, src.Id)
	}

	if apiUser.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiUser.Name, src.Name)
	}

	if apiUser.Email != src.Email {
		t.Errorf("Round-trip Email mismatch: got %s, want %s", apiUser.Email, src.Email)
	}

	if apiUser.Age != src.Age {
		t.Errorf("Round-trip Age mismatch: got %d, want %d", apiUser.Age, src.Age)
	}

	if apiUser.MemberNumber != src.MemberNumber {
		t.Errorf("Round-trip MemberNumber mismatch: got %s, want %s", apiUser.MemberNumber, src.MemberNumber)
	}

	// Verify round-trip timestamp conversions (should preserve time to second precision)
	if !apiUser.Birthday.AsTime().Equal(birthday) {
		t.Errorf("Round-trip Birthday mismatch: got %v, want %v", apiUser.Birthday.AsTime(), birthday)
	}

	if !apiUser.ActivatedAt.AsTime().Equal(activatedAt) {
		t.Errorf("Round-trip ActivatedAt mismatch: got %v, want %v", apiUser.ActivatedAt.AsTime(), activatedAt)
	}

	if !apiUser.CreatedAt.AsTime().Equal(createdAt) {
		t.Errorf("Round-trip CreatedAt mismatch: got %v, want %v", apiUser.CreatedAt.AsTime(), createdAt)
	}

	if !apiUser.UpdatedAt.AsTime().Equal(updatedAt) {
		t.Errorf("Round-trip UpdatedAt mismatch: got %v, want %v", apiUser.UpdatedAt.AsTime(), updatedAt)
	}
}

// TestUserConversion_NilTimestamps verifies that nil timestamp fields are handled correctly.
func TestUserConversion_NilTimestamps(t *testing.T) {
	src := &api.User{
		Id:           2,
		Name:         "Bob Jones",
		Email:        "bob@example.com",
		Age:          25,
		Birthday:     nil, // nil timestamp
		MemberNumber: "MEM-67890",
		ActivatedAt:  nil,
		CreatedAt:    nil,
		UpdatedAt:    nil,
	}

	gormUser, err := gorm.UserToUserGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserGORM failed: %v", err)
	}

	// Nil timestamps should result in zero time values
	if !gormUser.Birthday.IsZero() {
		t.Errorf("Expected zero Birthday, got %v", gormUser.Birthday)
	}

	if !gormUser.ActivatedAt.IsZero() {
		t.Errorf("Expected zero ActivatedAt, got %v", gormUser.ActivatedAt)
	}

	if !gormUser.CreatedAt.IsZero() {
		t.Errorf("Expected zero CreatedAt, got %v", gormUser.CreatedAt)
	}

	if !gormUser.UpdatedAt.IsZero() {
		t.Errorf("Expected zero UpdatedAt, got %v", gormUser.UpdatedAt)
	}

	// Convert back
	apiUser, err := gorm.UserFromUserGORM(nil, gormUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserGORM failed: %v", err)
	}

	// Zero time should convert back to nil timestamp (or zero timestamp)
	// The converters.TimeToTimestamp handles zero time appropriately
	if apiUser.Birthday != nil && !apiUser.Birthday.AsTime().IsZero() {
		t.Errorf("Round-trip: Expected nil or zero Birthday, got %v", apiUser.Birthday.AsTime())
	}
}

// TestUserConversion_WithCustomTimestamps verifies that custom timestamp types (int64) work.
func TestUserConversion_WithCustomTimestamps(t *testing.T) {
	createdAt := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)

	src := &api.User{
		Id:          3,
		Name:        "Carol White",
		Email:       "carol@example.com",
		Age:         28,
		CreatedAt:   timestamppb.New(createdAt),
		Birthday:    timestamppb.New(birthday),
		ActivatedAt: nil,
		UpdatedAt:   nil,
	}

	// Convert to UserWithCustomTimestamps
	gormUser, err := gorm.UserToUserWithCustomTimestamps(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserWithCustomTimestamps failed: %v", err)
	}

	// Verify CreatedAt is int64 (Unix timestamp in seconds)
	expectedCreatedAtInt := createdAt.Unix()
	if gormUser.CreatedAt != expectedCreatedAtInt {
		t.Errorf("CreatedAt mismatch: got %d, want %d", gormUser.CreatedAt, expectedCreatedAtInt)
	}

	// Birthday should still be time.Time
	if !gormUser.Birthday.Equal(birthday) {
		t.Errorf("Birthday mismatch: got %v, want %v", gormUser.Birthday, birthday)
	}

	// Convert back to API
	apiUser, err := gorm.UserFromUserWithCustomTimestamps(nil, gormUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserWithCustomTimestamps failed: %v", err)
	}

	// Verify round-trip
	if !apiUser.CreatedAt.AsTime().Equal(createdAt) {
		t.Errorf("Round-trip CreatedAt mismatch: got %v, want %v", apiUser.CreatedAt.AsTime(), createdAt)
	}

	if !apiUser.Birthday.AsTime().Equal(birthday) {
		t.Errorf("Round-trip Birthday mismatch: got %v, want %v", apiUser.Birthday.AsTime(), birthday)
	}
}

// TestAuthorConversion_SimpleStruct verifies simple struct conversion without timestamps.
func TestAuthorConversion_SimpleStruct(t *testing.T) {
	src := &api.Author{
		Name:  "Alice Smith",
		Email: "alice@example.com",
	}

	// Convert to GORM
	gormAuthor, err := gorm.AuthorToAuthorGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("AuthorToAuthorGORM failed: %v", err)
	}

	// Verify fields
	if gormAuthor.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", gormAuthor.Name, src.Name)
	}

	if gormAuthor.Email != src.Email {
		t.Errorf("Email mismatch: got %s, want %s", gormAuthor.Email, src.Email)
	}

	// Convert back to API
	apiAuthor, err := gorm.AuthorFromAuthorGORM(nil, gormAuthor, nil)
	if err != nil {
		t.Fatalf("AuthorFromAuthorGORM failed: %v", err)
	}

	// Verify round-trip
	if apiAuthor.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiAuthor.Name, src.Name)
	}

	if apiAuthor.Email != src.Email {
		t.Errorf("Round-trip Email mismatch: got %s, want %s", apiAuthor.Email, src.Email)
	}
}

// TestBlogConversion_NestedMessage verifies that nested message fields are correctly converted.
func TestBlogConversion_NestedMessage(t *testing.T) {
	src := &api.Blog{
		Id: 1,
		Author: &api.Author{
			Name:  "Alice Smith",
			Email: "alice@example.com",
		},
		Upvotes: 42,
		Title:   "Understanding Protobuf Converters",
	}

	// Convert to BlogGORM
	gormBlog, err := gorm.BlogToBlogGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("BlogToBlogGORM failed: %v", err)
	}

	// Verify basic fields
	if gormBlog.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormBlog.Id, src.Id)
	}

	if gormBlog.Upvotes != src.Upvotes {
		t.Errorf("Upvotes mismatch: got %d, want %d", gormBlog.Upvotes, src.Upvotes)
	}

	if gormBlog.Title != src.Title {
		t.Errorf("Title mismatch: got %s, want %s", gormBlog.Title, src.Title)
	}

	// Verify nested Author field
	if gormBlog.Author.Name != src.Author.Name {
		t.Errorf("Author.Name mismatch: got %s, want %s", gormBlog.Author.Name, src.Author.Name)
	}

	if gormBlog.Author.Email != src.Author.Email {
		t.Errorf("Author.Email mismatch: got %s, want %s", gormBlog.Author.Email, src.Author.Email)
	}

	// Convert back to API
	apiBlog, err := gorm.BlogFromBlogGORM(nil, gormBlog, nil)
	if err != nil {
		t.Fatalf("BlogFromBlogGORM failed: %v", err)
	}

	// Verify round-trip
	if apiBlog.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiBlog.Id, src.Id)
	}

	if apiBlog.Upvotes != src.Upvotes {
		t.Errorf("Round-trip Upvotes mismatch: got %d, want %d", apiBlog.Upvotes, src.Upvotes)
	}

	if apiBlog.Title != src.Title {
		t.Errorf("Round-trip Title mismatch: got %s, want %s", apiBlog.Title, src.Title)
	}

	if apiBlog.Author == nil {
		t.Fatal("Round-trip Author is nil")
	}

	if apiBlog.Author.Name != src.Author.Name {
		t.Errorf("Round-trip Author.Name mismatch: got %s, want %s", apiBlog.Author.Name, src.Author.Name)
	}

	if apiBlog.Author.Email != src.Author.Email {
		t.Errorf("Round-trip Author.Email mismatch: got %s, want %s", apiBlog.Author.Email, src.Author.Email)
	}
}

// TestBlogConversion_NilAuthor verifies that nil nested message is handled correctly.
func TestBlogConversion_NilAuthor(t *testing.T) {
	src := &api.Blog{
		Id:      2,
		Author:  nil, // nil nested message
		Upvotes: 10,
		Title:   "Anonymous Blog Post",
	}

	gormBlog, err := gorm.BlogToBlogGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("BlogToBlogGORM failed: %v", err)
	}

	// Verify basic fields
	if gormBlog.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormBlog.Id, src.Id)
	}

	// The Author field should have zero values (empty struct)
	// since GORM embeds it with embeddedPrefix
	if gormBlog.Author.Name != "" || gormBlog.Author.Email != "" {
		t.Errorf("Expected zero Author, got Name=%s, Email=%s", gormBlog.Author.Name, gormBlog.Author.Email)
	}

	// Convert back
	apiBlog, err := gorm.BlogFromBlogGORM(nil, gormBlog, nil)
	if err != nil {
		t.Fatalf("BlogFromBlogGORM failed: %v", err)
	}

	// Round-trip should preserve nil or create empty Author
	// Depending on implementation, Author might be nil or empty struct
	if apiBlog.Author != nil && (apiBlog.Author.Name != "" || apiBlog.Author.Email != "") {
		t.Errorf("Round-trip: Expected nil or empty Author, got %+v", apiBlog.Author)
	}
}

// TestProductConversion_RepeatedAndMapFields verifies that repeated primitives
// and map<K, primitive> fields are correctly converted using direct assignment.
// This test is moved from the root tests package.
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

// TestOrganizationConversion_MapMessageType verifies that map<K, MessageType> fields
// are correctly converted using loop-based converter application.
func TestOrganizationConversion_MapMessageType(t *testing.T) {
	// Create a source Organization with multiple departments
	src := &api.Organization{
		Id:   1,
		Name: "Tech Corp",
		Departments: map[string]*api.Author{
			"engineering": {Name: "Alice Smith", Email: "alice@example.com"},
			"marketing":   {Name: "Bob Jones", Email: "bob@example.com"},
			"sales":       {Name: "Carol White", Email: "carol@example.com"},
		},
	}

	// Convert to GORM
	gormOrg, err := gorm.OrganizationToOrganizationGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationGORM failed: %v", err)
	}

	// Verify basic fields
	if gormOrg.Id != src.Id {
		t.Errorf("Id mismatch: got %d, want %d", gormOrg.Id, src.Id)
	}

	if gormOrg.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", gormOrg.Name, src.Name)
	}

	// Verify Departments (map<string, MessageType>)
	if len(gormOrg.Departments) != len(src.Departments) {
		t.Fatalf("Departments length mismatch: got %d, want %d", len(gormOrg.Departments), len(src.Departments))
	}

	// Check each department was converted correctly
	for key, srcAuthor := range src.Departments {
		gormAuthor, exists := gormOrg.Departments[key]
		if !exists {
			t.Errorf("Department %s missing in GORM struct", key)
			continue
		}
		if gormAuthor.Name != srcAuthor.Name {
			t.Errorf("Departments[%s].Name mismatch: got %s, want %s", key, gormAuthor.Name, srcAuthor.Name)
		}
		if gormAuthor.Email != srcAuthor.Email {
			t.Errorf("Departments[%s].Email mismatch: got %s, want %s", key, gormAuthor.Email, srcAuthor.Email)
		}
	}

	// Convert back to API
	apiOrg, err := gorm.OrganizationFromOrganizationGORM(nil, gormOrg, nil)
	if err != nil {
		t.Fatalf("OrganizationFromOrganizationGORM failed: %v", err)
	}

	// Verify round-trip conversion
	if apiOrg.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiOrg.Id, src.Id)
	}

	if apiOrg.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiOrg.Name, src.Name)
	}

	// Verify Departments round-trip
	if len(apiOrg.Departments) != len(src.Departments) {
		t.Fatalf("Round-trip Departments length mismatch: got %d, want %d", len(apiOrg.Departments), len(src.Departments))
	}

	for key, srcAuthor := range src.Departments {
		apiAuthor, exists := apiOrg.Departments[key]
		if !exists {
			t.Errorf("Round-trip: Department %s missing", key)
			continue
		}
		if apiAuthor.Name != srcAuthor.Name {
			t.Errorf("Round-trip Departments[%s].Name mismatch: got %s, want %s", key, apiAuthor.Name, srcAuthor.Name)
		}
		if apiAuthor.Email != srcAuthor.Email {
			t.Errorf("Round-trip Departments[%s].Email mismatch: got %s, want %s", key, apiAuthor.Email, srcAuthor.Email)
		}
	}
}

// TestOrganizationConversion_EmptyDepartments verifies empty map handling.
func TestOrganizationConversion_EmptyDepartments(t *testing.T) {
	src := &api.Organization{
		Id:          2,
		Name:        "Empty Org",
		Departments: map[string]*api.Author{}, // empty map
	}

	gormOrg, err := gorm.OrganizationToOrganizationGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationGORM failed: %v", err)
	}

	if len(gormOrg.Departments) != 0 {
		t.Errorf("Expected empty Departments, got %v", gormOrg.Departments)
	}

	// Convert back
	apiOrg, err := gorm.OrganizationFromOrganizationGORM(nil, gormOrg, nil)
	if err != nil {
		t.Fatalf("OrganizationFromOrganizationGORM failed: %v", err)
	}

	if !reflect.DeepEqual(apiOrg.Departments, src.Departments) {
		t.Errorf("Round-trip: Expected empty Departments, got %v", apiOrg.Departments)
	}
}

// TestOrganizationConversion_NilDepartments verifies nil map handling.
func TestOrganizationConversion_NilDepartments(t *testing.T) {
	src := &api.Organization{
		Id:          3,
		Name:        "Nil Org",
		Departments: nil, // nil map
	}

	gormOrg, err := gorm.OrganizationToOrganizationGORM(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationGORM failed: %v", err)
	}

	if gormOrg.Departments != nil {
		t.Errorf("Expected nil Departments, got %v", gormOrg.Departments)
	}

	// Convert back
	apiOrg, err := gorm.OrganizationFromOrganizationGORM(nil, gormOrg, nil)
	if err != nil {
		t.Fatalf("OrganizationFromOrganizationGORM failed: %v", err)
	}

	if apiOrg.Departments != nil {
		t.Errorf("Round-trip: Expected nil Departments, got %v", apiOrg.Departments)
	}
}
