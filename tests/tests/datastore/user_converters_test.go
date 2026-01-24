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
	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestUserConversion_BasicFieldsAndTimestamps verifies that basic fields and
// google.protobuf.Timestamp fields are correctly converted to Datastore time.Time fields.
// Note: Datastore converts uint32 ID to string.
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

	// Convert to Datastore
	dsUser, err := datastore.UserToUserDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserDatastore failed: %v", err)
	}

	// Verify ID is converted to string
	if dsUser.Id != "1" {
		t.Errorf("Id mismatch: got %s, want '1'", dsUser.Id)
	}

	// Verify basic fields
	if dsUser.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsUser.Name, src.Name)
	}

	if dsUser.Email != src.Email {
		t.Errorf("Email mismatch: got %s, want %s", dsUser.Email, src.Email)
	}

	if dsUser.Age != src.Age {
		t.Errorf("Age mismatch: got %d, want %d", dsUser.Age, src.Age)
	}

	if dsUser.MemberNumber != src.MemberNumber {
		t.Errorf("MemberNumber mismatch: got %s, want %s", dsUser.MemberNumber, src.MemberNumber)
	}

	// Verify timestamp conversions
	if !dsUser.Birthday.Equal(birthday) {
		t.Errorf("Birthday mismatch: got %v, want %v", dsUser.Birthday, birthday)
	}

	if !dsUser.ActivatedAt.Equal(activatedAt) {
		t.Errorf("ActivatedAt mismatch: got %v, want %v", dsUser.ActivatedAt, activatedAt)
	}

	if !dsUser.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", dsUser.CreatedAt, createdAt)
	}

	if !dsUser.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt mismatch: got %v, want %v", dsUser.UpdatedAt, updatedAt)
	}

	// Convert back to API
	apiUser, err := datastore.UserFromUserDatastore(nil, dsUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserDatastore failed: %v", err)
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

	// Verify round-trip timestamp conversions
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

	dsUser, err := datastore.UserToUserDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserDatastore failed: %v", err)
	}

	// Nil timestamps should result in zero time values
	if !dsUser.Birthday.IsZero() {
		t.Errorf("Expected zero Birthday, got %v", dsUser.Birthday)
	}

	if !dsUser.ActivatedAt.IsZero() {
		t.Errorf("Expected zero ActivatedAt, got %v", dsUser.ActivatedAt)
	}

	if !dsUser.CreatedAt.IsZero() {
		t.Errorf("Expected zero CreatedAt, got %v", dsUser.CreatedAt)
	}

	if !dsUser.UpdatedAt.IsZero() {
		t.Errorf("Expected zero UpdatedAt, got %v", dsUser.UpdatedAt)
	}

	// Convert back
	apiUser, err := datastore.UserFromUserDatastore(nil, dsUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserDatastore failed: %v", err)
	}

	// Zero time should convert back to nil timestamp (or zero timestamp)
	if apiUser.Birthday != nil && !apiUser.Birthday.AsTime().IsZero() {
		t.Errorf("Round-trip: Expected nil or zero Birthday, got %v", apiUser.Birthday.AsTime())
	}
}

// TestUserConversion_NilSource verifies that nil source returns nil.
func TestUserConversion_NilSource(t *testing.T) {
	dsUser, err := datastore.UserToUserDatastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserDatastore with nil src failed: %v", err)
	}
	if dsUser != nil {
		t.Errorf("Expected nil result for nil source, got %v", dsUser)
	}

	apiUser, err := datastore.UserFromUserDatastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("UserFromUserDatastore with nil src failed: %v", err)
	}
	if apiUser != nil {
		t.Errorf("Expected nil result for nil source, got %v", apiUser)
	}
}

// TestUserConversion_WithDestination verifies conversion into existing destination.
func TestUserConversion_WithDestination(t *testing.T) {
	src := &api.User{
		Id:    100,
		Name:  "Destination Test",
		Email: "dest@example.com",
	}

	existingDest := &datastore.UserDatastore{
		Id:   "old-id",
		Name: "Old Name",
	}

	dsUser, err := datastore.UserToUserDatastore(src, existingDest, nil)
	if err != nil {
		t.Fatalf("UserToUserDatastore failed: %v", err)
	}

	// Should return the same pointer
	if dsUser != existingDest {
		t.Error("Expected returned pointer to match destination pointer")
	}

	// Fields should be overwritten
	if dsUser.Id != "100" {
		t.Errorf("Id not overwritten: got %s, want '100'", dsUser.Id)
	}

	if dsUser.Name != "Destination Test" {
		t.Errorf("Name not overwritten: got %s, want 'Destination Test'", dsUser.Name)
	}
}

