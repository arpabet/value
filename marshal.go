/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"crypto"
	"golang.org/x/xerrors"
	"reflect"
	"strings"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

type valueCodecProfile uint8

const (
	valueTags valueCodecProfile = iota
	jsonTags
	msgpackTags
)

// ValueCodec converts ordinary Go objects to and from the canonical Value data
// model using one struct-tag dialect. It is immutable and safe for concurrent
// use. The zero value is the classic `value:"field"` codec.
//
// A ValueCodec selects object-to-Value schema metadata; it does not select the
// final byte encoding. Call Pack/Unpack separately when canonical MessagePack
// bytes are required.
type ValueCodec struct {
	profile valueCodecProfile
}

// DefaultValueCodec returns the classic codec using `value:"field"` tags.
// Missing fields are required unless tagged `omitempty`, preserving the
// package-level Marshal/Unmarshal behavior.
func DefaultValueCodec() ValueCodec {
	return ValueCodec{profile: valueTags}
}

// JSONValueCodec returns a codec using `json:"field"` tags to name fields in
// the Value tree. Missing fields are accepted, as they are by encoding/json.
// This codec still maps []byte to Value binary strings and time.Time to Unix
// milliseconds; it does not produce or consume textual JSON.
func JSONValueCodec() ValueCodec {
	return ValueCodec{profile: jsonTags}
}

// MsgpackValueCodec returns a codec using `msgpack:"field"` tags to name fields
// in the Value tree. Missing fields are accepted. It provides a migration path
// for structs carrying common MessagePack field-name/omitempty metadata, but it
// does not reproduce library-specific MessagePack bytes or extension hooks.
func MsgpackValueCodec() ValueCodec {
	return ValueCodec{profile: msgpackTags}
}

func (c ValueCodec) tagName() string {
	switch c.profile {
	case jsonTags:
		return "json"
	case msgpackTags:
		return "msgpack"
	default:
		return "value"
	}
}

func (c ValueCodec) allowsMissingFields() bool {
	return c.profile != valueTags
}

// Marshal converts a plain Go value (typically a struct) into the dynamic Value
// model. Struct fields are mapped to a Map keyed by their `value:"name"` tag (or
// the field name when untagged); a `value:"-"` tag skips the field. Go types map
// as: bool→Bool, signed/unsigned ints→Number(long), floats→Number(double),
// string→String(utf8), []byte / [N]byte→String(raw), other slices/arrays→List,
// map[string]T→Map, nested struct→Map. A field whose type already implements
// Value is used as-is. A field tagged with the `omitempty` option (e.g.
// `value:"sig,omitempty"`) is dropped when it holds its zero value, matching
// encoding/json; a field without it is always written and is required by
// Unmarshal. This is the reflection counterpart of Unmarshal; pair it with Pack
// to get the canonical MessagePack wire bytes.
//
// Unlike PackStruct (a strict numeric-tag schema where fields must themselves be
// Value), Marshal works on ordinary Go structs and is the codec used for wire
// payloads and the signing projection (see SignBytes/SignHash).
func Marshal(obj interface{}) (Value, error) {
	return DefaultValueCodec().Marshal(obj)
}

// Marshal converts obj into a Value tree using c's struct-tag dialect.
func (c ValueCodec) Marshal(obj interface{}) (Value, error) {
	if obj == nil {
		return Null, nil
	}
	if v, ok := obj.(Value); ok {
		return v, nil
	}
	return c.toValue(reflect.ValueOf(obj))
}

// Unmarshal decodes a Value (produced by Marshal/Unpack) into the Go value
// pointed to by obj, using the same `value:"name"` field tags as Marshal. A field
// is required unless its tag carries the `omitempty` option: Unmarshal returns an
// error if a required field's key is absent from the map. An absent optional
// (omitempty) field — or any present-but-Null entry — leaves the destination at
// its zero value.
func Unmarshal(v Value, obj interface{}) error {
	return DefaultValueCodec().Unmarshal(v, obj)
}

// Unmarshal decodes a Value tree into obj using c's struct-tag dialect.
func (c ValueCodec) Unmarshal(v Value, obj interface{}) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return xerrors.Errorf("value: Unmarshal requires a non-nil pointer, got %T", obj)
	}
	return c.fromValue(v, rv.Elem())
}

