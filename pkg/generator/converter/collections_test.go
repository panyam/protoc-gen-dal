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

func TestBuildNestedConverterName(t *testing.T) {
	tests := []struct {
		name         string
		sourceType   string
		targetType   string
		wantToFunc   string
		wantFromFunc string
	}{
		{
			name:         "GORM target",
			sourceType:   "Author",
			targetType:   "AuthorGORM",
			wantToFunc:   "AuthorToAuthorGORM",
			wantFromFunc: "AuthorFromAuthorGORM",
		},
		{
			name:         "Datastore target",
			sourceType:   "User",
			targetType:   "UserDatastore",
			wantToFunc:   "UserToUserDatastore",
			wantFromFunc: "UserFromUserDatastore",
		},
		{
			name:         "Same base name",
			sourceType:   "Book",
			targetType:   "Book",
			wantToFunc:   "BookToBook",
			wantFromFunc: "BookFromBook",
		},
		{
			name:         "Nested type",
			sourceType:   "Address",
			targetType:   "AddressGORM",
			wantToFunc:   "AddressToAddressGORM",
			wantFromFunc: "AddressFromAddressGORM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToFunc, gotFromFunc := BuildNestedConverterName(tt.sourceType, tt.targetType)
			if gotToFunc != tt.wantToFunc {
				t.Errorf("BuildNestedConverterName() toFunc = %v, want %v", gotToFunc, tt.wantToFunc)
			}
			if gotFromFunc != tt.wantFromFunc {
				t.Errorf("BuildNestedConverterName() fromFunc = %v, want %v", gotFromFunc, tt.wantFromFunc)
			}
		})
	}
}

// Note: CheckMapValueType, CheckRepeatedElementType, ExtractMapMessages, and ExtractRepeatedMessages
// require actual protogen.Field descriptors to test properly. These functions are tested via
// integration tests in the GORM and Datastore generator test suites, where real proto descriptors
// are available from buf-generated code.
//
// The functions are designed to be simple field descriptor inspections with minimal logic,
// making them safe to test in integration rather than requiring complex mocking here.
