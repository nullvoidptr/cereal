// Copyright (c) 2018 James Bowen

package cereal

import (
	"fmt"
	"reflect"
)

// This files contains types representing python objects read for a cereal-format
// datastream. The standard method of parsing the cereal file is to first create
// intermediate instances of these types (and/or built-in types like int, string, etc..)
// and then perform an unmarshal operation into user provided target type.

// list is a representation of a python list
type list struct {
	vals []interface{}
}

// newList returns a new List instance
func newList() *list {
	return &list{}
}

// parse decodes a cereal format data defintion into the List instance
func (l *list) parse(b *buffer) error {

	// fmt.Println("Parsing *List data")

	count, err := b.readLineInt()
	if err != nil {
		return err
	}

	l.vals = make([]interface{}, count)

	for i := 0; i < count; i++ {
		val, err := parseElem(b)
		if err != nil {
			return fmt.Errorf("error parsing List value: %s", err)
		}

		l.vals[i] = val
	}

	return nil
}

// unmarshal takes values stored in List and copies to Value. Value must be
// setable and be a kind of slice.
func (l *list) unmarshal(rv reflect.Value) error {

	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal List into non-slice type %s", rv.Type())
	}

	elemType := rv.Type().Elem()

	for _, v := range l.vals {
		newVal, err := unmarshalType(elemType, v)
		if err != nil {
			return err
		}

		rv.Set(reflect.Append(rv, newVal))
	}
	return nil
}

// tuple is a representaion of a python tuple. It is functionally
// identical to List but stored differently by cerealizer.
type tuple struct {
	list
}

// newTuple returns a new Tuple instance
func newTuple() *tuple {
	return &tuple{}
}

// dict is a representation of a python dictionary
type dict struct {
	vals map[interface{}]interface{}
}

// newDict returns a new Dict instance
func newDict() *dict {
	var d dict

	d.vals = make(map[interface{}]interface{})
	return &d
}

// parse decodes a cereal format data defintion into the Dict instance
func (d *dict) parse(b *buffer) error {

	count, err := b.readLineInt()
	if err != nil {
		return err
	}

	d.vals = make(map[interface{}]interface{})

	for i := 0; i < count; i++ {
		val, err := parseElem(b)
		if err != nil {
			return fmt.Errorf("error parsing Dict value: %s", err)
		}

		key, err := parseElem(b)
		if err != nil {
			return fmt.Errorf("error parsing Dict key: %s", err)
		}

		d.vals[key] = val
	}

	return nil
}

// unmarshal takes values stored in Dict and copies to Value. Value must be
// setable and be a kind of map.
func (d *dict) unmarshal(rv reflect.Value) error {
	if rv.Kind() != reflect.Map {
		return fmt.Errorf("cannot unmarshal Dict into non-map type %s", rv.Type())
	}

	elemType := rv.Type().Elem()

	rv.Set(reflect.MakeMap(rv.Type()))

	for k, v := range d.vals {
		newVal, err := unmarshalType(elemType, v)
		if err != nil {
			return err
		}

		keyVal, err := unmarshalType(rv.Type().Key(), k)
		if err != nil {
			return err
		}

		rv.SetMapIndex(keyVal, newVal)
	}

	return nil
}

// obj is a representation of a generic python object/class
type obj struct {
	name  string
	attrs map[string]interface{}
}

// newObj returns a new Obj instance
func newObj(s string) *obj {
	var o obj

	o.name = s
	o.attrs = make(map[string]interface{})

	return &o
}

// parse decodes a cereal format data defintion into the Obj instance
func (o *obj) parse(b *buffer) error {
	val, err := parseElem(b)
	if err != nil {
		return fmt.Errorf("error parsing Obj value: %s", err)
	}

	d, ok := val.(*dict)
	if !ok {
		return fmt.Errorf("illegal reference for Obj data: must be type *Dict")
	}

	for k, v := range d.vals {
		str, ok := k.(string)
		if !ok {
			return fmt.Errorf("cannot assign non-string key '%v' to Obj attribute name", k)
		}
		// fmt.Printf("  setting Obj.%s = %#v\n", str, v)
		o.attrs[str] = v
	}

	return nil
}

// unmarshal takes values stored in Dict and copies to Value. Value must be
// setable and be a kind of struct. unmarshal will use field tags with the
// key 'cereal' to map between python attributes and struct fields.
func (o *obj) unmarshal(rv reflect.Value) error {

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("cannot unmarshal Obj into non-struct type %s", rv.Type())
	}

	rt := rv.Type()

	// Create a map of field names to struct fields

	// TODO: consider caching maps by type so we do not repeat operation each time
	//       we are invoked.

	// TODO: automatically handle the case of mapping all lower-case python attribute
	//       name to capitalized (exported) struct field (if field exists)
	fieldMap := make(map[string]int)

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		if tag, ok := f.Tag.Lookup("cereal"); ok {
			// fmt.Printf("Mapping %s -> %s.%s\n", tag, rt.Name(), f.Name)
			fieldMap[tag] = i
		} else {
			// fmt.Printf("Mapping %s -> %s.%s\n", f.Name, rt.Name(), f.Name)
			fieldMap[f.Name] = i
		}
	}

	for key, value := range o.attrs {
		idx, ok := fieldMap[key]
		if !ok {
			continue
		}

		if !rv.Field(idx).CanSet() {
			// XXX: Should we report error or ignore? Currently ignore
			continue
		}

		if p, ok := value.(parser); ok {
			err := p.unmarshal(rv.Field(idx))
			if err != nil {
				return err
			}
			continue
		}

		newVal, err := unmarshalType(rv.Field(idx).Type(), value)
		if err != nil {
			return err
		}

		rv.Field(idx).Set(newVal)
	}

	return nil
}