// SignBytes returns the canonical MessagePack bytes of obj's signing projection
// for the given selector: a Map built from only the struct fields whose `sign`
// tag carries that selector, packed deterministically. For example,
// `SignBytes(o, "license")` includes fields tagged `sign:"license"`. A field may
// carry several selectors (`sign:"license,audit"`) to belong to independent
// signatures.
//
// For backward compatibility, selectors embedded in a `value` tag continue to
// work: `value:"acc_id,sign"` is selected by `SignBytes(o, "sign")`. New code
// should keep field naming and signing policy separate with
// `value:"acc_id" sign:"sign"`.
//
// The selector must be non-empty. Sign these bytes (e.g. ed25519.Sign) and
// verify against the same projection rebuilt by the peer — canonical packing
// makes it stable across processes and field order. Include a domain field (a
// constant string tagged with the same selector) per message type for domain
// separation.
func SignBytes(obj interface{}, selector string) ([]byte, error) {
	return DefaultValueCodec().SignBytes(obj, selector)
}

// SignBytes returns the canonical signing projection using c's field names.
// The dedicated `sign:"selector"` tag works with every profile. Legacy signing
// selectors embedded in `value` options are recognized only by the default
// profile.
func (c ValueCodec) SignBytes(obj interface{}, selector string) ([]byte, error) {
	v, err := c.signProjection(obj, selector)
	if err != nil {
		return nil, err
	}
	return Pack(v)
}

// SignHash is SignBytes followed by the given hash; it returns the digest.
func SignHash(obj interface{}, hash crypto.Hash, selector string) ([]byte, error) {
	return DefaultValueCodec().SignHash(obj, hash, selector)
}

// SignHash is c.SignBytes followed by the selected hash.
func (c ValueCodec) SignHash(obj interface{}, hash crypto.Hash, selector string) ([]byte, error) {
	v, err := c.signProjection(obj, selector)
	if err != nil {
		return nil, err
	}
	_, digest, err := Hash(v, hash)
	return digest, err
}

// --- internals ---------------------------------------------------------------

// parseStructTag returns the map key for a struct field and its options. skip is
// true for unexported fields and for the active dialect's `-` tag.
func (c ValueCodec) parseStructTag(f reflect.StructField) (name string, opts map[string]struct{}, skip bool, err error) {
	if f.PkgPath != "" {
		return "", nil, true, nil // unexported
	}
	tagName := c.tagName()
	tag, ok := f.Tag.Lookup(tagName)
	if !ok {
		return f.Name, nil, false, nil
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "-" {
		return "", nil, true, nil
	}
	if name == "" {
		name = f.Name
	}
	if len(parts) > 1 {
		opts = make(map[string]struct{}, len(parts)-1)
		for _, o := range parts[1:] {
			if o != "" {
				if c.profile != valueTags && o != "omitempty" {
					return "", nil, false, xerrors.Errorf("value: unsupported %s tag option %q on field %s", tagName, o, f.Name)
				}
				opts[o] = struct{}{}
			}
		}
	}
	return name, opts, false, nil
}

// hasSignSelector reports whether a field belongs to selector. The dedicated
// sign tag is the preferred spelling. Arbitrary value-tag options are retained
// as legacy selectors so existing signed payloads keep exactly the same
// projection during migration.
func (c ValueCodec) hasSignSelector(f reflect.StructField, fieldOpts map[string]struct{}, selector string) bool {
	if c.profile == valueTags {
		if _, ok := fieldOpts[selector]; ok {
			return true
		}
	}
	tag, ok := f.Tag.Lookup("sign")
	if !ok || tag == "" || tag == "-" {
		return false
	}
	for _, candidate := range strings.Split(tag, ",") {
		if strings.TrimSpace(candidate) == selector {
			return true
		}
	}
	return false
}

func (c ValueCodec) toValue(rv reflect.Value) (Value, error) {
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
		return c.seqToList(rv)
	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, rv.Len())
			reflect.Copy(reflect.ValueOf(b), rv)
			return Raw(b, false), nil
		}
		return c.seqToList(rv)
	case reflect.Map:
		return c.mapToValue(rv)
	case reflect.Struct:
		return c.structToMap(rv, "")
	default:
		return nil, xerrors.Errorf("value: cannot marshal kind %s", rv.Kind())
	}
}

func (c ValueCodec) seqToList(rv reflect.Value) (Value, error) {
	items := make([]Value, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		v, err := c.toValue(rv.Index(i))
		if err != nil {
			return nil, err
		}
		items[i] = v
	}
	return ImmutableList(items), nil
}

func (c ValueCodec) mapToValue(rv reflect.Value) (Value, error) {
	if rv.Type().Key().Kind() != reflect.String {
		return nil, xerrors.Errorf("value: map key must be string, got %s", rv.Type().Key().Kind())
	}
	m := make(map[string]Value, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		v, err := c.toValue(iter.Value())
		if err != nil {
			return nil, err
		}
		m[iter.Key().String()] = v
	}
	return ImmutableMapOf(m), nil
}

