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

// TimestampHelperFunctions returns the Go source code for timestamp conversion helpers.
//
// These helper functions convert between google.protobuf.Timestamp and time.Time.
// Used by both GORM and Datastore converters when storing timestamps.
//
// Returns:
//   - Go source code string containing timestampToTime and timeToTimestamp functions
func TimestampHelperFunctions() string {
	return `
// timestampToTime converts a protobuf Timestamp to time.Time.
func timestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// timeToTimestamp converts time.Time to a protobuf Timestamp.
func timeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
`
}

// MustParseUintHelper returns the Go source code for the mustParseUint helper.
//
// This helper is Datastore-specific. Datastore stores entity IDs as strings, so uint32
// fields need to be converted to/from strings. This function provides panic-on-error
// parsing suitable for generated code.
//
// Returns:
//   - Go source code string containing mustParseUint function
func MustParseUintHelper() string {
	return `
// mustParseUint parses a string to uint64, panics on error (for generated code).
func mustParseUint(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic("failed to parse uint: " + err.Error())
	}
	return val
}
`
}
