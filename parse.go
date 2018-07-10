// Copyright (c) 2018 James Bowen

package cereal

import (
	"fmt"
	"reflect"
	"strconv"
)

// Cereal data format can contain references to previously defined objects.
// To allow the parser to handle thes refs properly, we track all the
// objects in a global slice.
var objs []parser

// parser is a type which can be parsed from cereal data format and also
// unmershaled into a user provided value.
type parser interface {
	parse(*buffer) error
	unmarshal(reflect.Value) error
}

// parseElem reads the next data element from the cereal format buffer and
// returns in appropriate golang native type.
func parseElem(buf *buffer) (interface{}, error) {

	code, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch code {
	case 'b': // boolean
		val, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		switch val {
		case '1':
			return true, nil
		case '0':
			return false, nil
		default:
			return nil, fmt.Errorf("unknown boolean value '%c'", val)
		}

	case 'i', 'l': // integer / long
		return buf.readLineInt()

	case 'f': // float
		str, err := buf.readLineStr()
		if err != nil {
			return nil, err
		}

		return strconv.ParseFloat(str, 64)

	case 'r': // reference
		idx, err := buf.readLineInt()
		if err != nil {
			return nil, err
		}

		if idx < 0 || idx >= len(objs) {
			return nil, fmt.Errorf("invalid reference index %d", idx)
		}

		return objs[idx], nil

	case 's', 'u': // string
		len, err := buf.readLineInt()
		if err != nil {
			return nil, fmt.Errorf("error reading string length: %s", err)
		}

		val := make([]byte, len)
		_, err = buf.Read(val)
		if err != nil {
			return nil, fmt.Errorf("error reading string value: %s", err)
		}

		return string(val), nil

	case 'n': // None
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown type code '%c'", code)
	}

}

// parse decodes the cereal format data into single root object, perhaps
// containing additional intermediate objects (eg. Obj, List, Dict, etc..)
// and/or built-in native types (eg. int, float64, string).
func parse(data []byte) (interface{}, error) {

	buf := newBuffer(data)

	// First line should be 'cereal1'
	line, err := buf.readLineStr()
	if err != nil {
		return nil, err
	}

	if line != "cereal1" {
		return nil, fmt.Errorf("Invalid format: magic string '%s' != 'cereal1'", line)
	}

	// Second line is number of objects
	count, err := buf.readLineInt()
	if err != nil {
		return nil, err
	}

	objs = make([]parser, count)

	// Count the number of tuples as we have to read their data
	// in this section and ensure we do not attempt to read again later
	var numTuples int

	for i := 0; i < count; i++ {
		objType, err := buf.readLineStr()
		if err != nil {
			return nil, fmt.Errorf("error reading type for object %d: %s", i, err)
		}

		switch objType {
		case "dict":
			objs[i] = newDict()

		case "list", "set":
			objs[i] = newList()

		case "tuple":
			objs[i] = newTuple()
			numTuples++

			err = objs[i].parse(buf)
			if err != nil {
				return nil, fmt.Errorf("error parsing tuple data (object #%d): %s", i, err)
			}

		default:
			objs[i] = newObj(objType)
		}
	}

	// Read in data definitions for all non-tuple objects
	for i := 0; i < count-numTuples; i++ {
		err := objs[i].parse(buf)
		if err != nil {
			return nil, fmt.Errorf("error parsing data for object #%d: %s", i, err)
		}
	}

	// Now read reference for root object
	return parseElem(buf)
}

// Unmarshal parses the cereal-encoded data and stores the result in
// the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {

	// v must be a non-nil pointer or we cannot set anything
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("invalid interface type: %s", reflect.TypeOf(v))
	}

	// Parse the cereal data into intermediate tree of objects
	root, err := parse(data)
	if err != nil {
		return err
	}

	val, err := unmarshalType(rv.Elem().Type(), root)
	if err != nil {
		return err
	}

	rv.Elem().Set(val)
	return nil
}

// unmarshalType will attempt to convert value v into type rt, returning a reflect.Value
// set accordingly. It will return an error if conversion is not possible.
func unmarshalType(rt reflect.Type, v interface{}) (reflect.Value, error) {

	var newVal reflect.Value

	// Dict, Obj, List, etc..
	if p, ok := v.(parser); ok {
		newVal = reflect.New(rt).Elem()
		err := p.unmarshal(newVal)

		return newVal, err

	}

	// Handle built in types

	switch rt.Kind() {

	// Integer
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:

		// All integers parsed from cereal are ints
		intVal, ok := v.(int)
		if !ok {
			return newVal, fmt.Errorf("illegal type '%s' for destination %s (expect 'int')", reflect.TypeOf(v), rt)
		}

		// New() returns pointer, Elem() converts to underlying value
		newVal = reflect.New(rt).Elem()

		// Check for overflow
		if newVal.OverflowInt(int64(intVal)) {
			return newVal, fmt.Errorf("integer overflow: value %d is not valid %s", v, rt)
		}

		newVal.SetInt(int64(intVal))

	// String
	case reflect.String:

		strVal, ok := v.(string)
		if !ok {
			return newVal, fmt.Errorf("illegal type '%s' for destination %s", reflect.TypeOf(v), rt)
		}

		newVal = reflect.New(rt).Elem()
		newVal.SetString(strVal)

	// Float
	case reflect.Float32, reflect.Float64:

		var floatVal float64

		switch v.(type) {
		case float32:
			floatVal = float64(v.(float32))
		case float64:
			floatVal = v.(float64)
		default:
			return newVal, fmt.Errorf("illegal type '%s' for destination %s", reflect.TypeOf(v), rt)
		}

		newVal = reflect.New(rt).Elem()
		newVal.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, ok := v.(bool)
		if !ok {
			return newVal, fmt.Errorf("illegal type '%s' for destination %s", reflect.TypeOf(v), rt)
		}

		newVal = reflect.New(rt).Elem()
		newVal.SetBool(boolVal)

	// interface{}
	case reflect.Interface:
		// TODO
		fallthrough
	default:
		return newVal, fmt.Errorf("unsupported element type %s", rt)
	}

	return newVal, nil

}