// TestUserConversion_WithDecorator verifies that decorator function is called.
func TestUserConversion_WithDecorator(t *testing.T) {
	src := &api.User{
		Id:    200,
		Name:  "Decorator Test",
		Email: "decorator@example.com",
	}

	decoratorCalled := false
	decorator := func(src *api.User, dest *datastore.UserDatastore) error {
		decoratorCalled = true
		// Modify the destination
		dest.Name = "Modified by Decorator"
		return nil
	}

	dsUser, err := datastore.UserToUserDatastore(src, nil, decorator)
	if err != nil {
		t.Fatalf("UserToUserDatastore failed: %v", err)
	}

	if !decoratorCalled {
		t.Error("Decorator was not called")
	}

	if dsUser.Name != "Modified by Decorator" {
		t.Errorf("Decorator modification not applied: got %s", dsUser.Name)
	}
}

// TestUserConversion_UserWithNamespace verifies UserWithNamespace variant.
func TestUserConversion_UserWithNamespace(t *testing.T) {
	birthday := time.Date(1985, 3, 20, 0, 0, 0, 0, time.UTC)

	src := &api.User{
		Id:           300,
		Name:         "Namespace User",
		Email:        "namespace@example.com",
		Age:          39,
		Birthday:     timestamppb.New(birthday),
		MemberNumber: "NS-001",
	}

	dsUser, err := datastore.UserToUserWithNamespace(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserWithNamespace failed: %v", err)
	}

	if dsUser.Id != "300" {
		t.Errorf("Id mismatch: got %s, want '300'", dsUser.Id)
	}

	if dsUser.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsUser.Name, src.Name)
	}

	// Round-trip
	apiUser, err := datastore.UserFromUserWithNamespace(nil, dsUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserWithNamespace failed: %v", err)
	}

	if apiUser.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiUser.Id, src.Id)
	}
}

// TestUserConversion_UserWithLargeText verifies UserWithLargeText variant.
func TestUserConversion_UserWithLargeText(t *testing.T) {
	src := &api.User{
		Id:           400,
		Name:         "Large Text User",
		Email:        "largetext@example.com",
		Age:          45,
		MemberNumber: "LT-001",
	}

	dsUser, err := datastore.UserToUserWithLargeText(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserWithLargeText failed: %v", err)
	}

	if dsUser.Id != "400" {
		t.Errorf("Id mismatch: got %s, want '400'", dsUser.Id)
	}

	// Round-trip
	apiUser, err := datastore.UserFromUserWithLargeText(nil, dsUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserWithLargeText failed: %v", err)
	}

	if apiUser.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiUser.Id, src.Id)
	}
}

// TestUserConversion_UserSimple verifies UserSimple variant.
func TestUserConversion_UserSimple(t *testing.T) {
	src := &api.User{
		Id:           500,
		Name:         "Simple User",
		Email:        "simple@example.com",
		Age:          30,
		MemberNumber: "SMP-001",
	}

	dsUser, err := datastore.UserToUserSimple(src, nil, nil)
	if err != nil {
		t.Fatalf("UserToUserSimple failed: %v", err)
	}

	if dsUser.Id != "500" {
		t.Errorf("Id mismatch: got %s, want '500'", dsUser.Id)
	}

	if dsUser.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsUser.Name, src.Name)
	}

	// Round-trip
	apiUser, err := datastore.UserFromUserSimple(nil, dsUser, nil)
	if err != nil {
		t.Fatalf("UserFromUserSimple failed: %v", err)
	}

	if apiUser.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiUser.Id, src.Id)
	}

	if apiUser.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiUser.Name, src.Name)
	}
}

