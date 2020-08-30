/*
 *
 * Copyright 2020-present Arpabet, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package value

import (
	"bytes"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

/**
	@author Alex Shvid
*/

func PackStruct(obj interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	p := MessagePacker(&buf)
	if obj != nil {
		if val, ok := obj.(Value); ok {
			val.Pack(p)
		} else if err := reflectPackStruct(p, obj); err != nil {
			return nil, err
		}
	} else {
		p.PackNil()
	}
	return buf.Bytes(), p.Error()
}

func UnpackStruct(buf []byte, obj interface{}, copy bool) error {
	unpacker := MessageUnpacker(buf, copy)
	parser := MessageParser()
	classPtr := reflect.TypeOf(obj)
	if classPtr.Kind() != reflect.Ptr {
		return errors.Errorf("non-pointer instance is not allowed in '%v'", classPtr)
	}
	if schema, err := reflectSchema(classPtr); err != nil {
		return errors.Errorf("error on reflect schema for '%v', %v", classPtr, err)
	} else {
		valuePtr := reflect.ValueOf(obj)
		value := valuePtr.Elem()
		return ParseStruct(unpacker, parser, value, schema)
	}
}

func reflectPackStruct(p *messagePacker, obj interface{}) error {
	classPtr := reflect.TypeOf(obj)
	if classPtr.Kind() != reflect.Ptr {
		return errors.Errorf("non-pointer instance is not allowed in '%v'", classPtr)
	}
	schema, err := reflectSchema(classPtr)
	if err != nil {
		return err
	}
	valuePtr := reflect.ValueOf(obj)
	value := valuePtr.Elem()
	return doReflectPackStruct(p, value, schema)
}

type packingField struct {
	field       *Field
	fieldValue  reflect.Value
}

func doReflectPackStruct(p *messagePacker, value reflect.Value, schema *Schema) error {
	var list []*packingField
	for _, field := range schema.SortedFields {
		f := &packingField {
			field: field,
			fieldValue: value.Field(field.FieldNum),
		}
		if !f.fieldValue.IsNil() {
			list = append(list, f)
		}
	}
	p.PackMap(len(list))
	for _, entry := range list {
		p.PackLong(int64(entry.field.Tag))
		if entry.field.ValueField {
			fieldObject := entry.fieldValue.Interface()
			if val, ok := fieldObject.(Value); ok {
				val.Pack(p)
			} else {
				return errors.Errorf("can not convert field %v to value.Value", entry.fieldValue)
			}
		} else {
			if err := doReflectPackStruct(p, entry.fieldValue.Elem(), entry.field.FieldSchema); err != nil {
				return errors.Errorf("can not pack field %v, inner struct error %v", entry.fieldValue, err)
			}
		}
	}
	return nil
}

type Field struct {
	FieldNum       int
	FieldType      reflect.Type
	ValueField     bool
	FieldSchema    *Schema
	Tag            int
}

type Schema struct {
	Fields        map[int]*Field   // tag is the key
	SortedFields  []*Field
}

var schemaCache sync.Map

type sortableFields []*Field

func (t sortableFields) Len() int {
	return len(t)
}

func (t sortableFields) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t sortableFields) Less(i, j int) bool {
	return t[i].Tag < t[j].Tag
}

func reflectSchema(classPtr reflect.Type) (*Schema, error) {
	if val, ok := schemaCache.Load(classPtr); ok {
		return val.(*Schema), nil
	} else if schema, err := doReflectSchema(classPtr); err != nil {
		return nil, err
	} else {
		schemaCache.Store(classPtr, schema)
		return schema, nil
	}
}

func doReflectSchema(classPtr reflect.Type) (*Schema, error) {
	fields := make(map[int]*Field)
	var sortedFields []*Field
	class := classPtr.Elem()
	for j := 0; j < class.NumField(); j++ {
		field := class.Field(j)
		if tagStr, ok := field.Tag.Lookup("tag"); ok {
			if tag, err := strconv.Atoi(tagStr); err != nil {
				return nil, errors.Errorf("invalid tag number '%s' in field '%s' in class '%v'", tagStr, field.Name, classPtr)
			} else if field.Type.Implements(ValueClass) {
				f := &Field{
					FieldNum: j,
					FieldType: field.Type,
					ValueField: true,
					Tag: tag,
				}
				fields[tag] = f
				sortedFields = append(sortedFields, f)
			} else if field.Type.Kind() != reflect.Ptr {
				return nil, errors.Errorf("tagged field '%s' in class '%v' does not implement value.Value interface and non-ptr", field.Name, classPtr)
			} else if fieldSchema, err := reflectSchema(field.Type); err != nil {
				return nil, errors.Errorf("struct field '%s' in class '%v' has wrong schema, %v", field.Name, classPtr, err)
			} else {
				f := &Field{
					FieldNum: j,
					FieldType: field.Type,
					ValueField: false,
					FieldSchema: fieldSchema,
					Tag: tag,
				}
				fields[tag] = f
				sortedFields = append(sortedFields, f)
			}
		} else {
			return nil, errors.Errorf("no tag in field '%s' in class '%v'", field.Name, classPtr)
		}
	}
	sort.Sort(sortableFields(sortedFields))
	return &Schema {
		Fields: fields,
		SortedFields: sortedFields,
	}, nil
}


func ParseStruct(unpacker Unpacker, parser Parser, value reflect.Value, schema *Schema) error {
	format, header := unpacker.Next()
	if format != MapHeader {
		return errors.Errorf("expected MapHeader for struct, but got %v", format)
	}
	cnt := parser.ParseMap(header)
	if parser.Error() != nil {
		return parser.Error()
	}
	for i := 0; i < cnt; i++ {
		key, err := doParse(unpacker, parser)
		if err != nil {
			return errors.Errorf("fail to parse key on position %d, %v", i, err)
		}
		if key.Kind() != NUMBER {
			return errors.Errorf("expected int key, but got %s on position %d", key.Kind().String(), i)
		}
		tag := int(key.(Number).Long())
		if field, ok := schema.Fields[tag]; ok {
			if field.ValueField {
				val, err := doParse(unpacker, parser)
				if err != nil {
					return errors.Errorf("fail to parse value on position %d, %v", i, err)
				}
				err = setFieldValue(value.Field(field.FieldNum), field.FieldType, val)
				if err != nil {
					return errors.Errorf("fail to set value on position %d, %v", i, err)
				}
			} else {
				fieldValue := value.Field(field.FieldNum)
				if fieldValue.IsNil() {
					if fieldValue.CanSet() {
						fieldValue.Set(reflect.New(field.FieldType.Elem()))
					} else {
						return errors.Errorf("can not set empty value to field %v", fieldValue)
					}
				}
				err := ParseStruct(unpacker, parser, fieldValue.Elem(), field.FieldSchema)
				if err != nil {
					return errors.Errorf("fail to set struct value on position %d, %v", i, err)
				}
			}
		} else {
			return errors.Errorf("unknown tag %d on position %d", tag, i)
		}
	}
	return nil
}


func setFieldValue(fieldValue reflect.Value, fieldType reflect.Type, val Value) error {
	if fieldValue.CanSet() {
		if !val.Class().AssignableTo(fieldType) {
			return errors.Errorf("expected value type %v, actual %v", fieldType, val.Class())
		}
		value := reflect.ValueOf(val)
		fieldValue.Set(value)
		return nil
	} else {
		return errors.Errorf("can not set value %v to field %v", val, fieldValue)
	}
}


