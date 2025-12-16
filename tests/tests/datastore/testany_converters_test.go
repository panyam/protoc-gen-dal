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

	"github.com/panyam/protoc-gen-dal/tests/gen/datastore"
	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestRecord1Conversion_AllFields verifies that google.protobuf.Any, timestamps,
// and enum fields are correctly converted.
func TestRecord1Conversion_AllFields(t *testing.T) {
	// Create a timestamp
	timeField := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	// Create a message to pack into Any field (using Author as example)
	author := &api.Author{
		Name:  "Alice Smith",
		Email: "alice@example.com",
	}

	// Pack the author into google.protobuf.Any
	anyData, err := anypb.New(author)
	if err != nil {
		t.Fatalf("Failed to create Any: %v", err)
	}

	src := &api.TestRecord1{
		TimeField: timestamppb.New(timeField),
		ExtraData: anyData,
		AnEnum:    api.SampleEnum_SAMPLE_ENUM_C,
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	// Verify TimeField conversion
	if !dsRecord.TimeField.Equal(timeField) {
		t.Errorf("TimeField mismatch: got %v, want %v", dsRecord.TimeField, timeField)
	}

	// Verify ExtraData conversion to []byte
	if dsRecord.ExtraData == nil {
		t.Fatal("ExtraData is nil after conversion")
	}

	if len(dsRecord.ExtraData) == 0 {
		t.Error("ExtraData is empty after conversion")
	}

	// Verify AnEnum conversion
	if dsRecord.AnEnum != src.AnEnum {
		t.Errorf("AnEnum mismatch: got %v, want %v", dsRecord.AnEnum, src.AnEnum)
	}

	// Convert back to API
	apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
	}

	// Verify round-trip TimeField
	if !apiRecord.TimeField.AsTime().Equal(timeField) {
		t.Errorf("Round-trip TimeField mismatch: got %v, want %v", apiRecord.TimeField.AsTime(), timeField)
	}

	// Verify round-trip ExtraData
	if apiRecord.ExtraData == nil {
		t.Fatal("Round-trip: ExtraData is nil")
	}

	// Verify we can unpack the Any field back to the original message
	var unpackedAuthor api.Author
	if err := apiRecord.ExtraData.UnmarshalTo(&unpackedAuthor); err != nil {
		t.Fatalf("Failed to unmarshal ExtraData: %v", err)
	}

	if unpackedAuthor.Name != author.Name {
		t.Errorf("Round-trip Author.Name mismatch: got %s, want %s", unpackedAuthor.Name, author.Name)
	}

	if unpackedAuthor.Email != author.Email {
		t.Errorf("Round-trip Author.Email mismatch: got %s, want %s", unpackedAuthor.Email, author.Email)
	}

	// Verify round-trip AnEnum
	if apiRecord.AnEnum != src.AnEnum {
		t.Errorf("Round-trip AnEnum mismatch: got %v, want %v", apiRecord.AnEnum, src.AnEnum)
	}
}

// TestRecord1Conversion_DifferentEnumValues verifies all enum values are handled.
func TestRecord1Conversion_DifferentEnumValues(t *testing.T) {
	testCases := []struct {
		name      string
		enumValue api.SampleEnum
	}{
		{"Unspecified", api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED},
		{"EnumA", api.SampleEnum_SAMPLE_ENUM_A},
		{"EnumB", api.SampleEnum_SAMPLE_ENUM_B},
		{"EnumC", api.SampleEnum_SAMPLE_ENUM_C},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := &api.TestRecord1{
				AnEnum: tc.enumValue,
			}

			// Convert to Datastore
			dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
			if err != nil {
				t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
			}

			if dsRecord.AnEnum != tc.enumValue {
				t.Errorf("AnEnum mismatch: got %v, want %v", dsRecord.AnEnum, tc.enumValue)
			}

			// Convert back
			apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
			if err != nil {
				t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
			}

			if apiRecord.AnEnum != tc.enumValue {
				t.Errorf("Round-trip AnEnum mismatch: got %v, want %v", apiRecord.AnEnum, tc.enumValue)
			}
		})
	}
}

