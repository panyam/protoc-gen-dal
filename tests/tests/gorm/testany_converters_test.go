package gorm

import (
	"testing"
	"time"

	"github.com/panyam/protoc-gen-dal/tests/gen/go/api"
	"github.com/panyam/protoc-gen-dal/tests/gen/gorm"
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

	// Convert to GORM
	gormRecord, err := gorm.TestRecord1ToTestRecord1GORM(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1GORM failed: %v", err)
	}

	// Verify TimeField conversion
	if !gormRecord.TimeField.Equal(timeField) {
		t.Errorf("TimeField mismatch: got %v, want %v", gormRecord.TimeField, timeField)
	}

	// Verify ExtraData conversion to []byte
	if gormRecord.ExtraData == nil {
		t.Fatal("ExtraData is nil after conversion")
	}

	if len(gormRecord.ExtraData) == 0 {
		t.Error("ExtraData is empty after conversion")
	}

	// Verify AnEnum conversion
	if gormRecord.AnEnum != src.AnEnum {
		t.Errorf("AnEnum mismatch: got %v, want %v", gormRecord.AnEnum, src.AnEnum)
	}

	// Convert back to API
	apiRecord, err := gorm.TestRecord1FromTestRecord1GORM(nil, gormRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1GORM failed: %v", err)
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

			// Convert to GORM
			gormRecord, err := gorm.TestRecord1ToTestRecord1GORM(src, nil, nil)
			if err != nil {
				t.Fatalf("TestRecord1ToTestRecord1GORM failed: %v", err)
			}

			if gormRecord.AnEnum != tc.enumValue {
				t.Errorf("AnEnum mismatch: got %v, want %v", gormRecord.AnEnum, tc.enumValue)
			}

			// Convert back
			apiRecord, err := gorm.TestRecord1FromTestRecord1GORM(nil, gormRecord, nil)
			if err != nil {
				t.Fatalf("TestRecord1FromTestRecord1GORM failed: %v", err)
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
		name string
		pack func() (*anypb.Any, error)
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
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Pack the message into Any
			anyData, err := tc.pack()
			if err != nil {
				t.Fatalf("Failed to create Any: %v", err)
			}

			src := &api.TestRecord1{
				ExtraData: anyData,
			}

			// Convert to GORM
			gormRecord, err := gorm.TestRecord1ToTestRecord1GORM(src, nil, nil)
			if err != nil {
				t.Fatalf("TestRecord1ToTestRecord1GORM failed: %v", err)
			}

			// Verify ExtraData is not nil
			if gormRecord.ExtraData == nil {
				t.Fatal("ExtraData is nil after conversion")
			}

			// Convert back
			apiRecord, err := gorm.TestRecord1FromTestRecord1GORM(nil, gormRecord, nil)
			if err != nil {
				t.Fatalf("TestRecord1FromTestRecord1GORM failed: %v", err)
			}

			// Verify ExtraData is not nil
			if apiRecord.ExtraData == nil {
				t.Fatal("Round-trip: ExtraData is nil")
			}

			// Verify using the test case's verify function
			tc.verify(t, apiRecord.ExtraData)
		})
	}
}

// TestRecord1Conversion_NilTimeField verifies nil timestamp handling.
func TestRecord1Conversion_NilTimeField(t *testing.T) {
	src := &api.TestRecord1{
		TimeField: nil,
		ExtraData: nil,
		AnEnum:    api.SampleEnum_SAMPLE_ENUM_A,
	}

	gormRecord, err := gorm.TestRecord1ToTestRecord1GORM(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1GORM failed: %v", err)
	}

	// Nil timestamp should result in zero time
	if !gormRecord.TimeField.IsZero() {
		t.Errorf("Expected zero TimeField, got %v", gormRecord.TimeField)
	}

	// Convert back
	apiRecord, err := gorm.TestRecord1FromTestRecord1GORM(nil, gormRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1GORM failed: %v", err)
	}

	// Verify timestamp is nil or zero
	if apiRecord.TimeField != nil && !apiRecord.TimeField.AsTime().IsZero() {
		t.Errorf("Round-trip: Expected nil or zero TimeField, got %v", apiRecord.TimeField.AsTime())
	}
}

// TestRecord1Conversion_NilAnyField verifies nil Any field handling.
func TestRecord1Conversion_NilAnyField(t *testing.T) {
	src := &api.TestRecord1{
		TimeField: timestamppb.New(time.Now()),
		ExtraData: nil, // nil Any field
		AnEnum:    api.SampleEnum_SAMPLE_ENUM_B,
	}

	gormRecord, err := gorm.TestRecord1ToTestRecord1GORM(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1GORM failed: %v", err)
	}

	// Nil Any should result in nil []byte
	if gormRecord.ExtraData != nil {
		t.Errorf("Expected nil ExtraData, got %v", gormRecord.ExtraData)
	}

	// Convert back
	apiRecord, err := gorm.TestRecord1FromTestRecord1GORM(nil, gormRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1GORM failed: %v", err)
	}

	// Verify ExtraData is nil
	if apiRecord.ExtraData != nil {
		t.Errorf("Round-trip: Expected nil ExtraData, got %v", apiRecord.ExtraData)
	}
}

// TestRecord1Conversion_AllFieldsNil verifies all nil fields are handled correctly.
func TestRecord1Conversion_AllFieldsNil(t *testing.T) {
	src := &api.TestRecord1{
		TimeField: nil,
		ExtraData: nil,
		AnEnum:    api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED,
	}

	gormRecord, err := gorm.TestRecord1ToTestRecord1GORM(src, nil, nil)
	if err != nil {
		t.Fatalf("TestRecord1ToTestRecord1GORM failed: %v", err)
	}

	// Verify all fields have expected values
	if !gormRecord.TimeField.IsZero() {
		t.Errorf("Expected zero TimeField, got %v", gormRecord.TimeField)
	}

	if gormRecord.ExtraData != nil {
		t.Errorf("Expected nil ExtraData, got %v", gormRecord.ExtraData)
	}

	if gormRecord.AnEnum != api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED {
		t.Errorf("AnEnum mismatch: got %v, want %v", gormRecord.AnEnum, api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED)
	}

	// Convert back
	apiRecord, err := gorm.TestRecord1FromTestRecord1GORM(nil, gormRecord, nil)
	if err != nil {
		t.Fatalf("TestRecord1FromTestRecord1GORM failed: %v", err)
	}

	// Verify round-trip
	if apiRecord.ExtraData != nil {
		t.Errorf("Round-trip: Expected nil ExtraData, got %v", apiRecord.ExtraData)
	}

	if apiRecord.AnEnum != api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED {
		t.Errorf("Round-trip AnEnum mismatch: got %v, want %v", apiRecord.AnEnum, api.SampleEnum_SAMPLE_ENUM_UNSPECIFIED)
	}
}
