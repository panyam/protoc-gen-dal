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

// MessageToAnyBytes converts a proto message to binary bytes via google.protobuf.Any.
// This packs the message into Any, then marshals to bytes.
// Returns empty bytes if message is nil.
// Returns error if packing or marshaling fails.
func MessageToAnyBytes(msg proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, nil
	}
	// Pack message into Any
	any, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}
	// Marshal Any to bytes
	return proto.Marshal(any)
}

// AnyBytesToMessage converts binary bytes back to a typed proto message via google.protobuf.Any.
// This unmarshals bytes to Any, then unpacks to the typed message.
// Returns nil if bytes is empty.
// Returns error if unmarshaling or unpacking fails.
// Usage: converters.AnyBytesToMessage[*api.Path](data)
func AnyBytesToMessage[T proto.Message](data []byte) (T, error) {
	var zero T
	if len(data) == 0 {
		return zero, nil
	}
	// Unmarshal bytes to Any
	any := &anypb.Any{}
	if err := proto.Unmarshal(data, any); err != nil {
		return zero, err
	}
	// Unpack Any to typed message
	msg, err := any.UnmarshalNew()
	if err != nil {
		return zero, err
	}
	// Type assert to expected type
	typed, ok := msg.(T)
	if !ok {
		return zero, nil // Return nil if type doesn't match
	}
	return typed, nil
}

// MessageToAnyBytesConverter converts a proto message to Any bytes with standard converter signature.
// This is a wrapper around MessageToAnyBytes that matches the standard converter function signature
// (dest, src, decorator) used in generated code loops.
// Returns the dest parameter for chaining (always nil for this converter).
func MessageToAnyBytesConverter(src proto.Message, dest *[]byte, decorator func(*[]byte)) (*[]byte, error) {
	bytes, err := MessageToAnyBytes(src)
	if err != nil {
		return nil, err
	}
	*dest = bytes
	if decorator != nil {
		decorator(dest)
	}
	return dest, nil
}

// AnyBytesToMessageConverter converts Any bytes to a typed message with standard converter signature.
// This is a wrapper around AnyBytesToMessage that matches the standard converter function signature
// (dest, src, decorator) used in generated code loops.
// Returns the converted message pointer.
func AnyBytesToMessageConverter[T proto.Message](dest T, src *[]byte, decorator func(T)) (T, error) {
	msg, err := AnyBytesToMessage[T](*src)
	if err != nil {
		var zero T
		return zero, err
	}
	if decorator != nil {
		decorator(msg)
	}
	return msg, nil
}
