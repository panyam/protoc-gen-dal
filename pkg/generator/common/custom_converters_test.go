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

package common

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	dalv1 "github.com/panyam/protoc-gen-dal/protos/gen/dal/v1"
)

func TestCollectCustomConverterImports(t *testing.T) {
	tests := []struct {
		name     string
		fields   []*protogen.Field
		want     []ImportSpec
	}{
		{
			name:   "no custom converters",
			fields: []*protogen.Field{},
			want:   []ImportSpec{},
		},
		{
			name: "to_func with explicit alias",
			fields: []*protogen.Field{
				createFieldWithCustomConverter(
					&dalv1.ConverterFunc{
						Package:  "github.com/test/converters",
						Alias:    "conv",
						Function: "ToMillis",
					},
					nil,
				),
			},
			want: []ImportSpec{
				{Path: "github.com/test/converters", Alias: "conv"},
			},
		},
		{
			name: "to_func without alias (auto-generated)",
			fields: []*protogen.Field{
				createFieldWithCustomConverter(
					&dalv1.ConverterFunc{
						Package:  "github.com/test/converters",
						Function: "ToMillis",
					},
					nil,
				),
			},
			want: []ImportSpec{
				{Path: "github.com/test/converters", Alias: "converters"},
			},
		},
		{
			name: "from_func with explicit alias",
			fields: []*protogen.Field{
				createFieldWithCustomConverter(
					nil,
					&dalv1.ConverterFunc{
						Package:  "github.com/test/utils",
						Alias:    "u",
						Function: "FromMillis",
					},
				),
			},
			want: []ImportSpec{
				{Path: "github.com/test/utils", Alias: "u"},
			},
		},
		{
			name: "both to_func and from_func",
			fields: []*protogen.Field{
				createFieldWithCustomConverter(
					&dalv1.ConverterFunc{
						Package:  "github.com/test/converters",
						Alias:    "conv",
						Function: "ToMillis",
					},
					&dalv1.ConverterFunc{
						Package:  "github.com/test/utils",
						Alias:    "u",
						Function: "FromMillis",
					},
				),
			},
			want: []ImportSpec{
				{Path: "github.com/test/converters", Alias: "conv"},
				{Path: "github.com/test/utils", Alias: "u"},
			},
		},
		{
			name: "multiple fields with custom converters",
			fields: []*protogen.Field{
				createFieldWithCustomConverter(
					&dalv1.ConverterFunc{
						Package:  "github.com/test/converters",
						Function: "ToMillis",
					},
					nil,
				),
				createFieldWithCustomConverter(
					nil,
					&dalv1.ConverterFunc{
						Package:  "github.com/test/formatters",
						Function: "Format",
					},
				),
			},
			want: []ImportSpec{
				{Path: "github.com/test/converters", Alias: "converters"},
				{Path: "github.com/test/formatters", Alias: "formatters"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := make(ImportMap)
			msg := &protogen.Message{
				Fields: tt.fields,
			}

			CollectCustomConverterImports(msg, imports)

			got := imports.ToSlice()

			if len(got) != len(tt.want) {
				t.Fatalf("got %d imports, want %d", len(got), len(tt.want))
			}

			// Check that all expected imports are present
			for _, wantImport := range tt.want {
				found := false
				for _, gotImport := range got {
					if gotImport.Path == wantImport.Path && gotImport.Alias == wantImport.Alias {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing import: %+v", wantImport)
				}
			}
		})
	}
}

func TestExtractCustomConverters(t *testing.T) {
	tests := []struct {
		name         string
		field        *protogen.Field
		fieldName    string
		wantToCode   string
		wantFromCode string
	}{
		{
			name:         "no custom converters",
			field:        createFieldWithCustomConverter(nil, nil),
			fieldName:    "TestField",
			wantToCode:   "",
			wantFromCode: "",
		},
		{
			name: "to_func with explicit alias",
			field: createFieldWithCustomConverter(
				&dalv1.ConverterFunc{
					Package:  "github.com/test/converters",
					Alias:    "conv",
					Function: "ToMillis",
				},
				nil,
			),
			fieldName:    "CreatedAt",
			wantToCode:   "conv.ToMillis(src.CreatedAt)",
			wantFromCode: "",
		},
		{
			name: "to_func without alias",
			field: createFieldWithCustomConverter(
				&dalv1.ConverterFunc{
					Package:  "github.com/test/converters",
					Function: "ToMillis",
				},
				nil,
			),
			fieldName:    "CreatedAt",
			wantToCode:   "converters.ToMillis(src.CreatedAt)",
			wantFromCode: "",
		},
		{
			name: "from_func with explicit alias",
			field: createFieldWithCustomConverter(
				nil,
				&dalv1.ConverterFunc{
					Package:  "github.com/test/utils",
					Alias:    "u",
					Function: "FromMillis",
				},
			),
			fieldName:    "CreatedAt",
			wantToCode:   "",
			wantFromCode: "u.FromMillis(src.CreatedAt)",
		},
		{
			name: "both to_func and from_func",
			field: createFieldWithCustomConverter(
				&dalv1.ConverterFunc{
					Package:  "github.com/test/converters",
					Alias:    "conv",
					Function: "ToMillis",
				},
				&dalv1.ConverterFunc{
					Package:  "github.com/test/utils",
					Alias:    "u",
					Function: "FromMillis",
				},
			),
			fieldName:    "CreatedAt",
			wantToCode:   "conv.ToMillis(src.CreatedAt)",
			wantFromCode: "u.FromMillis(src.CreatedAt)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToCode, gotFromCode := ExtractCustomConverters(tt.field, tt.fieldName)

			if gotToCode != tt.wantToCode {
				t.Errorf("toCode = %q, want %q", gotToCode, tt.wantToCode)
			}
			if gotFromCode != tt.wantFromCode {
				t.Errorf("fromCode = %q, want %q", gotFromCode, tt.wantFromCode)
			}
		})
	}
}

// mockFieldDescriptor is a minimal implementation for testing custom converters
// We only need to implement Options() - all other methods can use zero values
type mockFieldDescriptor struct {
	protoreflect.FieldDescriptor // Embed to get default implementations
	options                      *descriptorpb.FieldOptions
}

func (m *mockFieldDescriptor) Options() protoreflect.ProtoMessage {
	if m.options == nil {
		return (*descriptorpb.FieldOptions)(nil)
	}
	return m.options
}

// Implement the minimal required methods for FieldDescriptor
func (m *mockFieldDescriptor) Name() protoreflect.Name                         { return "" }
func (m *mockFieldDescriptor) FullName() protoreflect.FullName                 { return "" }
func (m *mockFieldDescriptor) Index() int                                      { return 0 }
func (m *mockFieldDescriptor) Syntax() protoreflect.Syntax                     { return protoreflect.Proto3 }
func (m *mockFieldDescriptor) IsExtension() bool                               { return false }
func (m *mockFieldDescriptor) IsWeak() bool                                    { return false }
func (m *mockFieldDescriptor) IsPacked() bool                                  { return false }
func (m *mockFieldDescriptor) IsList() bool                                    { return false }
func (m *mockFieldDescriptor) IsMap() bool                                     { return false }
func (m *mockFieldDescriptor) MapKey() protoreflect.FieldDescriptor            { return nil }
func (m *mockFieldDescriptor) MapValue() protoreflect.FieldDescriptor          { return nil }
func (m *mockFieldDescriptor) HasDefault() bool                                { return false }
func (m *mockFieldDescriptor) Default() protoreflect.Value                     { return protoreflect.Value{} }
func (m *mockFieldDescriptor) DefaultEnumValue() protoreflect.EnumValueDescriptor { return nil }
func (m *mockFieldDescriptor) ContainingOneof() protoreflect.OneofDescriptor   { return nil }
func (m *mockFieldDescriptor) ContainingMessage() protoreflect.MessageDescriptor { return nil }
func (m *mockFieldDescriptor) Enum() protoreflect.EnumDescriptor               { return nil }
func (m *mockFieldDescriptor) Message() protoreflect.MessageDescriptor         { return nil }
func (m *mockFieldDescriptor) Number() protoreflect.FieldNumber                { return 1 }
func (m *mockFieldDescriptor) Cardinality() protoreflect.Cardinality           { return protoreflect.Optional }
func (m *mockFieldDescriptor) Kind() protoreflect.Kind                         { return protoreflect.Int64Kind }
func (m *mockFieldDescriptor) HasJSONName() bool                               { return false }
func (m *mockFieldDescriptor) JSONName() string                                { return "" }
func (m *mockFieldDescriptor) TextName() string                                { return "" }
func (m *mockFieldDescriptor) HasPresence() bool                               { return false }
func (m *mockFieldDescriptor) HasOptionalKeyword() bool                        { return false }
func (m *mockFieldDescriptor) ProtoType(protoreflect.FieldDescriptor) {}

// Helper to create a proto field with custom converter options
func createFieldWithCustomConverter(toFunc, fromFunc *dalv1.ConverterFunc) *protogen.Field {
	var opts *descriptorpb.FieldOptions

	// Add column options if custom converters are specified
	if toFunc != nil || fromFunc != nil {
		opts = &descriptorpb.FieldOptions{}
		colOpts := &dalv1.ColumnOptions{
			ToFunc:   toFunc,
			FromFunc: fromFunc,
		}
		proto.SetExtension(opts, dalv1.E_Column, colOpts)
	}

	return &protogen.Field{
		Desc: &mockFieldDescriptor{options: opts},
	}
}