// structToMap projects a struct into a Value map. signSelector == "" is marshal
// mode (the full wire map, honoring omitempty); a non-empty signSelector is the
// signing projection — only fields selected by sign or legacy value metadata
// are included, with omitempty ignored so the signed shape is fixed.
func (c ValueCodec) structToMap(rv reflect.Value, signSelector string) (Value, error) {
	rt := rv.Type()
	m := make(map[string]Value)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		name, opts, skip, err := c.parseStructTag(f)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}
		if signSelector != "" {
			// Signing projection: include a field iff its dedicated sign tag or
			// legacy value-tag options carry the requested selector.
			if !c.hasSignSelector(f, opts, signSelector) {
				continue
			}
		} else if _, omit := opts["omitempty"]; omit && isEmptyValue(rv.Field(i)) {
			// Optional field at its zero value: omit it from the wire map.
			continue
		}
		v, err := c.toValue(rv.Field(i))
		if err != nil {
			return nil, xerrors.Errorf("value: field %q: %w", f.Name, err)
		}
		m[name] = v
	}
	return ImmutableMapOf(m), nil
}

// isEmptyValue reports whether rv holds its type's zero/empty value, matching
// encoding/json's omitempty semantics (so value and json/msgpack agree on which
// fields are dropped). Structs are never "empty"; use a pointer for an optional
// struct/time field.
func isEmptyValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

func signProjection(obj interface{}, selector string) (Value, error) {
	return DefaultValueCodec().signProjection(obj, selector)
}

func (c ValueCodec) signProjection(obj interface{}, selector string) (Value, error) {
	if selector == "" {
		return nil, xerrors.Errorf("value: SignBytes/SignHash require a non-empty signature marker")
	}
	rv := reflect.ValueOf(obj)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, xerrors.Errorf("value: SignBytes/SignHash got a nil pointer")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, xerrors.Errorf("value: SignBytes/SignHash require a struct, got %s", rv.Kind())
	}
	return c.structToMap(rv, selector)
}

func (c ValueCodec) fromValue(v Value, rv reflect.Value) error {
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
		return c.fromValue(v, rv.Elem())
	}
	if v == nil || v.Kind() == NULL {
		return nil // leave the zero value
	}
	if rv.Type() == timeType {
		if n, ok := v.(Number); ok {
			rv.Set(reflect.ValueOf(time.UnixMilli(n.Long()).UTC()))
			return nil
		}
		return xerrors.Errorf("value: cannot unmarshal %s into time.Time", v.Kind())
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
				if err := c.fromValue(item, out.Index(i)); err != nil {
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
			return c.mapFromValue(m, rv)
		}
	case reflect.Struct:
		if m, ok := v.(Map); ok {
			return c.mapToStruct(m, rv)
		}
	}
	return xerrors.Errorf("value: cannot unmarshal %s into %s", v.Kind(), rv.Type())
}

func (c ValueCodec) mapFromValue(m Map, rv reflect.Value) error {
	if rv.Type().Key().Kind() != reflect.String {
		return xerrors.Errorf("value: map key must be string, got %s", rv.Type().Key().Kind())
	}
	out := reflect.MakeMapWithSize(rv.Type(), m.Len())
	elemType := rv.Type().Elem()
	for _, e := range m.Entries() {
		ev := reflect.New(elemType).Elem()
		if err := c.fromValue(e.Value(), ev); err != nil {
			return err
		}
		out.SetMapIndex(reflect.ValueOf(e.Key()), ev)
	}
	rv.Set(out)
	return nil
}

func (c ValueCodec) mapToStruct(m Map, rv reflect.Value) error {
	// Get returns Null for both absent and present-Null keys, so use the presence
	// map to distinguish them (required fields must actually be present).
	hm := m.HashMap()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		name, opts, skip, err := c.parseStructTag(f)
		if err != nil {
			return err
		}
		if skip {
			continue
		}
		fv, present := hm[name]
		if !present {
			// Absent key: an error for required fields (no omitempty), tolerated
			// for optional ones. This enforces the wire schema in the library.
			if _, optional := opts["omitempty"]; !optional && !c.allowsMissingFields() {
				return xerrors.Errorf("value: missing required field %q", name)
			}
			continue
		}
		if fv == nil || fv.Kind() == NULL {
			continue // present but null leaves the zero value
		}
		if err := c.fromValue(fv, rv.Field(i)); err != nil {
			return xerrors.Errorf("value: field %q: %w", f.Name, err)
		}
	}
	return nil
}