// TestAuthorConversion verifies Author to AuthorDatastore conversion.
func TestAuthorConversion(t *testing.T) {
	src := &api.Author{
		Name:  "Jane Author",
		Email: "jane@authors.com",
	}

	dsAuthor, err := datastore.AuthorToAuthorDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("AuthorToAuthorDatastore failed: %v", err)
	}

	if dsAuthor.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsAuthor.Name, src.Name)
	}

	if dsAuthor.Email != src.Email {
		t.Errorf("Email mismatch: got %s, want %s", dsAuthor.Email, src.Email)
	}

	// Round-trip
	apiAuthor, err := datastore.AuthorFromAuthorDatastore(nil, dsAuthor, nil)
	if err != nil {
		t.Fatalf("AuthorFromAuthorDatastore failed: %v", err)
	}

	if apiAuthor.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiAuthor.Name, src.Name)
	}

	if apiAuthor.Email != src.Email {
		t.Errorf("Round-trip Email mismatch: got %s, want %s", apiAuthor.Email, src.Email)
	}
}

// TestAuthorConversion_Nil verifies nil handling for Author.
func TestAuthorConversion_Nil(t *testing.T) {
	dsAuthor, err := datastore.AuthorToAuthorDatastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("AuthorToAuthorDatastore with nil failed: %v", err)
	}
	if dsAuthor != nil {
		t.Errorf("Expected nil result for nil source")
	}

	apiAuthor, err := datastore.AuthorFromAuthorDatastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("AuthorFromAuthorDatastore with nil failed: %v", err)
	}
	if apiAuthor != nil {
		t.Errorf("Expected nil result for nil source")
	}
}

// TestProductConversion_BasicFields verifies Product conversion with basic fields.
func TestProductConversion_BasicFields(t *testing.T) {
	src := &api.Product{
		Id:   1001,
		Name: "Widget Pro",
	}

	dsProduct, err := datastore.ProductToProductDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductDatastore failed: %v", err)
	}

	if dsProduct.Id != "1001" {
		t.Errorf("Id mismatch: got %s, want '1001'", dsProduct.Id)
	}

	if dsProduct.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsProduct.Name, src.Name)
	}

	// Round-trip
	apiProduct, err := datastore.ProductFromProductDatastore(nil, dsProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductDatastore failed: %v", err)
	}

	if apiProduct.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiProduct.Id, src.Id)
	}

	if apiProduct.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiProduct.Name, src.Name)
	}
}

// TestProductConversion_RepeatedStrings verifies repeated string fields (Tags, Categories).
func TestProductConversion_RepeatedStrings(t *testing.T) {
	src := &api.Product{
		Id:         1002,
		Name:       "Tagged Product",
		Tags:       []string{"electronics", "sale", "featured"},
		Categories: []string{"tech", "gadgets"},
	}

	dsProduct, err := datastore.ProductToProductDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(dsProduct.Tags, src.Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", dsProduct.Tags, src.Tags)
	}

	if !reflect.DeepEqual(dsProduct.Categories, src.Categories) {
		t.Errorf("Categories mismatch: got %v, want %v", dsProduct.Categories, src.Categories)
	}

	// Round-trip
	apiProduct, err := datastore.ProductFromProductDatastore(nil, dsProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(apiProduct.Tags, src.Tags) {
		t.Errorf("Round-trip Tags mismatch: got %v, want %v", apiProduct.Tags, src.Tags)
	}

	if !reflect.DeepEqual(apiProduct.Categories, src.Categories) {
		t.Errorf("Round-trip Categories mismatch: got %v, want %v", apiProduct.Categories, src.Categories)
	}
}

// TestProductConversion_RepeatedInts verifies repeated int32 fields (Ratings).
func TestProductConversion_RepeatedInts(t *testing.T) {
	src := &api.Product{
		Id:      1003,
		Name:    "Rated Product",
		Ratings: []int32{5, 4, 5, 3, 4, 5},
	}

	dsProduct, err := datastore.ProductToProductDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(dsProduct.Ratings, src.Ratings) {
		t.Errorf("Ratings mismatch: got %v, want %v", dsProduct.Ratings, src.Ratings)
	}

	// Round-trip
	apiProduct, err := datastore.ProductFromProductDatastore(nil, dsProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(apiProduct.Ratings, src.Ratings) {
		t.Errorf("Round-trip Ratings mismatch: got %v, want %v", apiProduct.Ratings, src.Ratings)
	}
}

// TestProductConversion_MapField verifies map<string, string> field (Metadata).
func TestProductConversion_MapField(t *testing.T) {
	src := &api.Product{
		Id:   1004,
		Name: "Metadata Product",
		Metadata: map[string]string{
			"color":    "blue",
			"size":     "large",
			"material": "cotton",
		},
	}

	dsProduct, err := datastore.ProductToProductDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(dsProduct.Metadata, src.Metadata) {
		t.Errorf("Metadata mismatch: got %v, want %v", dsProduct.Metadata, src.Metadata)
	}

	// Round-trip
	apiProduct, err := datastore.ProductFromProductDatastore(nil, dsProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductDatastore failed: %v", err)
	}

	if !reflect.DeepEqual(apiProduct.Metadata, src.Metadata) {
		t.Errorf("Round-trip Metadata mismatch: got %v, want %v", apiProduct.Metadata, src.Metadata)
	}
}

// TestProductConversion_EmptyCollections verifies empty collections are handled.
func TestProductConversion_EmptyCollections(t *testing.T) {
	src := &api.Product{
		Id:         1005,
		Name:       "Empty Collections",
		Tags:       []string{},
		Categories: []string{},
		Metadata:   map[string]string{},
		Ratings:    []int32{},
	}

	dsProduct, err := datastore.ProductToProductDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductDatastore failed: %v", err)
	}

	// Empty slices should remain empty (not nil)
	if dsProduct.Tags == nil {
		t.Error("Tags should not be nil for empty slice")
	}
	if len(dsProduct.Tags) != 0 {
		t.Errorf("Tags should be empty, got %v", dsProduct.Tags)
	}

	// Round-trip
	apiProduct, err := datastore.ProductFromProductDatastore(nil, dsProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductDatastore failed: %v", err)
	}

	if apiProduct.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiProduct.Id, src.Id)
	}
}

