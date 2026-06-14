/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package value

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
)


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
		return fmt.Errorf("non-pointer instance is not allowed in '%v'", classPtr)
	}
	if schema, err := reflectSchema(classPtr); err != nil {
		return fmt.Errorf("error on reflect schema for '%v', %v", classPtr, err)
	} else {
		valuePtr := reflect.ValueOf(obj)
		value := valuePtr.Elem()
		return ParseStruct(unpacker, parser, value, schema)
	}
}

func reflectPackStruct(p *messagePacker, obj interface{}) error {
	classPtr := reflect.TypeOf(obj)
	if classPtr.Kind() != reflect.Ptr {
		return fmt.Errorf("non-pointer instance is not allowed in '%v'", classPtr)
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
	cnt := 0
	for _, field := range schema.SortedFields {
		f := &packingField {
			field: field,
			fieldValue: value.Field(field.FieldNum),
		}
		if !f.fieldValue.IsNil() {
			list = append(list, f)
			if f.field.Array && f.field.Repeated {
				cnt += f.fieldValue.Len()
			} else {
				cnt++
			}
		}
	}
	p.PackMap(cnt)
	for _, entry := range list {

		if entry.field.Array {

			cnt := entry.fieldValue.Len()
			if !entry.field.Repeated {
				p.PackLong(int64(entry.field.Tag))
				p.PackList(cnt)
			}

			for i := 0; i < cnt; i++ {
				if entry.field.Repeated {
					p.PackLong(int64(entry.field.Tag))
				}
				elem := entry.fieldValue.Index(i)
				if err := doReflectPackValue(p, elem, entry); err != nil {
					return err
				}
			}
		} else {
			p.PackLong(int64(entry.field.Tag))
			if err := doReflectPackValue(p,  entry.fieldValue, entry); err != nil {
				return err
			}
		}
	}
	return nil
}

func doReflectPackValue(p *messagePacker, value reflect.Value, entry *packingField) error {
	if entry.field.Struct {
		if err := doReflectPackStruct(p, value.Elem(), entry.field.FieldSchema); err != nil {
			return fmt.Errorf("can not pack field %v, inner struct error %v", value, err)
		}
	} else {
		fieldObject := value.Interface()
		if val, ok := fieldObject.(Value); ok {
			val.Pack(p)
		} else {
			return fmt.Errorf("can not convert field %v to value.Value", value)
		}
	}
	return nil
}

type Field struct {
	FieldNum       int
	FieldType      reflect.Type
	FieldName      string
	Array          bool
	Struct         bool
	Repeated       bool
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
		repeated := false
		if rep, ok := field.Tag.Lookup("repeated"); ok {
			repeated, _ = strconv.ParseBool(rep)
		}
		tagStr, ok := field.Tag.Lookup("tag")
		if !ok {
			return nil, fmt.Errorf("no tag in field '%s' in class '%v'", field.Name, classPtr)
		}
		tag, err := strconv.Atoi(tagStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tag number '%s' in field '%s' in class '%v'", tagStr, field.Name, classPtr)
		}
		array := false
		fieldType := field.Type
		if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
			array = true
		}
		if fieldType.Implements(ValueClass) {
			f := &Field{
				FieldNum:   j,
				FieldType:  field.Type,
				FieldName:  field.Name,
				Array:      array,
				Struct:     false,
				Repeated:   repeated,
				Tag:        tag,
			}
			fields[tag] = f
			sortedFields = append(sortedFields, f)
		} else if fieldType.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("tagged field '%s' in class '%v' with type '%v' does not implement value.Value interface and non-ptr", field.Name, field.Type, classPtr)
		} else if fieldSchema, err := reflectSchema(fieldType); err != nil {
			return nil, fmt.Errorf("struct field '%s' in class '%v' has wrong schema, %v", field.Name, classPtr, err)
		} else {
			f := &Field{
				FieldNum: j,
				FieldType: field.Type,
				FieldName: field.Name,
				Array:   array,
				Struct:   true,
				Repeated: repeated,
				FieldSchema: fieldSchema,
				Tag: tag,
			}
			fields[tag] = f
			sortedFields = append(sortedFields, f)
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
		return fmt.Errorf("expected MapHeader for struct, but got %v", format)
	}
	cnt := parser.ParseMap(header)
	if parser.Error() != nil {
		return parser.Error()
	}
	for i := 0; i < cnt; i++ {
		key, err := doParse(unpacker, parser)
		if err != nil {
			return fmt.Errorf("fail to parse key on position %d, %v", i, err)
		}
		if key.Kind() != NUMBER {
			return fmt.Errorf("expected int key, but got %s on position %d", key.Kind().String(), i)
		}
		tag := int(key.(Number).Long())
		if field, ok := schema.Fields[tag]; ok {
			fieldValue := value.Field(field.FieldNum)
			if field.Array {

				if !fieldValue.CanSet() {
					return fmt.Errorf("can not set empty slice value to field %v", field.FieldName)
				}

				if !field.Repeated {
					listFormat, listHeader := unpacker.Next()
					if listFormat != ListHeader {
						return fmt.Errorf("expected ListHeader for array field, but got %v", listFormat)
					}
					listCnt := parser.ParseList(listHeader)
					arrayType := reflect.ArrayOf(listCnt, field.FieldType.Elem())
					arrayValue := reflect.New(arrayType).Elem()

					for j := 0; j < listCnt; j++ {
						elemValue := arrayValue.Index(j)
						if field.Struct {
							structValue := reflect.New(elemValue.Type().Elem())
							elemValue.Set(structValue)
							err := ParseStruct(unpacker, parser, elemValue.Elem(), field.FieldSchema)
							if err != nil {
								return fmt.Errorf("fail to set struct value %v", err)
							}
						} else {
							val, err := doParse(unpacker, parser)
							if err != nil {
								return fmt.Errorf("fail to parse value %v", err)
							}
							err = setFieldValue(elemValue, field.FieldType.Elem(), val)
							if err != nil {
								return fmt.Errorf("fail to set value %v", err)
							}
						}
					}
					fieldValue.Set(arrayValue.Slice(0, listCnt))
				} else {
					var sliceValue reflect.Value
					if !fieldValue.IsNil() {
						sliceValue = fieldValue
					} else {
						sliceValue = reflect.New(field.FieldType).Elem()
					}
					var elemValue reflect.Value
					if field.Struct {
						ptrType := field.FieldType.Elem()
						elemValue := reflect.New(ptrType).Elem()
						structValue := reflect.New(ptrType.Elem())
						elemValue.Set(structValue)
						sliceValue = reflect.Append(sliceValue, elemValue)
						err := ParseStruct(unpacker, parser, elemValue.Elem(), field.FieldSchema)
						if err != nil {
							return fmt.Errorf("fail to set struct value %v", err)
						}
					} else {
						elemValue = reflect.New(field.FieldType.Elem()).Elem()
						sliceValue = reflect.Append(sliceValue, elemValue)
						val, err := doParse(unpacker, parser)
						if err != nil {
							return fmt.Errorf("fail to parse value %v", err)
						}
						err = setFieldValue(elemValue, field.FieldType.Elem(), val)
						if err != nil {
							return fmt.Errorf("fail to set value %v", err)
						}
					}
					fieldValue.Set(sliceValue)
				}
			} else {
				err = parseFieldValue(unpacker, parser, field, fieldValue)
				if err != nil {
					return fmt.Errorf("parse field on position %d, %v", i, err)
				}
			}
		} else {
			return fmt.Errorf("unknown tag %d on position %d", tag, i)
		}
	}
	return nil
}

