package jsonobj

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	CamelCase = NameConv(iota)
	SnakeCase
	LowerCase
	UpperCase
)

var (
	ErrUnkNameConv = fmt.Errorf("unknown name convention")
)

// NameConv defines a custom type for identifying
// naming convention
type NameConv int

// Options to apply modification on JSONObject fields
type Options struct {
	FieldNameConvention NameConv
}

// Field structure definition
type Field struct {
	Name  string
	Value any
}

// JSONObject structure definition
type JSONObject struct {
	fields  []Field
	cache   map[string]int
	Options *Options
}

func (o *JSONObject) newChild() (new *JSONObject) {
	new = New()
	new.Options = o.Options
	return
}

func (o *JSONObject) newFromValue(v reflect.Value) (new *JSONObject) {
	new = o.newChild()
	new.fromStruct(v)
	return
}

func isBaseType(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		return true
	default:
		return false
	}
}

func resolve(v reflect.Value) reflect.Value {
	var t reflect.Type

	// we deref type until we get something not Ptr
	for t = v.Type(); ; t = v.Type() {
		if t.Kind() != reflect.Ptr && t.Kind() != reflect.Interface {
			break
		}
		v = v.Elem()
	}
	return v
}

func (o *JSONObject) convertSlice(slice reflect.Value) (s []interface{}) {
	if slice.Kind() != reflect.Slice {
		panic("can convert only slice")
	}

	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)

		v = resolve(v)

		if isBaseType(v) {
			s = append(s, v.Interface())
			continue
		}

		switch v.Kind() {
		case reflect.Struct, reflect.Ptr:
			s = append(s, o.newFromValue(v))
		case reflect.Slice:
			s = append(s, o.convertSlice(v))
		case reflect.Map:
			s = append(s, o.newFromMapValue(v))
		default:
			panic("not handled")
		}
	}
	return
}

func (o *JSONObject) fromMapValue(m reflect.Value) {
	iter := m.MapRange()

	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		if k.Kind() != reflect.String {
			panic("only map with string key are supported")
		}

		fieldName := k.Interface().(string)

		v = resolve(v)

		if isBaseType(v) {
			o.SetField(fieldName, v.Interface())
			continue
		}

		switch v.Kind() {
		case reflect.Struct, reflect.Ptr:
			o.SetField(fieldName, o.newFromValue(v))
		case reflect.Slice:
			o.SetField(fieldName, o.convertSlice(v))
		case reflect.Map:
			o.SetField(fieldName, o.newFromMapValue(v))
		default:
			panic("not handled")
		}
	}
}

func (o *JSONObject) newFromMapValue(m reflect.Value) (new *JSONObject) {
	new = o.newChild()
	new.fromMapValue(m)
	return
}

func (o *JSONObject) fromStruct(v reflect.Value) {
	v = resolve(v)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		panic("not a structure")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		fieldName := fieldType.Name

		if !fieldType.IsExported() {
			continue
		}

		if isBaseType(field) {
			o.SetField(fieldName, field.Interface())
			continue
		}

		v = resolve(v)

		switch field.Kind() {
		case reflect.Struct, reflect.Ptr:
			// this is an embedded structure
			if fieldType.Anonymous {
				o.fromStruct(field)
			} else {
				o.SetField(fieldName, o.newFromValue(field))
			}
		case reflect.Slice:
			o.SetField(fieldName, o.convertSlice(field))
		case reflect.Map:
			o.SetField(fieldName, o.newFromMapValue(field))
		default:
			panic("not handled")
		}
	}
}

// New creates an empty new JSONObject
func New() *JSONObject {
	return &JSONObject{
		Options: &Options{},
		cache:   make(map[string]int),
	}
}

// FromMap creates a new JSONObject from a structure
func FromStruct(s any) (o *JSONObject) {
	o = New()
	return o.newFromValue(reflect.ValueOf(s))
}

// FromMap creates a new JSONObject with options from a structure
func FromStructWithOptions(s any, opt Options) (o *JSONObject) {
	o = New()
	o.Options = &opt
	o.fromStruct(reflect.ValueOf(s))
	return
}

// FromMap creates a new JSONObject from a map
func FromMap(m map[string]any) (o *JSONObject) {
	o = New()
	o.fromMapValue(reflect.ValueOf(m))
	return
}

// json marshaling implementation
func (o JSONObject) MarshalJSON() (out []byte, err error) {
	// object opening
	out = append(out, '{')

	for i, f := range o.fields {
		var name []byte
		var value []byte

		name, err = json.Marshal(f.Name)
		if err != nil {
			return
		}
		value, err = json.Marshal(f.Value)
		if err != nil {
			return
		}
		out = append(out, name...)
		out = append(out, ':')
		out = append(out, value...)
		if i != len(o.fields)-1 {
			out = append(out, ',')
		}
	}

	// object closing
	out = append(out, '}')
	return
}

// ConvertSlice converts any slice to a JSONObject ready slice.
// It is useful to recursively convert slice of structures, so that
// options are also applied to slice elements. On slices containing
// base types, this method has no effect.
func (o *JSONObject) ConvertSlice(slice any) (s []any) {
	return o.convertSlice(reflect.ValueOf(slice))
}

// SetField sets a field of the object. If the field is
// already existing, value is replaced
func (o *JSONObject) SetField(name string, value any) {

	if o.Options != nil {
		switch o.Options.FieldNameConvention {
		case SnakeCase:
			name = camelToSnake(name)
		case LowerCase:
			name = strings.ToLower(name)
		case UpperCase:
			name = strings.ToUpper(name)
		}
	}

	// we replace value if field alread there
	if i, ok := o.cache[name]; ok {
		o.fields[i].Value = value
		return
	}

	o.fields = append(o.fields, Field{name, value})
	o.cache[name] = len(o.fields)
}

// HasField returns true if field is present in JSONObject
func (o *JSONObject) HasField(name string) bool {
	_, ok := o.cache[name]
	return ok
}

// GetField returns field's value if any. It panics if field is not present
func (o *JSONObject) GetField(name string) any {
	if i, ok := o.cache[name]; ok {
		return o.fields[i].Value
	}
	panic("unknown field")
}

// IsEmpty checks wether the JSONObject is empty
func (o *JSONObject) IsEmpty() bool {
	return len(o.fields) == 0
}