// TestProductConversion_NilCollections verifies nil collections are handled.
func TestProductConversion_NilCollections(t *testing.T) {
	src := &api.Product{
		Id:         1006,
		Name:       "Nil Collections",
		Tags:       nil,
		Categories: nil,
		Metadata:   nil,
		Ratings:    nil,
	}

	dsProduct, err := datastore.ProductToProductDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductDatastore failed: %v", err)
	}

	// Nil slices should remain nil
	if dsProduct.Tags != nil {
		t.Errorf("Tags should be nil, got %v", dsProduct.Tags)
	}

	if dsProduct.Metadata != nil {
		t.Errorf("Metadata should be nil, got %v", dsProduct.Metadata)
	}

	// Round-trip
	apiProduct, err := datastore.ProductFromProductDatastore(nil, dsProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductDatastore failed: %v", err)
	}

	if apiProduct.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiProduct.Id, src.Id)
	}
}

// TestProductConversion_FullProduct verifies a fully populated Product.
func TestProductConversion_FullProduct(t *testing.T) {
	src := &api.Product{
		Id:         1007,
		Name:       "Full Product",
		Tags:       []string{"new", "featured", "bestseller"},
		Categories: []string{"electronics", "accessories"},
		Metadata: map[string]string{
			"brand":    "Acme",
			"warranty": "2 years",
			"origin":   "USA",
		},
		Ratings: []int32{5, 5, 4, 5, 4, 5, 3, 5},
	}

	dsProduct, err := datastore.ProductToProductDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("ProductToProductDatastore failed: %v", err)
	}

	// Verify all fields
	if dsProduct.Id != "1007" {
		t.Errorf("Id mismatch: got %s, want '1007'", dsProduct.Id)
	}

	if dsProduct.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsProduct.Name, src.Name)
	}

	if !reflect.DeepEqual(dsProduct.Tags, src.Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", dsProduct.Tags, src.Tags)
	}

	if !reflect.DeepEqual(dsProduct.Categories, src.Categories) {
		t.Errorf("Categories mismatch: got %v, want %v", dsProduct.Categories, src.Categories)
	}

	if !reflect.DeepEqual(dsProduct.Metadata, src.Metadata) {
		t.Errorf("Metadata mismatch: got %v, want %v", dsProduct.Metadata, src.Metadata)
	}

	if !reflect.DeepEqual(dsProduct.Ratings, src.Ratings) {
		t.Errorf("Ratings mismatch: got %v, want %v", dsProduct.Ratings, src.Ratings)
	}

	// Full round-trip
	apiProduct, err := datastore.ProductFromProductDatastore(nil, dsProduct, nil)
	if err != nil {
		t.Fatalf("ProductFromProductDatastore failed: %v", err)
	}

	if apiProduct.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiProduct.Id, src.Id)
	}

	if apiProduct.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiProduct.Name, src.Name)
	}

	if !reflect.DeepEqual(apiProduct.Tags, src.Tags) {
		t.Errorf("Round-trip Tags mismatch: got %v, want %v", apiProduct.Tags, src.Tags)
	}

	if !reflect.DeepEqual(apiProduct.Categories, src.Categories) {
		t.Errorf("Round-trip Categories mismatch: got %v, want %v", apiProduct.Categories, src.Categories)
	}

	if !reflect.DeepEqual(apiProduct.Metadata, src.Metadata) {
		t.Errorf("Round-trip Metadata mismatch: got %v, want %v", apiProduct.Metadata, src.Metadata)
	}

	if !reflect.DeepEqual(apiProduct.Ratings, src.Ratings) {
		t.Errorf("Round-trip Ratings mismatch: got %v, want %v", apiProduct.Ratings, src.Ratings)
	}
}

