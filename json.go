/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: Apache-2.0
 */

package value

import (
	"encoding/base64"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/shopspring/decimal"
	"golang.org/x/xerrors"
)

/*
Lossless JSON (CANONICAL.md, "JSON").

Plain JSON has numbers, strings, booleans, null, arrays and objects; a value has
more kinds. AppendJSON writes each kind JSON lacks as an object with a single
wrapper key, and ParseJSON reads it back:

	{"$long": "5"}                            a 64-bit integer
	{"$double": "1.5"}                        a double; also "NaN", "Infinity", "-Infinity"
	{"$bigint": "123"}                        a big integer, in decimal
	{"$decimal": "1.10"}                      a decimal, with its exact digits
	{"$bin": "AQI="}                          binary, in standard base64
	{"$timestamp": "2026-10-30T07:28:39.075Z"} an instant, RFC 3339 in UTC
	{"$sparse": {"0": "a", "5": "b"}}         a sparse list, by index
	{"$ext": {"base64": "AQ==", "type": 7}}   an extension value does not know
	{"$map": {"$bin": "x"}}                   a map whose only key is a wrapper key

JSONCanonical wraps every one of them, so ParseJSON gives back the value
AppendJSON wrote. JSONRelaxed is for people and for plain JSON tools: it writes
longs, doubles and decimals as JSON numbers with their exact digits and wraps
the rest. ParseJSON reads a bare JSON number, which carries no type, as a
Decimal built from its exact text, integers included: 1 is Decimal 1 and 1.10
keeps both digits.

Jsonify and PrintJSON keep their earlier rendering, which value-rpc's JSON wire
codec speaks.
*/

// JSONMode selects how AppendJSON writes the kinds plain JSON lacks.
type JSONMode int

const (
	// JSONRelaxed writes longs, doubles and decimals as JSON numbers with their
	// exact digits and wraps binary, timestamps, big integers, sparse lists and
	// unknown extensions. ParseJSON reads its numbers back as Decimal.
	JSONRelaxed JSONMode = iota

	// JSONCanonical wraps every kind plain JSON lacks, numbers included, so
	// ParseJSON reads back the same value.
	JSONCanonical
)

// The wrapper keys of lossless JSON.
const (
	JSONLong      = "$long"
	JSONDouble    = "$double"
	JSONBigInt    = "$bigint"
	JSONDecimal   = "$decimal"
	JSONBinary    = "$bin"
	JSONTimestamp = "$timestamp"
	JSONSparse    = "$sparse"
	JSONExt       = "$ext"
	JSONMap       = "$map"
)

func isJSONWrapper(key string) bool {
	switch key {
	case JSONLong, JSONDouble, JSONBigInt, JSONDecimal, JSONBinary, JSONTimestamp, JSONSparse, JSONExt, JSONMap:
		return true
	}
	return false
}

// FormatJSON returns the JSON text of a value in the mode.
func FormatJSON(val Value, mode JSONMode) ([]byte, error) {
	return AppendJSON(nil, val, mode)
}

// AppendJSON appends the JSON text of a value in the mode to dst. A string
// that is not valid UTF-8 has no JSON form and fails.
func AppendJSON(dst []byte, val Value, mode JSONMode) ([]byte, error) {
	return appendJSON(dst, val, mode, 0)
}

