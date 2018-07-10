// Copyright (c) 2018 James Bowen

//go:generate python testdata/generate_data.py

package cereal

import (
	"io/ioutil"
	"reflect"
	"testing"
)

// test types need to match the python class definitions from generate_data.py

type test01 struct {
	Int   int     `cereal:"field_int"`
	Str   string  `cereal:"field_str"`
	Float float64 `cereal:"field_float"`
	Bool  bool    `cereal:"field_bool"`
}

type test01a struct {
	Int int    `cereal:"field_int"`
	Str string `cereal:"field_str"`
	// unexported fields
	notFloat float64 `cereal:"field_float"`
	none     int
}

type test02 struct {
	ListInt    []int          `cereal:"field_list"`
	ListFloat  []float32      `cereal:"field_list_float"`
	ListString []string       `cereal:"field_list_string"`
	DictStrInt map[string]int `cereal:"field_dict"`
	TupleInt   []int          `cereal:"field_tuple"`
	Bool       bool           `cereal:"field_bool"`
}

type test03 struct {
	Obj1 test01 `cereal:"field_obj1"`
	Obj2 test02 `cereal:"field_obj2"`
}

func Test01(t *testing.T) {

	var rootObj test01

	expected := test01{
		Int:   12345,
		Str:   "one two three four five",
		Float: -1.234E9,
		Bool:  true,
	}

	data, err := ioutil.ReadFile("testdata/test01.dat")
	if err != nil {
		t.Errorf("Error reading data file: %s", err)
	}

	err = Unmarshal(data, &rootObj)
	if err != nil {
		t.Fatalf("Unmarshal() error: %s", err)
	}

	t.Logf("returned:  %#v\n", rootObj)
	t.Logf("expected:  %#v\n", expected)

	if !reflect.DeepEqual(rootObj, expected) {
		t.Errorf("Returned object != expected")
	}
}

func Test02(t *testing.T) {

	var rootObj test02

	expected := test02{
		ListInt:    []int{1, 2, 3, 4, 5, 300},
		ListFloat:  []float32{1.0, 4.1, 5.0, -3.2},
		ListString: []string{"aaaa", "bbbb", "cccc", "dddd"},
		DictStrInt: map[string]int{"key1": 1234, "key2": 5678, "key3": 9012},
		TupleInt:   []int{1, 2, 3, 4},
		Bool:       false,
	}

	data, err := ioutil.ReadFile("testdata/test02.dat")
	if err != nil {
		t.Errorf("Error reading data file: %s", err)
	}

	err = Unmarshal(data, &rootObj)
	if err != nil {
		t.Fatalf("Unmarshal() error: %s", err)
	}

	t.Logf("returned:  %#v\n", rootObj)
	t.Logf("expected:  %#v\n", expected)

	if !reflect.DeepEqual(rootObj, expected) {
		t.Errorf("Returned object != expected")
	}
}

func Test03(t *testing.T) {

	var rootObj test03

	expected := test03{
		Obj1: test01{
			Int:   12345,
			Str:   "one two three four five",
			Float: -1.234E9,
			Bool:  true,
		},
		Obj2: test02{
			ListInt:    []int{1, 2, 3, 4, 5, 300},
			ListFloat:  []float32{1.0, 4.1, 5.0, -3.2},
			ListString: []string{"aaaa", "bbbb", "cccc", "dddd"},
			DictStrInt: map[string]int{"key1": 1234, "key2": 5678, "key3": 9012},
			TupleInt:   []int{1, 2, 3, 4},
			Bool:       false,
		},
	}

	data, err := ioutil.ReadFile("testdata/test03.dat")
	if err != nil {
		t.Errorf("Error reading data file: %s", err)
	}

	err = Unmarshal(data, &rootObj)
	if err != nil {
		t.Fatalf("Unmarshal() error: %s", err)
	}

	t.Logf("returned:  %#v\n", rootObj)
	t.Logf("expected:  %#v\n", expected)

	if !reflect.DeepEqual(rootObj, expected) {
		t.Errorf("Returned object != expected")
	}
}
