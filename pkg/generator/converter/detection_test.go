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

package converter

import (
	"testing"
)

func TestIsNumericConversion(t *testing.T) {
	tests := []struct {
		name       string
		sourceKind string
		targetKind string
		want       bool
	}{
		{
			name:       "int32 to int64",
			sourceKind: "int32",
			targetKind: "int64",
			want:       true,
		},
		{
			name:       "uint32 to uint64",
			sourceKind: "uint32",
			targetKind: "uint64",
			want:       true,
		},
		{
			name:       "float to double",
			sourceKind: "float",
			targetKind: "double",
			want:       true,
		},
		{
			name:       "string to int32",
			sourceKind: "string",
			targetKind: "int32",
			want:       false,
		},
		{
			name:       "message to int32",
			sourceKind: "message",
			targetKind: "int32",
			want:       false,
		},
		{
			name:       "bool to int32",
			sourceKind: "bool",
			targetKind: "int32",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumericConversion(tt.sourceKind, tt.targetKind); got != tt.want {
				t.Errorf("IsNumericConversion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSameScalarType(t *testing.T) {
	tests := []struct {
		name       string
		sourceKind string
		targetKind string
		want       bool
	}{
		{
			name:       "same string type",
			sourceKind: "string",
			targetKind: "string",
			want:       true,
		},
		{
			name:       "same int32 type",
			sourceKind: "int32",
			targetKind: "int32",
			want:       true,
		},
		{
			name:       "different scalar types",
			sourceKind: "int32",
			targetKind: "int64",
			want:       false,
		},
		{
			name:       "same message type",
			sourceKind: "message",
			targetKind: "message",
			want:       false, // messages don't count as "same scalar"
		},
		{
			name:       "different types",
			sourceKind: "string",
			targetKind: "int32",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSameScalarType(tt.sourceKind, tt.targetKind); got != tt.want {
				t.Errorf("IsSameScalarType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildNumericCast(t *testing.T) {
	tests := []struct {
		name            string
		sourceFieldName string
		targetKind      string
		want            string
	}{
		{
			name:            "cast to int64",
			sourceFieldName: "Count",
			targetKind:      "int64",
			want:            "int64(src.Count)",
		},
		{
			name:            "cast to uint32",
			sourceFieldName: "ID",
			targetKind:      "uint32",
			want:            "uint32(src.ID)",
		},
		{
			name:            "cast to float64",
			sourceFieldName: "Score",
			targetKind:      "double",
			want:            "float64(src.Score)",
		},
		{
			name:            "cast to float32",
			sourceFieldName: "Rating",
			targetKind:      "float",
			want:            "float32(src.Rating)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildNumericCast(tt.sourceFieldName, tt.targetKind); got != tt.want {
				t.Errorf("BuildNumericCast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildTimestampConversion(t *testing.T) {
	tests := []struct {
		name            string
		sourceFieldName string
		wantToCode      string
		wantFromCode    string
	}{
		{
			name:            "CreatedAt field",
			sourceFieldName: "CreatedAt",
			wantToCode:      "timestampToTime(src.CreatedAt)",
			wantFromCode:    "timeToTimestamp(src.CreatedAt)",
		},
		{
			name:            "UpdatedAt field",
			sourceFieldName: "UpdatedAt",
			wantToCode:      "timestampToTime(src.UpdatedAt)",
			wantFromCode:    "timeToTimestamp(src.UpdatedAt)",
		},
		{
			name:            "Birthday field",
			sourceFieldName: "Birthday",
			wantToCode:      "timestampToTime(src.Birthday)",
			wantFromCode:    "timeToTimestamp(src.Birthday)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToCode, gotFromCode := BuildTimestampConversion(tt.sourceFieldName)
			if gotToCode != tt.wantToCode {
				t.Errorf("BuildTimestampConversion() toCode = %v, want %v", gotToCode, tt.wantToCode)
			}
			if gotFromCode != tt.wantFromCode {
				t.Errorf("BuildTimestampConversion() fromCode = %v, want %v", gotFromCode, tt.wantFromCode)
			}
		})
	}
}

// Note: IsTimestampToInt64 requires actual protogen.Field descriptors to test properly.
// This function is tested via integration tests in the GORM and Datastore generator
// test suites, where real proto descriptors are available from buf-generated code.