func appendJSON(dst []byte, val Value, mode JSONMode, depth int) ([]byte, error) {
	if err := checkDepth(depth); err != nil {
		return nil, err
	}
	if val == nil {
		return append(dst, "null"...), nil
	}
	switch val.Kind() {
	case NULL:
		return append(dst, "null"...), nil
	case BOOL:
		return strconv.AppendBool(dst, val.(Bool).Boolean()), nil
	case NUMBER:
		return appendJSONNumber(dst, val.(Number), mode)
	case STRING:
		s := val.(String)
		if s.Type() == RAW {
			return appendJSONWrapper(dst, JSONBinary, base64.StdEncoding.EncodeToString(s.Raw())), nil
		}
		return appendJSONString(dst, s.Utf8())
	case TIMESTAMP:
		return appendJSONWrapper(dst, JSONTimestamp, val.(Timestamp).Time().Format(timestampJSONLayout)), nil
	case LIST:
		if sparse, ok := val.(sparseListValue); ok {
			return appendJSONSparse(dst, sparse, mode, depth)
		}
		dst = append(dst, '[')
		for i, el := range val.(List).Values() {
			if i > 0 {
				dst = append(dst, ',')
			}
			var err error
			if dst, err = appendJSON(dst, el, mode, depth+1); err != nil {
				return nil, err
			}
		}
		return append(dst, ']'), nil
	case MAP:
		entries := val.(Map).Entries()
		if !sort.SliceIsSorted(entries, func(i, j int) bool { return entries[i].Key() < entries[j].Key() }) {
			entries = append([]MapEntry(nil), entries...)
			sort.Slice(entries, func(i, j int) bool { return entries[i].Key() < entries[j].Key() })
		}
		if len(entries) == 1 && isJSONWrapper(entries[0].Key()) {
			// The map would read back as the wrapped kind: escape it.
			dst = append(dst, `{"`+JSONMap+`":`...)
			dst, err := appendJSONEntries(dst, entries, mode, depth+1)
			if err != nil {
				return nil, err
			}
			return append(dst, '}'), nil
		}
		return appendJSONEntries(dst, entries, mode, depth)
	case UNKNOWN:
		x, ok := val.(unknownValue)
		if !ok || len(x) == 0 {
			return nil, xerrors.Errorf("value: extension %T has no JSON form", val)
		}
		dst = append(dst, `{"`+JSONExt+`":{"base64":`...)
		dst = strconv.AppendQuote(dst, base64.StdEncoding.EncodeToString(x.Data()))
		dst = append(dst, `,"type":`...)
		dst = strconv.AppendInt(dst, int64(int8(x.Tag())), 10)
		return append(dst, "}}"...), nil
	}
	return nil, xerrors.Errorf("value: kind %s has no JSON form", val.Kind())
}

// timestampJSONLayout is RFC 3339 with the fraction the instant needs.
const timestampJSONLayout = "2006-01-02T15:04:05.999999999Z07:00"

func appendJSONEntries(dst []byte, entries []MapEntry, mode JSONMode, depth int) ([]byte, error) {
	dst = append(dst, '{')
	for i, e := range entries {
		if i > 0 {
			dst = append(dst, ',')
		}
		var err error
		if dst, err = appendJSONString(dst, e.Key()); err != nil {
			return nil, err
		}
		dst = append(dst, ':')
		if dst, err = appendJSON(dst, e.Value(), mode, depth+1); err != nil {
			return nil, err
		}
	}
	return append(dst, '}'), nil
}

func appendJSONSparse(dst []byte, items sparseListValue, mode JSONMode, depth int) ([]byte, error) {
	sorted := []ListItem(items)
	if !sort.SliceIsSorted(sorted, func(i, j int) bool { return sorted[i].Key() < sorted[j].Key() }) {
		sorted = append([]ListItem(nil), sorted...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Key() < sorted[j].Key() })
	}
	dst = append(dst, `{"`+JSONSparse+`":{`...)
	for i, item := range sorted {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = append(dst, '"')
		dst = strconv.AppendInt(dst, int64(item.Key()), 10)
		dst = append(dst, `":`...)
		var err error
		if dst, err = appendJSON(dst, item.Value(), mode, depth+1); err != nil {
			return nil, err
		}
	}
	return append(dst, "}}"...), nil
}