// TestRecord1Conversion_AnyWithDifferentMessages verifies Any field can hold different message types.
func TestRecord1Conversion_AnyWithDifferentMessages(t *testing.T) {
	testCases := []struct {
		name   string
		pack   func() (*anypb.Any, error)
		verify func(t *testing.T, anyData *anypb.Any)
	}{
		{
			name: "Author",
			pack: func() (*anypb.Any, error) {
				return anypb.New(&api.Author{
					Name:  "Bob Jones",
					Email: "bob@example.com",
				})
			},
			verify: func(t *testing.T, anyData *anypb.Any) {
				var author api.Author
				if err := anyData.UnmarshalTo(&author); err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}
				if author.Name != "Bob Jones" || author.Email != "bob@example.com" {
					t.Errorf("Author fields mismatch: got Name=%s, Email=%s", author.Name, author.Email)
				}
			},
		},
		{
			name: "User",
			pack: func() (*anypb.Any, error) {
				return anypb.New(&api.User{
					Id:    100,
					Name:  "Test User",
					Email: "test@example.com",
					Age:   30,
				})
			},
			verify: func(t *testing.T, anyData *anypb.Any) {
				var user api.User
				if err := anyData.UnmarshalTo(&user); err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}
				if user.Id != 100 || user.Name != "Test User" {
					t.Errorf("User fields mismatch: got Id=%d, Name=%s", user.Id, user.Name)
				}
			},
		},
		{
			name: "Product",
			pack: func() (*anypb.Any, error) {
				return anypb.New(&api.Product{
					Id:   50,
					Name: "Test Product",
					Tags: []string{"tag1", "tag2"},
				})
			},
			verify: func(t *testing.T, anyData *anypb.Any) {
				var product api.Product
				if err := anyData.UnmarshalTo(&product); err != nil {
					t.Fatalf("Failed to unmarshal: %v", err)
				}
				if product.Id != 50 || product.Name != "Test Product" {
					t.Errorf("Product fields mismatch: got Id=%d, Name=%s", product.Id, product.Name)
				}
				if len(product.Tags) != 2 {
					t.Errorf("Product Tags mismatch: got %v", product.Tags)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			anyData, err := tc.pack()
			if err != nil {
				t.Fatalf("Failed to pack: %v", err)
			}

			src := &api.TestRecord1{
				ExtraData: anyData,
			}

			// Convert to Datastore
			dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
			if err != nil {
				t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
			}

			if dsRecord.ExtraData == nil {
				t.Fatal("ExtraData is nil")
			}

			// Convert back
			apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
			if err != nil {
				t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
			}

			if apiRecord.ExtraData == nil {
				t.Fatal("Round-trip: ExtraData is nil")
			}

			// Verify the unpacked message
			tc.verify(t, apiRecord.ExtraData)
		})
	}
}

// TestRecord1Conversion_NilAny verifies nil Any field is handled correctly.
func TestRecord1Conversion_NilAny(t *testing.T) {
	src := &api.TestRecord1{
		ExtraData: nil,
		AnEnum:    api.SampleEnum_SAMPLE_ENUM_A,
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if dsRecord.ExtraData != nil {
		t.Errorf("ExtraData should be nil, got %v", dsRecord.ExtraData)
	}

	// Convert back
	apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
	}

	if apiRecord.ExtraData != nil {
		t.Errorf("Round-trip ExtraData should be nil, got %v", apiRecord.ExtraData)
	}
}

// TestRecord1Conversion_NilTimestamp verifies nil timestamp field is handled correctly.
func TestRecord1Conversion_NilTimestamp(t *testing.T) {
	src := &api.TestRecord1{
		TimeField: nil,
		AnEnum:    api.SampleEnum_SAMPLE_ENUM_B,
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	// Nil timestamp should result in zero time
	if !dsRecord.TimeField.IsZero() {
		t.Errorf("TimeField should be zero, got %v", dsRecord.TimeField)
	}

	// Convert back
	apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
	}

	// Zero time converts back to a zero timestamp
	if apiRecord.TimeField != nil && !apiRecord.TimeField.AsTime().IsZero() {
		t.Errorf("Round-trip TimeField should be nil or zero, got %v", apiRecord.TimeField.AsTime())
	}
}

// TestRecord1Conversion_NilSource verifies nil source returns nil.
func TestRecord1Conversion_NilSource(t *testing.T) {
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore with nil failed: %v", err)
	}
	if dsRecord != nil {
		t.Errorf("Expected nil result for nil source, got %v", dsRecord)
	}

	apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1Datastore with nil failed: %v", err)
	}
	if apiRecord != nil {
		t.Errorf("Expected nil result for nil source, got %v", apiRecord)
	}
}

