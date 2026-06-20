/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"crypto"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

// Marshal converts a plain Go value (typically a struct) into the dynamic Value
// model. Struct fields are mapped to a Map keyed by their `value:"name"` tag (or
// the field name when untagged); a `value:"-"` tag skips the field. Go types map
// as: bool→Bool, signed/unsigned ints→Number(long), floats→Number(double),
// string→String(utf8), []byte / [N]byte→String(raw), other slices/arrays→List,
// map[string]T→Map, nested struct→Map. A field whose type already implements
// Value is used as-is. This is the reflection counterpart of Unmarshal; pair it
// with Pack to get the canonical MessagePack wire bytes.
//
// Unlike PackStruct (a strict numeric-tag schema where fields must themselves be
// Value), Marshal works on ordinary Go structs and is the codec used for wire
// payloads and the signing projection (see SignBytes/SignHash).
func Marshal(obj interface{}) (Value, error) {
	if obj == nil {
		return Null, nil
	}
	if v, ok := obj.(Value); ok {
		return v, nil
	}
	return toValue(reflect.ValueOf(obj))
}

// Unmarshal decodes a Value (produced by Marshal/Unpack) into the Go value
// pointed to by obj, using the same `value:"name"` field tags as Marshal. A
// missing or Null map entry leaves the destination field at its zero value.
func Unmarshal(v Value, obj interface{}) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("value: Unmarshal requires a non-nil pointer, got %T", obj)
	}
	return fromValue(v, rv.Elem())
}

// SignBytes returns the canonical MessagePack bytes of obj's signing
// projection: a Map built from only the struct fields tagged with the "sign"
// option (e.g. `value:"acc_id,sign"`), packed deterministically. Sign those
// bytes (e.g. ed25519.Sign) and verify against the same projection rebuilt by the
// peer — the canonical packing makes it stable across processes and field order.
// Include a domain field (a constant string tagged `value:"...,sign"`) per
// message type for domain separation.
func SignBytes(obj interface{}) ([]byte, error) {
	v, err := signProjection(obj)
	if err != nil {
		return nil, err
	}
	return Pack(v)
}

// SignHash is SignBytes followed by the given hash; it returns the digest.
func SignHash(obj interface{}, hash crypto.Hash) ([]byte, error) {
	v, err := signProjection(obj)
	if err != nil {
		return nil, err
	}
	_, digest, err := Hash(v, hash)
	return digest, err
}

// --- internals ---------------------------------------------------------------

// parseValueTag returns the map key for a struct field and its options. skip is
// true for unexported fields and for a `value:"-"` tag.
func parseValueTag(f reflect.StructField) (name string, opts map[string]struct{}, skip bool) {
	if f.PkgPath != "" {
		return "", nil, true // unexported
	}
	tag, ok := f.Tag.Lookup("value")
	if !ok {
		return f.Name, nil, false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "-" {
		return "", nil, true
	}
	if name == "" {
		name = f.Name
	}
	if len(parts) > 1 {
		opts = make(map[string]struct{}, len(parts)-1)
		for _, o := range parts[1:] {
			if o != "" {
				opts[o] = struct{}{}
			}
		}
	}
	return name, opts, false
}

func toValue(rv reflect.Value) (Value, error) {
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return Null, nil
		}
		if v, ok := rv.Interface().(Value); ok {
			return v, nil
		}
		rv = rv.Elem()
	}
	if rv.CanInterface() {
		if v, ok := rv.Interface().(Value); ok {
			return v, nil
		}
	}
	// time.Time is a struct but encodes as Unix milliseconds (deterministic,
	// location-independent), not a map of its unexported fields.
	if rv.Type() == timeType {
		return Long(rv.Interface().(time.Time).UnixMilli()), nil
	}
	switch rv.Kind() {
	case reflect.Bool:
		return Boolean(rv.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Long(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Long(int64(rv.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return Double(rv.Float()), nil
	case reflect.String:
		return Utf8(rv.String()), nil
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return Raw(rv.Bytes(), true), nil
		}
		return seqToList(rv)
	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, rv.Len())
			reflect.Copy(reflect.ValueOf(b), rv)
			return Raw(b, false), nil
		}
		return seqToList(rv)
	case reflect.Map:
		return mapToValue(rv)
	case reflect.Struct:
		return structToMap(rv, false)
	default:
		return nil, fmt.Errorf("value: cannot marshal kind %s", rv.Kind())
	}
}

func seqToList(rv reflect.Value) (Value, error) {
	items := make([]Value, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		v, err := toValue(rv.Index(i))
		if err != nil {
			return nil, err
		}
		items[i] = v
	}
	return ImmutableList(items), nil
}