func appendJSONNumber(dst []byte, n Number, mode JSONMode) ([]byte, error) {
	switch n.Type() {
	case LONG:
		if mode == JSONCanonical {
			return appendJSONWrapper(dst, JSONLong, strconv.FormatInt(n.Long(), 10)), nil
		}
		return strconv.AppendInt(dst, n.Long(), 10), nil
	case DOUBLE:
		d := n.Double()
		if mode == JSONCanonical || math.IsNaN(d) || math.IsInf(d, 0) {
			return appendJSONWrapper(dst, JSONDouble, DoubleText(d)), nil
		}
		return append(dst, DoubleText(d)...), nil
	case BIGINT:
		return appendJSONWrapper(dst, JSONBigInt, n.BigInt().String()), nil
	case DECIMAL:
		text := DecimalText(n.Decimal())
		if mode == JSONCanonical {
			return appendJSONWrapper(dst, JSONDecimal, text), nil
		}
		return append(dst, text...), nil
	}
	return nil, xerrors.Errorf("value: number type %s has no JSON form", n.Type())
}

func appendJSONWrapper(dst []byte, key, text string) []byte {
	dst = append(dst, `{"`...)
	dst = append(dst, key...)
	dst = append(dst, `":`...)
	dst = strconv.AppendQuote(dst, text)
	return append(dst, '}')
}

const hexDigits = "0123456789abcdef"

// appendJSONString writes a JSON string, escaping only what JSON requires.
func appendJSONString(dst []byte, s string) ([]byte, error) {
	if !utf8.ValidString(s) {
		return nil, xerrors.Errorf("value: a string that is not valid UTF-8 has no JSON form")
	}
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '"', '\\':
			dst = append(dst, '\\', c)
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		case '\b':
			dst = append(dst, '\\', 'b')
		case '\f':
			dst = append(dst, '\\', 'f')
		default:
			if c < 0x20 {
				dst = append(dst, '\\', 'u', '0', '0', hexDigits[c>>4], hexDigits[c&0xf])
			} else {
				dst = append(dst, c)
			}
		}
	}
	return append(dst, '"'), nil
}

// DoubleText is the shortest text that reads back as the double, in the form
// JavaScript's Number.prototype.toString writes: plain digits from 1e-7 to
// 1e21, an exponent outside. -0 keeps its sign; the non-finite doubles are
// "NaN", "Infinity" and "-Infinity".
func DoubleText(d float64) string {
	switch {
	case math.IsNaN(d):
		return "NaN"
	case math.IsInf(d, 1):
		return "Infinity"
	case math.IsInf(d, -1):
		return "-Infinity"
	case d == 0:
		if math.Signbit(d) {
			return "-0"
		}
		return "0"
	}
	e := strconv.FormatFloat(math.Abs(d), 'e', -1, 64) // d.ddde±x
	at := strings.IndexByte(e, 'e')
	digits := strings.Replace(e[:at], ".", "", 1)
	exp, _ := strconv.Atoi(e[at+1:])
	k, n := len(digits), exp+1
	var b strings.Builder
	if d < 0 {
		b.WriteByte('-')
	}
	switch {
	case k <= n && n <= 21:
		b.WriteString(digits)
		b.WriteString(strings.Repeat("0", n-k))
	case 0 < n && n <= 21:
		b.WriteString(digits[:n])
		b.WriteByte('.')
		b.WriteString(digits[n:])
	case -6 < n && n <= 0:
		b.WriteString("0.")
		b.WriteString(strings.Repeat("0", -n))
		b.WriteString(digits)
	default:
		b.WriteByte(digits[0])
		if k > 1 {
			b.WriteByte('.')
			b.WriteString(digits[1:])
		}
		b.WriteByte('e')
		if n-1 >= 0 {
			b.WriteByte('+')
		}
		b.WriteString(strconv.Itoa(n - 1))
	}
	return b.String()
}

