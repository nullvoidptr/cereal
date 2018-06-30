package cereal

import (
	"fmt"
	"reflect"
)

// This files contains types representing python objects read for a cereal-format
// datastream. The standard method of parsing the cereal file is to first create
// intermediate instances of these types (and/or built-in types like int, string, etc..)
// and then perform an unmarshal operation into user provided target type.

// List is a representation of a python list
type List struct {
	vals []interface{}
}

// NewList returns a new List instance
func NewList() *List {
	return &List{}
}

// Parse decodes a cereal format data defintion into the List instance
func (l *List) Parse(b *buffer) error {

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

		// fmt.Printf("  setting List[%d] = %v\n", i, val)
		l.vals[i] = val
	}

	// fmt.Printf("  List = %#v\n", l)
	return nil
}

// Unmarshal takes values stored in List and copies to Value. Value must be
// setable and be a kind of slice.
func (l *List) Unmarshal(rv reflect.Value) error {

	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal List into non-slice type %s", rv.Type())
	}

	elemType := rv.Type().Elem()

	for _, v := range l.vals {
		// fmt.Printf("  assigning List[%d] = %v\n", i, v)

		newVal, err := unmarshalType(elemType, v)
		if err != nil {
			return err
		}

		rv.Set(reflect.Append(rv, newVal))
	}
	return nil
}

// Set is a representation of a python set object. It is functionally
// identical to List but stored differently by cerealizer.
type Set struct {
	List
}

// NewSet returns a new Set instance
func NewSet() *Set {
	return &Set{}
}

// Tuple is a representaion of a python tuple. It is functionally
// identical to List but stored differently by cerealizer.
type Tuple struct {
	List
}

// NewTuple returns a new Tuple instance
func NewTuple() *Tuple {
	return &Tuple{}
}

// Dict is a representation of a python dictionary
type Dict struct {
	vals map[interface{}]interface{}
}

// NewDict returns a new Dict instance
func NewDict() *Dict {
	var d Dict

	d.vals = make(map[interface{}]interface{})
	return &d
}

// Parse decodes a cereal format data defintion into the Dict instance
func (d *Dict) Parse(b *buffer) error {
	// fmt.Println("Parsing *Dict data")

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

		// fmt.Printf("  setting Dict[%v] = %v\n", key, val)
		d.vals[key] = val
	}

	// fmt.Printf("  Dict = %#v\n", d)
	return nil
}

// Unmarshal takes values stored in Dict and copies to Value. Value must be
// setable and be a kind of map.
func (d *Dict) Unmarshal(rv reflect.Value) error {
	if rv.Kind() != reflect.Map {
		return fmt.Errorf("cannot unmarshal Dict into non-map type %s", rv.Type())
	}

	elemType := rv.Type().Elem()

	rv.Set(reflect.MakeMap(rv.Type()))

	for k, v := range d.vals {
		// fmt.Printf("  assigning Dict[%v] = %v\n", k, v)

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

// Obj is a representation of a generic python object/class
type Obj struct {
	name  string
	attrs map[string]interface{}
}

// NewObj returns a new Obj instance
func NewObj(s string) *Obj {
	var o Obj

	o.name = s
	o.attrs = make(map[string]interface{})

	return &o
}

// Parse decodes a cereal format data defintion into the Obj instance
func (o *Obj) Parse(b *buffer) error {
	// fmt.Println("Parsing *Obj data")

	val, err := parseElem(b)
	if err != nil {
		return fmt.Errorf("error parsing Obj value: %s", err)
	}

	dict, ok := val.(*Dict)
	if !ok {
		return fmt.Errorf("illegal reference for Obj data: must be type *Dict")
	}

	for k, v := range dict.vals {
		str, ok := k.(string)
		if !ok {
			return fmt.Errorf("cannot assign non-string key '%v' to Obj attribute name", k)
		}
		// fmt.Printf("  setting Obj.%s = %#v\n", str, v)
		o.attrs[str] = v
	}

	return nil
}

// Unmarshal takes values stored in Dict and copies to Value. Value must be
// setable and be a kind of struct. Unmarshal will use field tags with the
// key 'cereal' to map between python attributes and struct fields.
func (o *Obj) Unmarshal(rv reflect.Value) error {

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
			// fmt.Printf("skipping attribute %s: no matching field\n", key)
			continue
		}

		if !rv.Field(idx).CanSet() {
			// fmt.Printf("skipping non-settable (unexported?) field %s\n", key)
			continue
		}
		// fmt.Printf("Setting field %s = %v\n", rt.Field(idx).Name, value)

		if parser, ok := value.(Parser); ok {
			err := parser.Unmarshal(rv.Field(idx))
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