// TestLibraryConversion_BasicFields verifies Library conversion with basic fields.
func TestLibraryConversion_BasicFields(t *testing.T) {
	src := &api.Library{
		Id:   2001,
		Name: "Central Library",
	}

	dsLibrary, err := datastore.LibraryToLibraryDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("LibraryToLibraryDatastore failed: %v", err)
	}

	if dsLibrary.Id != "2001" {
		t.Errorf("Id mismatch: got %s, want '2001'", dsLibrary.Id)
	}

	if dsLibrary.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsLibrary.Name, src.Name)
	}

	// Round-trip
	apiLibrary, err := datastore.LibraryFromLibraryDatastore(nil, dsLibrary, nil)
	if err != nil {
		t.Fatalf("LibraryFromLibraryDatastore failed: %v", err)
	}

	if apiLibrary.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiLibrary.Id, src.Id)
	}

	if apiLibrary.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiLibrary.Name, src.Name)
	}
}

// TestLibraryConversion_RepeatedMessages verifies repeated message field (Contributors).
func TestLibraryConversion_RepeatedMessages(t *testing.T) {
	src := &api.Library{
		Id:   2002,
		Name: "Authors Library",
		Contributors: []*api.Author{
			{Name: "Alice Author", Email: "alice@authors.com"},
			{Name: "Bob Writer", Email: "bob@writers.org"},
			{Name: "Carol Editor", Email: "carol@editors.net"},
		},
	}

	dsLibrary, err := datastore.LibraryToLibraryDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("LibraryToLibraryDatastore failed: %v", err)
	}

	// Verify contributors count
	if len(dsLibrary.Contributors) != 3 {
		t.Errorf("Contributors count mismatch: got %d, want 3", len(dsLibrary.Contributors))
	}

	// Verify each contributor
	for i, contributor := range src.Contributors {
		if dsLibrary.Contributors[i].Name != contributor.Name {
			t.Errorf("Contributor[%d] Name mismatch: got %s, want %s",
				i, dsLibrary.Contributors[i].Name, contributor.Name)
		}
		if dsLibrary.Contributors[i].Email != contributor.Email {
			t.Errorf("Contributor[%d] Email mismatch: got %s, want %s",
				i, dsLibrary.Contributors[i].Email, contributor.Email)
		}
	}

	// Round-trip
	apiLibrary, err := datastore.LibraryFromLibraryDatastore(nil, dsLibrary, nil)
	if err != nil {
		t.Fatalf("LibraryFromLibraryDatastore failed: %v", err)
	}

	if len(apiLibrary.Contributors) != 3 {
		t.Errorf("Round-trip Contributors count mismatch: got %d, want 3", len(apiLibrary.Contributors))
	}

	for i, contributor := range src.Contributors {
		if apiLibrary.Contributors[i].Name != contributor.Name {
			t.Errorf("Round-trip Contributor[%d] Name mismatch: got %s, want %s",
				i, apiLibrary.Contributors[i].Name, contributor.Name)
		}
		if apiLibrary.Contributors[i].Email != contributor.Email {
			t.Errorf("Round-trip Contributor[%d] Email mismatch: got %s, want %s",
				i, apiLibrary.Contributors[i].Email, contributor.Email)
		}
	}
}