// TestRecord1Conversion_ListOfEnums verifies repeated enum field handling.
func TestRecord1Conversion_ListOfEnums(t *testing.T) {
	src := &api.TestRecord1{
		ListOfEnums: []api.SampleEnum{
			api.SampleEnum_SAMPLE_ENUM_A,
			api.SampleEnum_SAMPLE_ENUM_B,
			api.SampleEnum_SAMPLE_ENUM_C,
			api.SampleEnum_SAMPLE_ENUM_A,
		},
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if !reflect.DeepEqual(dsRecord.ListOfEnums, src.ListOfEnums) {
		t.Errorf("ListOfEnums mismatch: got %v, want %v", dsRecord.ListOfEnums, src.ListOfEnums)
	}

	// Convert back
	apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
	}

	if !reflect.DeepEqual(apiRecord.ListOfEnums, src.ListOfEnums) {
		t.Errorf("Round-trip ListOfEnums mismatch: got %v, want %v", apiRecord.ListOfEnums, src.ListOfEnums)
	}
}

// TestRecord1Conversion_EmptyListOfEnums verifies empty repeated enum field handling.
func TestRecord1Conversion_EmptyListOfEnums(t *testing.T) {
	src := &api.TestRecord1{
		ListOfEnums: []api.SampleEnum{},
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if dsRecord.ListOfEnums == nil {
		t.Error("ListOfEnums should not be nil for empty slice")
	}

	if len(dsRecord.ListOfEnums) != 0 {
		t.Errorf("ListOfEnums should be empty, got %v", dsRecord.ListOfEnums)
	}
}

// TestRecord1Conversion_NilListOfEnums verifies nil repeated enum field handling.
func TestRecord1Conversion_NilListOfEnums(t *testing.T) {
	src := &api.TestRecord1{
		ListOfEnums: nil,
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if dsRecord.ListOfEnums != nil {
		t.Errorf("ListOfEnums should be nil, got %v", dsRecord.ListOfEnums)
	}
}

// TestRecord1Conversion_MapStringToEnum verifies map<string, SampleEnum> field handling.
func TestRecord1Conversion_MapStringToEnum(t *testing.T) {
	src := &api.TestRecord1{
		MapStringToEnum: map[string]api.SampleEnum{
			"first":  api.SampleEnum_SAMPLE_ENUM_A,
			"second": api.SampleEnum_SAMPLE_ENUM_B,
			"third":  api.SampleEnum_SAMPLE_ENUM_C,
		},
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if !reflect.DeepEqual(dsRecord.MapStringToEnum, src.MapStringToEnum) {
		t.Errorf("MapStringToEnum mismatch: got %v, want %v", dsRecord.MapStringToEnum, src.MapStringToEnum)
	}

	// Convert back
	apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
	}

	if !reflect.DeepEqual(apiRecord.MapStringToEnum, src.MapStringToEnum) {
		t.Errorf("Round-trip MapStringToEnum mismatch: got %v, want %v", apiRecord.MapStringToEnum, src.MapStringToEnum)
	}
}

// TestRecord1Conversion_EmptyMapStringToEnum verifies empty map field handling.
func TestRecord1Conversion_EmptyMapStringToEnum(t *testing.T) {
	src := &api.TestRecord1{
		MapStringToEnum: map[string]api.SampleEnum{},
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if dsRecord.MapStringToEnum == nil {
		t.Error("MapStringToEnum should not be nil for empty map")
	}

	if len(dsRecord.MapStringToEnum) != 0 {
		t.Errorf("MapStringToEnum should be empty, got %v", dsRecord.MapStringToEnum)
	}
}

// TestRecord1Conversion_NilMapStringToEnum verifies nil map field handling.
func TestRecord1Conversion_NilMapStringToEnum(t *testing.T) {
	src := &api.TestRecord1{
		MapStringToEnum: nil,
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if dsRecord.MapStringToEnum != nil {
		t.Errorf("MapStringToEnum should be nil, got %v", dsRecord.MapStringToEnum)
	}
}

// TestRecord1Conversion_FullRecord verifies fully populated TestRecord1.
func TestRecord1Conversion_FullRecord(t *testing.T) {
	timeField := time.Date(2024, 12, 25, 12, 0, 0, 0, time.UTC)

	author := &api.Author{
		Name:  "Full Author",
		Email: "full@example.com",
	}
	anyData, err := anypb.New(author)
	if err != nil {
		t.Fatalf("Failed to create Any: %v", err)
	}

	src := &api.TestRecord1{
		TimeField: timestamppb.New(timeField),
		ExtraData: anyData,
		AnEnum:    api.SampleEnum_SAMPLE_ENUM_C,
		ListOfEnums: []api.SampleEnum{
			api.SampleEnum_SAMPLE_ENUM_A,
			api.SampleEnum_SAMPLE_ENUM_B,
		},
		MapStringToEnum: map[string]api.SampleEnum{
			"key1": api.SampleEnum_SAMPLE_ENUM_C,
			"key2": api.SampleEnum_SAMPLE_ENUM_A,
		},
	}

	// Convert to Datastore
	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	// Verify all fields
	if !dsRecord.TimeField.Equal(timeField) {
		t.Errorf("TimeField mismatch: got %v, want %v", dsRecord.TimeField, timeField)
	}

	if dsRecord.ExtraData == nil || len(dsRecord.ExtraData) == 0 {
		t.Error("ExtraData should not be nil or empty")
	}

	if dsRecord.AnEnum != api.SampleEnum_SAMPLE_ENUM_C {
		t.Errorf("AnEnum mismatch: got %v, want %v", dsRecord.AnEnum, api.SampleEnum_SAMPLE_ENUM_C)
	}

	if len(dsRecord.ListOfEnums) != 2 {
		t.Errorf("ListOfEnums length mismatch: got %d, want 2", len(dsRecord.ListOfEnums))
	}

	if len(dsRecord.MapStringToEnum) != 2 {
		t.Errorf("MapStringToEnum length mismatch: got %d, want 2", len(dsRecord.MapStringToEnum))
	}

	// Full round-trip
	apiRecord, err := datastore.TestRecord1FromTestRecord1Datastore(nil, dsRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1Datastore failed: %v", err)
	}

	// Verify round-trip
	if !apiRecord.TimeField.AsTime().Equal(timeField) {
		t.Errorf("Round-trip TimeField mismatch: got %v, want %v", apiRecord.TimeField.AsTime(), timeField)
	}

	if apiRecord.ExtraData == nil {
		t.Fatal("Round-trip ExtraData is nil")
	}

	var unpackedAuthor api.Author
	if err := apiRecord.ExtraData.UnmarshalTo(&unpackedAuthor); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if unpackedAuthor.Name != author.Name {
		t.Errorf("Round-trip Author.Name mismatch: got %s, want %s", unpackedAuthor.Name, author.Name)
	}

	if apiRecord.AnEnum != src.AnEnum {
		t.Errorf("Round-trip AnEnum mismatch: got %v, want %v", apiRecord.AnEnum, src.AnEnum)
	}

	if !reflect.DeepEqual(apiRecord.ListOfEnums, src.ListOfEnums) {
		t.Errorf("Round-trip ListOfEnums mismatch: got %v, want %v", apiRecord.ListOfEnums, src.ListOfEnums)
	}

	if !reflect.DeepEqual(apiRecord.MapStringToEnum, src.MapStringToEnum) {
		t.Errorf("Round-trip MapStringToEnum mismatch: got %v, want %v", apiRecord.MapStringToEnum, src.MapStringToEnum)
	}
}

// TestRecord1Conversion_WithDecorator verifies decorator function is called.
func TestRecord1Conversion_WithDecorator(t *testing.T) {
	src := &api.TestRecord1{
		AnEnum: api.SampleEnum_SAMPLE_ENUM_A,
	}

	decoratorCalled := false
	decorator := func(src *api.TestRecord1, dest *datastore.TestRecord1Datastore) error {
		decoratorCalled = true
		// Change the enum value in decorator
		dest.AnEnum = api.SampleEnum_SAMPLE_ENUM_C
		return nil
	}

	dsRecord, err := datastore.TestRecord1ToTestRecord1Datastore(src, nil, decorator)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1Datastore failed: %v", err)
	}

	if !decoratorCalled {
		t.Error("Decorator was not called")
	}

	// Verify decorator modification was applied
	if dsRecord.AnEnum != api.SampleEnum_SAMPLE_ENUM_C {
		t.Errorf("AnEnum not modified by decorator: got %v, want %v", dsRecord.AnEnum, api.SampleEnum_SAMPLE_ENUM_C)
	}
}