// DecimalText is the exact text of a decimal: its coefficient's digits and its
// exponent, so that 1.10 keeps its trailing zero, 1e3 its exponent and 1e-400
// does not spell out its zeros. ParseJSON and the $decimal wrapper read it
// back to the same coefficient and exponent.
func DecimalText(d decimal.Decimal) string {
	c, e := d.Coefficient(), int(d.Exponent())
	digits := new(big.Int).Abs(c).String()
	var b strings.Builder
	if c.Sign() < 0 {
		b.WriteByte('-')
	}
	switch scale := -e; {
	case e == 0:
		b.WriteString(digits)
	case e > 0 || scale-len(digits) > 6:
		b.WriteString(digits)
		b.WriteByte('e')
		b.WriteString(strconv.Itoa(e))
	case scale < len(digits):
		b.WriteString(digits[:len(digits)-scale])
		b.WriteByte('.')
		b.WriteString(digits[len(digits)-scale:])
	default:
		b.WriteString("0.")
		b.WriteString(strings.Repeat("0", scale-len(digits)))
		b.WriteString(digits)
	}
	return b.String()
}

// ParseJSON reads a value from JSON text written in either mode. A bare number
// is a Decimal from its exact text; an object whose only key is a wrapper key
// is the wrapped kind, and a malformed wrapper fails rather than reading as a
// map. Duplicate keys, strings that are not valid UTF-8, lone surrogates and
// text after the value fail too. The parse limits apply, and a number may have
// at most MaxDecimalExponent characters.
func ParseJSON(data []byte) (Value, error) {
	p := jsonParser{data: data}
	p.space()
	n, err := p.node(0)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.pos != len(p.data) {
		return nil, p.errorf("text after the value")
	}
	return jsonBuild(n, false)
}

// The parser reads JSON into nodes first: a leaf is its Value, an array a
// jsonArray and an object a jsonObject. Only when an object is complete is it
// known whether it is a wrapper, and whether the object inside a $map escape
// is a plain map, so wrappers are read when the nodes are built.
type jsonArray []any

type jsonObject struct {
	at      int
	entries []jsonEntry
}

type jsonEntry struct {
	key  string
	node any
}

type jsonParser struct {
	data []byte
	pos  int
}

func (p *jsonParser) errorf(format string, args ...any) error {
	return jsonErrorf(p.pos, format, args...)
}

func jsonErrorf(at int, format string, args ...any) error {
	return xerrors.Errorf("value: json at offset %d: "+format, append([]any{at}, args...)...)
}

