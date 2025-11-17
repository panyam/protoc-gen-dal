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

package converters

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// AnyToBytes converts a google.protobuf.Any to binary bytes.
// Returns empty bytes if any is nil.
// Returns error if marshaling fails.
func AnyToBytes(any *anypb.Any) ([]byte, error) {
	if any == nil {
		return nil, nil
	}
	return proto.Marshal(any)
}

// BytesToAny converts binary bytes back to google.protobuf.Any.
// Returns nil if bytes is empty.
// Returns error if unmarshaling fails.
func BytesToAny(data []byte) (*anypb.Any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	any := &anypb.Any{}
	if err := proto.Unmarshal(data, any); err != nil {
		return nil, err
	}
	return any, nil
}