func parseFieldValue(unpacker Unpacker, parser Parser, field *Field, fieldValue reflect.Value) error {
	if field.Struct {
		if fieldValue.IsNil() {
			if fieldValue.CanSet() {
				fieldValue.Set(reflect.New(field.FieldType.Elem()))
			} else {
				return fmt.Errorf("can not set empty struct value to field %v", field.FieldName)
			}
		}
		err := ParseStruct(unpacker, parser, fieldValue.Elem(), field.FieldSchema)
		if err != nil {
			return fmt.Errorf("fail to set struct value %v", err)
		}
	} else {
		val, err := doParse(unpacker, parser)
		if err != nil {
			return fmt.Errorf("fail to parse value %v", err)
		}
		err = setFieldValue(fieldValue, field.FieldType, val)
		if err != nil {
			return fmt.Errorf("fail to set value %v", err)
		}
	}
	return nil
}


func setFieldValue(fieldValue reflect.Value, fieldType reflect.Type, val Value) error {
	if fieldValue.CanSet() {
		if !val.Class().AssignableTo(fieldType) {
			return fmt.Errorf("expected value type %v, actual %v", fieldType, val.Class())
		}
		value := reflect.ValueOf(val)
		fieldValue.Set(value)
		return nil
	} else {
		return fmt.Errorf("can not set value '%v' to field %v", val, fieldType)
	}
}