func (p *jsonParser) space() {
	for p.pos < len(p.data) {
		switch p.data[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *jsonParser) next(c byte) bool {
	if p.pos < len(p.data) && p.data[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func (p *jsonParser) node(depth int) (any, error) {
	if err := checkDepth(depth); err != nil {
		return nil, err
	}
	if p.pos >= len(p.data) {
		return nil, p.errorf("unexpected end")
	}
	switch c := p.data[p.pos]; {
	case c == '{':
		return p.object(depth)
	case c == '[':
		return p.array(depth)
	case c == '"':
		s, err := p.str()
		if err != nil {
			return nil, err
		}
		return Utf8(s), nil
	case c == 't':
		return p.literal("true", Boolean(true))
	case c == 'f':
		return p.literal("false", Boolean(false))
	case c == 'n':
		return p.literal("null", Null)
	case c == '-' || (c >= '0' && c <= '9'):
		at := p.pos
		text, ok := scanJSONNumber(p.data[p.pos:])
		if !ok {
			return nil, p.errorf("malformed number")
		}
		p.pos += len(text)
		d, err := parseDecimalText(string(text))
		if err != nil {
			return nil, jsonErrorf(at, "%v", err)
		}
		return Decimal(d), nil
	default:
		return nil, p.errorf("unexpected %q", c)
	}
}

func (p *jsonParser) literal(word string, v Value) (any, error) {
	if !strings.HasPrefix(string(p.data[p.pos:min(len(p.data), p.pos+len(word))]), word) {
		return nil, p.errorf("unexpected text")
	}
	p.pos += len(word)
	return v, nil
}

func (p *jsonParser) object(depth int) (any, error) {
	obj := jsonObject{at: p.pos}
	p.pos++
	p.space()
	if p.next('}') {
		return obj, nil
	}
	for {
		p.space()
		if p.pos >= len(p.data) || p.data[p.pos] != '"' {
			return nil, p.errorf("expected a key")
		}
		key, err := p.str()
		if err != nil {
			return nil, err
		}
		p.space()
		if !p.next(':') {
			return nil, p.errorf("expected ':'")
		}
		p.space()
		n, err := p.node(depth + 1)
		if err != nil {
			return nil, err
		}
		obj.entries = append(obj.entries, jsonEntry{key: key, node: n})
		if err := checkCollectionLen(len(obj.entries)); err != nil {
			return nil, err
		}
		p.space()
		if p.next(',') {
			continue
		}
		if p.next('}') {
			return obj, nil
		}
		return nil, p.errorf("expected ',' or '}'")
	}
}

func (p *jsonParser) array(depth int) (any, error) {
	p.pos++
	p.space()
	var list jsonArray
	if p.next(']') {
		return list, nil
	}
	for {
		p.space()
		n, err := p.node(depth + 1)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
		if err := checkCollectionLen(len(list)); err != nil {
			return nil, err
		}
		p.space()
		if p.next(',') {
			continue
		}
		if p.next(']') {
			return list, nil
		}
		return nil, p.errorf("expected ',' or ']'")
	}
}

// jsonBuild turns a node into its value. A literal object is the body of a
// $map escape: a map even when its only key is a wrapper key.
func jsonBuild(n any, literal bool) (Value, error) {
	switch x := n.(type) {
	case Value:
		return x, nil
	case jsonArray:
		if len(x) == 0 {
			return EmptyImmutableList(), nil
		}
		list := make([]Value, len(x))
		for i, el := range x {
			v, err := jsonBuild(el, false)
			if err != nil {
				return nil, err
			}
			list[i] = v
		}
		return ImmutableList(list), nil
	case jsonObject:
		if !literal && len(x.entries) == 1 && isJSONWrapper(x.entries[0].key) {
			key, inner := x.entries[0].key, x.entries[0].node
			if key == JSONMap {
				body, ok := inner.(jsonObject)
				if !ok {
					return nil, jsonErrorf(x.at, "%s wants an object", key)
				}
				return jsonBuild(body, true)
			}
			v, err := jsonBuild(inner, false)
			if err != nil {
				return nil, err
			}
			if v, err = jsonWrapped(key, v); err != nil {
				return nil, jsonErrorf(x.at, "%v", err)
			}
			return v, nil
		}
		if len(x.entries) == 0 {
			return EmptyImmutableMap(), nil
		}
		entries := make([]MapEntry, len(x.entries))
		for i, e := range x.entries {
			v, err := jsonBuild(e.node, false)
			if err != nil {
				return nil, err
			}
			entries[i] = ImmutableEntry(e.key, v)
		}
		sort.SliceStable(entries, func(i, j int) bool { return entries[i].Key() < entries[j].Key() })
		for i := 1; i < len(entries); i++ {
			if entries[i-1].Key() == entries[i].Key() {
				return nil, jsonErrorf(x.at, "duplicate key %q", entries[i].Key())
			}
		}
		return ImmutableMap(entries, true), nil
	}
	return nil, xerrors.Errorf("value: json node %T", n)
}

// str reads a string: escapes decoded, surrogate pairs joined, lone
// surrogates, raw control characters and invalid UTF-8 refused.
func (p *jsonParser) str() (string, error) {
	p.pos++
	var b []byte
	from := p.pos
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		switch {
		case c == '"':
			b = append(b, p.data[from:p.pos]...)
			p.pos++
			if !utf8.Valid(b) {
				return "", p.errorf("string is not valid UTF-8")
			}
			if err := checkByteLen(len(b)); err != nil {
				return "", err
			}
			return string(b), nil
		case c == '\\':
			b = append(b, p.data[from:p.pos]...)
			r, err := p.escape()
			if err != nil {
				return "", err
			}
			b = utf8.AppendRune(b, r)
			from = p.pos
		case c < 0x20:
			return "", p.errorf("control character in a string")
		default:
			p.pos++
		}
	}
	return "", p.errorf("unterminated string")
}

func (p *jsonParser) escape() (rune, error) {
	p.pos++
	if p.pos >= len(p.data) {
		return 0, p.errorf("unterminated escape")
	}
	c := p.data[p.pos]
	p.pos++
	switch c {
	case '"', '\\', '/':
		return rune(c), nil
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'u':
		r, err := p.hex4()
		if err != nil {
			return 0, err
		}
		if utf16.IsSurrogate(r) {
			if r >= 0xdc00 || !p.next('\\') || !p.next('u') {
				return 0, p.errorf("lone surrogate")
			}
			r2, err := p.hex4()
			if err != nil {
				return 0, err
			}
			if r = utf16.DecodeRune(r, r2); r == utf8.RuneError {
				return 0, p.errorf("lone surrogate")
			}
		}
		return r, nil
	}
	return 0, p.errorf("unknown escape %q", c)
}

func (p *jsonParser) hex4() (rune, error) {
	if p.pos+4 > len(p.data) {
		return 0, p.errorf("short \\u escape")
	}
	var r rune
	for _, c := range p.data[p.pos : p.pos+4] {
		r <<= 4
		switch {
		case c >= '0' && c <= '9':
			r |= rune(c - '0')
		case c >= 'a' && c <= 'f':
			r |= rune(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			r |= rune(c - 'A' + 10)
		default:
			return 0, p.errorf("malformed \\u escape")
		}
	}
	p.pos += 4
	return r, nil
}

// scanJSONNumber returns the JSON number b starts with, by RFC 8259's grammar.
func scanJSONNumber(b []byte) ([]byte, bool) {
	i := 0
	digits := func() int {
		n := 0
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
			n++
		}
		return n
	}
	if i < len(b) && b[i] == '-' {
		i++
	}
	switch {
	case i < len(b) && b[i] == '0':
		i++
	case digits() == 0:
		return nil, false
	}
	if i < len(b) && b[i] == '.' {
		i++
		if digits() == 0 {
			return nil, false
		}
	}
	if i < len(b) && (b[i] == 'e' || b[i] == 'E') {
		i++
		if i < len(b) && (b[i] == '+' || b[i] == '-') {
			i++
		}
		if digits() == 0 {
			return nil, false
		}
	}
	return b[:i], true
}

// isJSONNumber reports whether s is one whole JSON number.
func isJSONNumber(s string) bool {
	n, ok := scanJSONNumber([]byte(s))
	return ok && len(n) == len(s)
}

func parseDecimalText(text string) (decimal.Decimal, error) {
	if MaxDecimalExponent > 0 && len(text) > MaxDecimalExponent {
		return decimal.Decimal{}, xerrors.Errorf("number of %d characters exceeds limit %d", len(text), MaxDecimalExponent)
	}
	d, err := decimal.NewFromString(text)
	if err != nil {
		return decimal.Decimal{}, xerrors.Errorf("number %q: %w", text, err)
	}
	if err := checkDecimalExponent(int64(d.Exponent())); err != nil {
		return decimal.Decimal{}, err
	}
	return d, nil
}

// jsonWrapped reads the value of a wrapper key.
func jsonWrapped(key string, v Value) (Value, error) {
	text := func() (string, error) {
		s, ok := v.(String)
		if !ok || s.Type() != UTF8 {
			return "", xerrors.Errorf("%s holds %s, want a string", key, v.Kind())
		}
		return s.Utf8(), nil
	}
	switch key {
	case JSONLong:
		s, err := text()
		if err != nil {
			return nil, err
		}
		if !isJSONNumber(s) || strings.ContainsAny(s, ".eE") {
			return nil, xerrors.Errorf("%s %q is not an integer", key, s)
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, xerrors.Errorf("%s %q: %w", key, s, err)
		}
		return Long(n), nil
	case JSONDouble:
		s, err := text()
		if err != nil {
			return nil, err
		}
		switch s {
		case "NaN":
			return Double(math.NaN()), nil
		case "Infinity":
			return Double(math.Inf(1)), nil
		case "-Infinity":
			return Double(math.Inf(-1)), nil
		}
		if !isJSONNumber(s) {
			return nil, xerrors.Errorf("%s %q is not a number", key, s)
		}
		d, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, xerrors.Errorf("%s %q: %w", key, s, err)
		}
		return Double(d), nil
	case JSONBigInt:
		s, err := text()
		if err != nil {
			return nil, err
		}
		if !isJSONNumber(s) || strings.ContainsAny(s, ".eE") {
			return nil, xerrors.Errorf("%s %q is not an integer", key, s)
		}
		if MaxDecimalExponent > 0 && len(s) > MaxDecimalExponent {
			return nil, xerrors.Errorf("%s of %d digits exceeds limit %d", key, len(s), MaxDecimalExponent)
		}
		n, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, xerrors.Errorf("%s %q is not an integer", key, s)
		}
		return BigInt(n), nil
	case JSONDecimal:
		s, err := text()
		if err != nil {
			return nil, err
		}
		if !isJSONNumber(s) {
			return nil, xerrors.Errorf("%s %q is not a number", key, s)
		}
		d, err := parseDecimalText(s)
		if err != nil {
			return nil, err
		}
		return Decimal(d), nil
	case JSONBinary:
		s, err := text()
		if err != nil {
			return nil, err
		}
		b, err := decodeJSONBase64(s)
		if err != nil {
			return nil, xerrors.Errorf("%s: %w", key, err)
		}
		return Raw(b, false), nil
	case JSONTimestamp:
		s, err := text()
		if err != nil {
			return nil, err
		}
		return ParseTimestamp(s)
	case JSONSparse:
		m, ok := v.(Map)
		if !ok {
			return nil, xerrors.Errorf("%s holds %s, want an object", key, v.Kind())
		}
		items := make([]ListItem, 0, m.Len())
		for _, e := range m.Entries() {
			k := e.Key()
			if k == "" || len(k) > 10 || (k[0] == '0' && len(k) > 1) || strings.Trim(k, "0123456789") != "" {
				return nil, xerrors.Errorf("%s index %q is not a non-negative integer", key, k)
			}
			i, err := strconv.Atoi(k)
			if err != nil || (MaxParseCollectionLen > 0 && i > MaxParseCollectionLen) {
				return nil, xerrors.Errorf("%s index %q out of range", key, k)
			}
			items = append(items, ImmutableItem(i, e.Value()))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Key() < items[j].Key() })
		return SparseList(items, true), nil
	case JSONExt:
		m, ok := v.(Map)
		if !ok || m.Len() != 2 {
			return nil, xerrors.Errorf("%s wants an object with base64 and type", key)
		}
		s, ok := m.Get("base64").(String)
		n, ok2 := m.Get("type").(Number)
		if !ok || !ok2 || s.Type() != UTF8 {
			return nil, xerrors.Errorf("%s wants an object with base64 and type", key)
		}
		t := n.Decimal()
		if !t.IsInteger() || t.LessThan(decimal.NewFromInt(-128)) || t.GreaterThan(decimal.NewFromInt(127)) {
			return nil, xerrors.Errorf("%s type %s is not an extension type", key, t)
		}
		data, err := decodeJSONBase64(s.Utf8())
		if err != nil {
			return nil, xerrors.Errorf("%s: %w", key, err)
		}
		return doParseExt(append([]byte{byte(int8(t.IntPart()))}, data...))
	}
	return nil, xerrors.Errorf("unknown wrapper %s", key)
}

// decodeJSONBase64 reads standard base64, padded or not.
func decodeJSONBase64(s string) ([]byte, error) {
	if err := checkByteLen(base64.RawStdEncoding.DecodedLen(len(strings.TrimRight(s, "=")))); err != nil {
		return nil, err
	}
	if strings.HasSuffix(s, "=") || len(s)%4 == 0 {
		return base64.StdEncoding.DecodeString(s)
	}
	return base64.RawStdEncoding.DecodeString(s)
}