func mapToValue(rv reflect.Value) (Value, error) {
	if rv.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("value: map key must be string, got %s", rv.Type().Key().Kind())
	}
	m := make(map[string]Value, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		v, err := toValue(iter.Value())
		if err != nil {
			return nil, err
		}
		m[iter.Key().String()] = v
	}
	return ImmutableMapOf(m), nil
}

func structToMap(rv reflect.Value, signOnly bool) (Value, error) {
	rt := rv.Type()
	m := make(map[string]Value)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		name, opts, skip := parseValueTag(f)
		if skip {
			continue
		}
		if signOnly {
			if _, ok := opts["sign"]; !ok {
				continue
			}
		}
		v, err := toValue(rv.Field(i))
		if err != nil {
			return nil, fmt.Errorf("value: field %q: %w", f.Name, err)
		}
		m[name] = v
	}
	return ImmutableMapOf(m), nil
}

func signProjection(obj interface{}) (Value, error) {
	rv := reflect.ValueOf(obj)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("value: SignBytes/SignHash got a nil pointer")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("value: SignBytes/SignHash require a struct, got %s", rv.Kind())
	}
	return structToMap(rv, true)
}

func fromValue(v Value, rv reflect.Value) error {
	// Destination is the Value interface (or a concrete Value): assign directly.
	if rv.Kind() == reflect.Interface && rv.Type() == ValueClass {
		if v == nil {
			rv.Set(reflect.ValueOf(Null))
		} else {
			rv.Set(reflect.ValueOf(v))
		}
		return nil
	}
	if rv.Kind() == reflect.Ptr {
		if v == nil || v.Kind() == NULL {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return fromValue(v, rv.Elem())
	}
	if v == nil || v.Kind() == NULL {
		return nil // leave the zero value
	}
	if rv.Type() == timeType {
		if n, ok := v.(Number); ok {
			rv.Set(reflect.ValueOf(time.UnixMilli(n.Long()).UTC()))
			return nil
		}
		return fmt.Errorf("value: cannot unmarshal %s into time.Time", v.Kind())
	}
	switch rv.Kind() {
	case reflect.Bool:
		if b, ok := v.(Bool); ok {
			rv.SetBool(b.Boolean())
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, ok := v.(Number); ok {
			rv.SetInt(n.Long())
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, ok := v.(Number); ok {
			rv.SetUint(uint64(n.Long()))
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if n, ok := v.(Number); ok {
			rv.SetFloat(n.Double())
			return nil
		}
	case reflect.String:
		if s, ok := v.(String); ok {
			rv.SetString(s.Utf8())
			return nil
		}
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			if s, ok := v.(String); ok {
				rv.SetBytes(s.Raw())
				return nil
			}
			break
		}
		if l, ok := v.(List); ok {
			vals := l.Values()
			out := reflect.MakeSlice(rv.Type(), len(vals), len(vals))
			for i, item := range vals {
				if err := fromValue(item, out.Index(i)); err != nil {
					return err
				}
			}
			rv.Set(out)
			return nil
		}
	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			if s, ok := v.(String); ok {
				reflect.Copy(rv, reflect.ValueOf(s.Raw()))
				return nil
			}
		}
	case reflect.Map:
		if m, ok := v.(Map); ok {
			return mapFromValue(m, rv)
		}
	case reflect.Struct:
		if m, ok := v.(Map); ok {
			return mapToStruct(m, rv)
		}
	}
	return fmt.Errorf("value: cannot unmarshal %s into %s", v.Kind(), rv.Type())
}

func mapFromValue(m Map, rv reflect.Value) error {
	if rv.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("value: map key must be string, got %s", rv.Type().Key().Kind())
	}
	out := reflect.MakeMapWithSize(rv.Type(), m.Len())
	elemType := rv.Type().Elem()
	for _, e := range m.Entries() {
		ev := reflect.New(elemType).Elem()
		if err := fromValue(e.Value(), ev); err != nil {
			return err
		}
		out.SetMapIndex(reflect.ValueOf(e.Key()), ev)
	}
	rv.Set(out)
	return nil
}

func mapToStruct(m Map, rv reflect.Value) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		name, _, skip := parseValueTag(f)
		if skip {
			continue
		}
		fv := m.Get(name)
		if fv == nil || fv.Kind() == NULL {
			continue
		}
		if err := fromValue(fv, rv.Field(i)); err != nil {
			return fmt.Errorf("value: field %q: %w", f.Name, err)
		}
	}
	return nil
}
