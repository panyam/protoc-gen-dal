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
	"strings"
	"testing"
)

func TestTimestampHelperFunctions(t *testing.T) {
	code := TimestampHelperFunctions()

	// Check that the code contains expected function signatures
	if !strings.Contains(code, "func timestampToTime") {
		t.Error("TimestampHelperFunctions() should contain timestampToTime function")
	}

	if !strings.Contains(code, "func timeToTimestamp") {
		t.Error("TimestampHelperFunctions() should contain timeToTimestamp function")
	}

	// Check for key conversion logic
	if !strings.Contains(code, "ts.AsTime()") {
		t.Error("TimestampHelperFunctions() should contain timestamp to time.Time conversion")
	}

	if !strings.Contains(code, "timestamppb.New(t)") {
		t.Error("TimestampHelperFunctions() should contain time.Time to timestamp conversion")
	}

	// Check for nil/zero handling
	if !strings.Contains(code, "if ts == nil") {
		t.Error("TimestampHelperFunctions() should handle nil timestamps")
	}

	if !strings.Contains(code, "t.IsZero()") {
		t.Error("TimestampHelperFunctions() should handle zero time values")
	}
}

func TestMustParseUintHelper(t *testing.T) {
	code := MustParseUintHelper()

	// Check that the code contains expected function signature
	if !strings.Contains(code, "func mustParseUint") {
		t.Error("MustParseUintHelper() should contain mustParseUint function")
	}

	// Check for strconv.ParseUint usage
	if !strings.Contains(code, "strconv.ParseUint") {
		t.Error("MustParseUintHelper() should use strconv.ParseUint")
	}

	// Check for panic on error
	if !strings.Contains(code, "panic") {
		t.Error("MustParseUintHelper() should panic on parse error")
	}

	// Check for base 10 parsing
	if !strings.Contains(code, ", 10, 32)") {
		t.Error("MustParseUintHelper() should parse base 10, 32-bit uint")
	}
}