// TestLibraryConversion_EmptyContributors verifies empty Contributors slice.
func TestLibraryConversion_EmptyContributors(t *testing.T) {
	src := &api.Library{
		Id:           2003,
		Name:         "Empty Contributors Library",
		Contributors: []*api.Author{},
	}

	dsLibrary, err := datastore.LibraryToLibraryDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("LibraryToLibraryDatastore failed: %v", err)
	}

	if dsLibrary.Contributors == nil {
		t.Error("Contributors should not be nil for empty slice")
	}

	if len(dsLibrary.Contributors) != 0 {
		t.Errorf("Contributors should be empty, got %v", dsLibrary.Contributors)
	}
}

// TestLibraryConversion_NilContributors verifies nil Contributors slice.
func TestLibraryConversion_NilContributors(t *testing.T) {
	src := &api.Library{
		Id:           2004,
		Name:         "Nil Contributors Library",
		Contributors: nil,
	}

	dsLibrary, err := datastore.LibraryToLibraryDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("LibraryToLibraryDatastore failed: %v", err)
	}

	if dsLibrary.Contributors != nil {
		t.Errorf("Contributors should be nil, got %v", dsLibrary.Contributors)
	}

	// Round-trip
	apiLibrary, err := datastore.LibraryFromLibraryDatastore(nil, dsLibrary, nil)
	if err != nil {
		t.Fatalf("LibraryFromLibraryDatastore failed: %v", err)
	}

	if apiLibrary.Contributors != nil {
		t.Errorf("Round-trip Contributors should be nil, got %v", apiLibrary.Contributors)
	}
}

// TestOrganizationConversion_BasicFields verifies Organization conversion with basic fields.
func TestOrganizationConversion_BasicFields(t *testing.T) {
	src := &api.Organization{
		Id:   3001,
		Name: "Acme Corp",
	}

	dsOrg, err := datastore.OrganizationToOrganizationDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationDatastore failed: %v", err)
	}

	if dsOrg.Id != "3001" {
		t.Errorf("Id mismatch: got %s, want '3001'", dsOrg.Id)
	}

	if dsOrg.Name != src.Name {
		t.Errorf("Name mismatch: got %s, want %s", dsOrg.Name, src.Name)
	}

	// Round-trip
	apiOrg, err := datastore.OrganizationFromOrganizationDatastore(nil, dsOrg, nil)
	if err != nil {
		t.Fatalf("OrganizationFromOrganizationDatastore failed: %v", err)
	}

	if apiOrg.Id != src.Id {
		t.Errorf("Round-trip Id mismatch: got %d, want %d", apiOrg.Id, src.Id)
	}

	if apiOrg.Name != src.Name {
		t.Errorf("Round-trip Name mismatch: got %s, want %s", apiOrg.Name, src.Name)
	}
}

// TestOrganizationConversion_MapOfMessages verifies map<string, Author> field (Departments).
func TestOrganizationConversion_MapOfMessages(t *testing.T) {
	src := &api.Organization{
		Id:   3002,
		Name: "Departments Corp",
		Departments: map[string]*api.Author{
			"engineering": {Name: "Alice Engineer", Email: "alice@eng.com"},
			"sales":       {Name: "Bob Sales", Email: "bob@sales.com"},
			"marketing":   {Name: "Carol Marketing", Email: "carol@marketing.com"},
		},
	}

	dsOrg, err := datastore.OrganizationToOrganizationDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationDatastore failed: %v", err)
	}

	// Verify departments count
	if len(dsOrg.Departments) != 3 {
		t.Errorf("Departments count mismatch: got %d, want 3", len(dsOrg.Departments))
	}

	// Verify each department
	for key, author := range src.Departments {
		dsDept, ok := dsOrg.Departments[key]
		if !ok {
			t.Errorf("Missing department key: %s", key)
			continue
		}
		if dsDept.Name != author.Name {
			t.Errorf("Department[%s] Name mismatch: got %s, want %s", key, dsDept.Name, author.Name)
		}
		if dsDept.Email != author.Email {
			t.Errorf("Department[%s] Email mismatch: got %s, want %s", key, dsDept.Email, author.Email)
		}
	}

	// Round-trip
	apiOrg, err := datastore.OrganizationFromOrganizationDatastore(nil, dsOrg, nil)
	if err != nil {
		t.Fatalf("OrganizationFromOrganizationDatastore failed: %v", err)
	}

	if len(apiOrg.Departments) != 3 {
		t.Errorf("Round-trip Departments count mismatch: got %d, want 3", len(apiOrg.Departments))
	}

	for key, author := range src.Departments {
		apiDept, ok := apiOrg.Departments[key]
		if !ok {
			t.Errorf("Round-trip missing department key: %s", key)
			continue
		}
		if apiDept.Name != author.Name {
			t.Errorf("Round-trip Department[%s] Name mismatch: got %s, want %s",
				key, apiDept.Name, author.Name)
		}
		if apiDept.Email != author.Email {
			t.Errorf("Round-trip Department[%s] Email mismatch: got %s, want %s",
				key, apiDept.Email, author.Email)
		}
	}
}

// TestOrganizationConversion_EmptyDepartments verifies empty Departments map.
func TestOrganizationConversion_EmptyDepartments(t *testing.T) {
	src := &api.Organization{
		Id:          3003,
		Name:        "Empty Departments Corp",
		Departments: map[string]*api.Author{},
	}

	dsOrg, err := datastore.OrganizationToOrganizationDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationDatastore failed: %v", err)
	}

	if dsOrg.Departments == nil {
		t.Error("Departments should not be nil for empty map")
	}

	if len(dsOrg.Departments) != 0 {
		t.Errorf("Departments should be empty, got %v", dsOrg.Departments)
	}
}

// TestOrganizationConversion_NilDepartments verifies nil Departments map.
func TestOrganizationConversion_NilDepartments(t *testing.T) {
	src := &api.Organization{
		Id:          3004,
		Name:        "Nil Departments Corp",
		Departments: nil,
	}

	dsOrg, err := datastore.OrganizationToOrganizationDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationDatastore failed: %v", err)
	}

	if dsOrg.Departments != nil {
		t.Errorf("Departments should be nil, got %v", dsOrg.Departments)
	}

	// Round-trip
	apiOrg, err := datastore.OrganizationFromOrganizationDatastore(nil, dsOrg, nil)
	if err != nil {
		t.Fatalf("OrganizationFromOrganizationDatastore failed: %v", err)
	}

	if apiOrg.Departments != nil {
		t.Errorf("Round-trip Departments should be nil, got %v", apiOrg.Departments)
	}
}

// TestOrganizationConversion_SingleDepartment verifies map with single entry.
func TestOrganizationConversion_SingleDepartment(t *testing.T) {
	src := &api.Organization{
		Id:   3005,
		Name: "Single Department Corp",
		Departments: map[string]*api.Author{
			"ceo": {Name: "Chief Executive", Email: "ceo@corp.com"},
		},
	}

	dsOrg, err := datastore.OrganizationToOrganizationDatastore(src, nil, nil)
	if err != nil {
		t.Fatalf("OrganizationToOrganizationDatastore failed: %v", err)
	}

	if len(dsOrg.Departments) != 1 {
		t.Errorf("Departments count mismatch: got %d, want 1", len(dsOrg.Departments))
	}

	ceo, ok := dsOrg.Departments["ceo"]
	if !ok {
		t.Error("Missing 'ceo' department")
	} else {
		if ceo.Name != "Chief Executive" {
			t.Errorf("CEO Name mismatch: got %s", ceo.Name)
		}
	}

	// Round-trip
	apiOrg, err := datastore.OrganizationFromOrganizationDatastore(nil, dsOrg, nil)
	if err != nil {
		t.Fatalf("OrganizationFromOrganizationDatastore failed: %v", err)
	}

	if len(apiOrg.Departments) != 1 {
		t.Errorf("Round-trip Departments count mismatch: got %d, want 1", len(apiOrg.Departments))
	}
}
