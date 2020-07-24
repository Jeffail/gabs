package gabs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {
	sample := []byte(`{"test":{"value":10},"test2":20}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if result, ok := val.Search([]string{"test", "value"}...).Data().(float64); ok {
		if result != 10 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find test.value")
	}

	if _, ok := val.Search("test2", "value").Data().(string); ok {
		t.Errorf("Somehow found a field that shouldn't exist")
	}

	if result, ok := val.Search("test2").Data().(float64); ok {
		if result != 20 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find test2")
	}

	if result := val.Bytes(); string(result) != string(sample) {
		t.Errorf("Wrong []byte conversion: %s != %s", result, sample)
	}
}

func TestNilMethods(t *testing.T) {
	var n *Container
	if exp, act := "null", n.String(); exp != act {
		t.Errorf("Unexpected value: %v != %v", act, exp)
	}
	if exp, act := "null", string(n.Bytes()); exp != act {
		t.Errorf("Unexpected value: %v != %v", act, exp)
	}
	if n.Search("foo", "bar") != nil {
		t.Error("non nil result")
	}
	if n.Path("foo.bar") != nil {
		t.Error("non nil result")
	}
	if _, err := n.Array("foo"); err == nil {
		t.Error("expected error")
	}
	if err := n.ArrayAppend("foo", "bar"); err == nil {
		t.Error("expected error")
	}
	if err := n.ArrayRemove(1, "foo", "bar"); err == nil {
		t.Error("expected error")
	}
	if n.Exists("foo", "bar") {
		t.Error("expected false")
	}
	if n.Index(1) != nil {
		t.Error("non nil result")
	}
	if n.Children() != nil {
		t.Error("non nil result")
	}
	if len(n.ChildrenMap()) > 0 {
		t.Error("non nil result")
	}
	if err := n.Delete("foo"); err == nil {
		t.Error("expected error")
	}
}

var bigSample = []byte(`{
	"a": {
		"nested1": {
			"value1": 5
		}
	},
	"": {
		"can we access": "this?"
	},
	"what/a/pain": "ouch1",
	"what~a~pain": "ouch2",
	"what~/a/~pain": "ouch3",
	"what.a.pain": "ouch4",
	"what~.a.~pain": "ouch5",
	"b": 10,
	"c": [
		"first",
		"second",
		{
			"nested2": {
				"value2": 15
			}
		},
		[
			"fifth",
			"sixth"
		],
		"fourth"
	],
	"d": {
		"": {
			"what about": "this?"
		}
	}
}`)

func TestJSONPointer(t *testing.T) {
	type testCase struct {
		path  string
		value string
		err   string
	}
	tests := []testCase{
		{
			path: "foo",
			err:  "failed to resolve JSON pointer: path must begin with '/'",
		},
		{
			path: "/a/doesnotexist",
			err:  "failed to resolve path segment '1': key 'doesnotexist' was not found",
		},
		{
			path:  "/a",
			value: `{"nested1":{"value1":5}}`,
		},
		{
			path:  "/what~1a~1pain",
			value: `"ouch1"`,
		},
		{
			path:  "/what~0a~0pain",
			value: `"ouch2"`,
		},
		{
			path:  "/what~0~1a~1~0pain",
			value: `"ouch3"`,
		},
		{
			path:  "/what.a.pain",
			value: `"ouch4"`,
		},
		{
			path:  "/what~0.a.~0pain",
			value: `"ouch5"`,
		},
		{
			path:  "/",
			value: `{"can we access":"this?"}`,
		},
		{
			path:  "//can we access",
			value: `"this?"`,
		},
		{
			path:  "/d/",
			value: `{"what about":"this?"}`,
		},
		{
			path:  "/d//what about",
			value: `"this?"`,
		},
		{
			path:  "/c/1",
			value: `"second"`,
		},
		{
			path:  "/c/2/nested2/value2",
			value: `15`,
		},
		{
			path: "/c/notindex/value2",
			err:  `failed to resolve path segment '1': found array but segment value 'notindex' could not be parsed into array index: strconv.Atoi: parsing "notindex": invalid syntax`,
		},
		{
			path: "/c/10/value2",
			err:  `failed to resolve path segment '1': found array but index '10' exceeded target array size of '5'`,
		},
	}

	root, err := ParseJSON(bigSample)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	for _, test := range tests {
		t.Run(test.path, func(tt *testing.T) {
			var result *Container
			result, err = root.JSONPointer(test.path)
			if len(test.err) > 0 {
				if err == nil {
					tt.Errorf("Expected error: %v", test.err)
				} else if exp, act := test.err, err.Error(); exp != act {
					tt.Errorf("Wrong error returned: %v != %v", act, exp)
				}
				return
			} else if err != nil {
				tt.Error(err)
				return
			}
			if exp, act := test.value, result.String(); exp != act {
				tt.Errorf("Wrong result: %v != %v", act, exp)
			}
		})
	}
}

func TestDotPath(t *testing.T) {
	type testCase struct {
		path  string
		value string
	}
	tests := []testCase{
		{
			path:  "foo",
			value: "null",
		},
		{
			path:  "a.doesnotexist",
			value: "null",
		},
		{
			path:  "a",
			value: `{"nested1":{"value1":5}}`,
		},
		{
			path:  "what/a/pain",
			value: `"ouch1"`,
		},
		{
			path:  "what~0a~0pain",
			value: `"ouch2"`,
		},
		{
			path:  "what~0/a/~0pain",
			value: `"ouch3"`,
		},
		{
			path:  "what~1a~1pain",
			value: `"ouch4"`,
		},
		{
			path:  "what~0~1a~1~0pain",
			value: `"ouch5"`,
		},
		{
			path:  "",
			value: `{"can we access":"this?"}`,
		},
		{
			path:  ".can we access",
			value: `"this?"`,
		},
		{
			path:  "d.",
			value: `{"what about":"this?"}`,
		},
		{
			path:  "d..what about",
			value: `"this?"`,
		},
		{
			path:  "c.1",
			value: `"second"`,
		},
		{
			path:  "c.2.nested2.value2",
			value: `15`,
		},
		{
			path:  "c.notindex.value2",
			value: "null",
		},
		{
			path:  "c.10.value2",
			value: "null",
		},
	}

	root, err := ParseJSON(bigSample)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	for _, test := range tests {
		t.Run(test.path, func(tt *testing.T) {
			result := root.Path(test.path)
			if exp, act := test.value, result.String(); exp != act {
				tt.Errorf("Wrong result: %v != %v", act, exp)
			}
		})
	}
}

func TestArrayWildcard(t *testing.T) {
	sample := []byte(`{"test":[{"value":10},{"value":20}]}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if act, ok := val.Search([]string{"test", "0", "value"}...).Data().(float64); ok {
		if exp := float64(10); !reflect.DeepEqual(act, exp) {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	} else {
		t.Errorf("Didn't find test.0.value")
	}

	if act, ok := val.Search([]string{"test", "1", "value"}...).Data().(float64); ok {
		if exp := float64(20); !reflect.DeepEqual(act, exp) {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	} else {
		t.Errorf("Didn't find test.1.value")
	}

	if act, ok := val.Search([]string{"test", "*", "value"}...).Data().([]interface{}); ok {
		if exp := []interface{}{float64(10), float64(20)}; !reflect.DeepEqual(act, exp) {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	} else {
		t.Errorf("Didn't find test.*.value")
	}

	if act := val.Search([]string{"test", "*", "notmatched"}...); act != nil {
		t.Errorf("Expected nil result, received: %v", act)
	}

	if act, ok := val.Search([]string{"test", "*"}...).Data().([]interface{}); ok {
		if exp := []interface{}{map[string]interface{}{"value": float64(10)}, map[string]interface{}{"value": float64(20)}}; !reflect.DeepEqual(act, exp) {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	} else {
		t.Errorf("Didn't find test.*.value")
	}
}

func TestArrayAppendWithSet(t *testing.T) {
	gObj := New()
	if _, err := gObj.Set([]interface{}{}, "foo"); err != nil {
		t.Fatal(err)
	}
	if _, err := gObj.Set(1, "foo", "-"); err != nil {
		t.Fatal(err)
	}
	if _, err := gObj.Set([]interface{}{}, "foo", "-", "baz"); err != nil {
		t.Fatal(err)
	}
	if _, err := gObj.Set(2, "foo", "1", "baz", "-"); err != nil {
		t.Fatal(err)
	}
	if _, err := gObj.Set(3, "foo", "1", "baz", "-"); err != nil {
		t.Fatal(err)
	}
	if _, err := gObj.Set(4, "foo", "-"); err != nil {
		t.Fatal(err)
	}

	exp := `{"foo":[1,{"baz":[2,3]},4]}`
	if act := gObj.String(); act != exp {
		t.Errorf("Unexpected value: %v != %v", act, exp)
	}
}

func TestExists(t *testing.T) {
	sample := []byte(`{"test":{"value":10,"nullvalue":null},"test2":20,"testnull":null}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	paths := []struct {
		Path   []string
		Exists bool
	}{
		{[]string{"one", "two", "three"}, false},
		{[]string{"test"}, true},
		{[]string{"test", "value"}, true},
		{[]string{"test", "nullvalue"}, true},
		{[]string{"test2"}, true},
		{[]string{"testnull"}, true},
		{[]string{"test2", "value"}, false},
		{[]string{"test", "value2"}, false},
		{[]string{"test", "VALUE"}, false},
	}

	for _, p := range paths {
		if exp, actual := p.Exists, val.Exists(p.Path...); exp != actual {
			t.Errorf("Wrong result from Exists: %v != %v, for path: %v", exp, actual, p.Path)
		}
		if exp, actual := p.Exists, val.ExistsP(strings.Join(p.Path, ".")); exp != actual {
			t.Errorf("Wrong result from ExistsP: %v != %v, for path: %v", exp, actual, p.Path)
		}
	}
}

func TestExistsWithArrays(t *testing.T) {
	sample := []byte(`{"foo":{"bar":{"baz":[10, 2, 3]}}}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if exp, actual := true, val.Exists("foo", "bar", "baz"); exp != actual {
		t.Errorf("Wrong result from array based Exists: %v != %v", exp, actual)
	}

	sample = []byte(`{"foo":{"bar":[{"baz":10},{"baz":2},{"baz":3}]}}`)

	if val, err = ParseJSON(sample); err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if exp, actual := true, val.Exists("foo", "bar", "0", "baz"); exp != actual {
		t.Errorf("Wrong result from array based Exists: %v != %v", exp, actual)
	}
	if exp, actual := false, val.Exists("foo", "bar", "1", "baz_NOPE"); exp != actual {
		t.Errorf("Wrong result from array based Exists: %v != %v", exp, actual)
	}

	sample = []byte(`{"foo":[{"bar":{"baz":10}},{"bar":{"baz":2}},{"bar":{"baz":3}}]}`)

	if val, err = ParseJSON(sample); err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if exp, actual := true, val.Exists("foo", "0", "bar", "baz"); exp != actual {
		t.Errorf("Wrong result from array based Exists: %v != %v", exp, actual)
	}
	if exp, actual := false, val.Exists("foo", "0", "bar", "baz_NOPE"); exp != actual {
		t.Errorf("Wrong result from array based Exists: %v != %v", exp, actual)
	}

	sample =
		[]byte(`[{"foo":{"bar":{"baz":10}}},{"foo":{"bar":{"baz":2}}},{"foo":{"bar":{"baz":3}}}]`)

	if val, err = ParseJSON(sample); err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if exp, actual := true, val.Exists("0", "foo", "bar", "baz"); exp != actual {
		t.Errorf("Wrong result from array based Exists: %v != %v", exp, actual)
	}
	if exp, actual := false, val.Exists("0", "foo", "bar", "baz_NOPE"); exp != actual {
		t.Errorf("Wrong result from array based Exists: %v != %v", exp, actual)
	}
}

func TestBasicWithBuffer(t *testing.T) {
	sample := bytes.NewReader([]byte(`{"test":{"value":10},"test2":20}`))

	_, err := ParseJSONBuffer(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}
}

func TestBasicWithDecoder(t *testing.T) {
	sample := []byte(`{"test":{"int":10, "float":6.66}}`)
	dec := json.NewDecoder(bytes.NewReader(sample))
	dec.UseNumber()

	val, err := ParseJSONDecoder(dec)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	checkNumber := func(path string, expectedVal json.Number) {
		data := val.Path(path).Data()
		asNumber, isNumber := data.(json.Number)
		if !isNumber {
			t.Error("Failed to parse using decoder UseNumber policy")
		}
		if expectedVal != asNumber {
			t.Errorf("Expected[%s] but got [%s]", expectedVal, asNumber)
		}
	}

	checkNumber("test.int", "10")
	checkNumber("test.float", "6.66")
}

func TestFailureWithDecoder(t *testing.T) {
	sample := []byte(`{"test":{" "invalidCrap":.66}}`)
	dec := json.NewDecoder(bytes.NewReader(sample))

	_, err := ParseJSONDecoder(dec)
	if err == nil {
		t.Fatal("Expected parsing error")
	}
}

func TestDeletes(t *testing.T) {
	jsonParsed, _ := ParseJSON([]byte(`{
		"outter":{
			"inner":{
				"value1":10,
				"value2":22,
				"value3":32
			},
			"alsoInner":{
				"value1":20,
				"value2":42,
				"value3":92
			},
			"another":{
				"value1":null,
				"value2":null,
				"value3":null
			}
		}
	}`))

	if err := jsonParsed.Delete("outter", "inner", "value2"); err != nil {
		t.Error(err)
	}
	if err := jsonParsed.Delete("outter", "inner", "value4"); err == nil {
		t.Error("value4 should not have been found in outter.inner")
	}
	if err := jsonParsed.Delete("outter", "another", "value1"); err != nil {
		t.Error(err)
	}
	if err := jsonParsed.Delete("outter", "another", "value4"); err == nil {
		t.Error("value4 should not have been found in outter.another")
	}
	if err := jsonParsed.DeleteP("outter.alsoInner.value1"); err != nil {
		t.Error(err)
	}
	if err := jsonParsed.DeleteP("outter.alsoInner.value4"); err == nil {
		t.Error("value4 should not have been found in outter.alsoInner")
	}
	if err := jsonParsed.DeleteP("outter.another.value2"); err != nil {
		t.Error(err)
	}
	if err := jsonParsed.Delete("outter.another.value4"); err == nil {
		t.Error("value4 should not have been found in outter.another")
	}

	expected := `{"outter":{"alsoInner":{"value2":42,"value3":92},"another":{"value3":null},"inner":{"value1":10,"value3":32}}}`
	if actual := jsonParsed.String(); actual != expected {
		t.Errorf("Unexpected result from deletes: %v != %v", actual, expected)
	}
}

func TestDeletesWithArrays(t *testing.T) {
	rawJSON := `{
		"outter":[
			{
				"foo":{
					"value1":10,
					"value2":22,
					"value3":32
				},
				"bar": [
					20,
					42,
					92
				]
			},
			{
				"baz":{
					"value1":null,
					"value2":null,
					"value3":null
				}
			}
		]
	}`

	jsonParsed, err := ParseJSON([]byte(rawJSON))
	if err != nil {
		t.Fatal(err)
	}
	if err = jsonParsed.Delete("outter", "1", "baz", "value1"); err != nil {
		t.Error(err)
	}

	expected := `{"outter":[{"bar":[20,42,92],"foo":{"value1":10,"value2":22,"value3":32}},{"baz":{"value2":null,"value3":null}}]}`
	if actual := jsonParsed.String(); actual != expected {
		t.Errorf("Unexpected result from array deletes: %v != %v", actual, expected)
	}

	jsonParsed, err = ParseJSON([]byte(rawJSON))
	if err != nil {
		t.Fatal(err)
	}
	if err = jsonParsed.Delete("outter", "1", "baz"); err != nil {
		t.Error(err)
	}

	expected = `{"outter":[{"bar":[20,42,92],"foo":{"value1":10,"value2":22,"value3":32}},{}]}`
	if actual := jsonParsed.String(); actual != expected {
		t.Errorf("Unexpected result from array deletes: %v != %v", actual, expected)
	}

	jsonParsed, err = ParseJSON([]byte(rawJSON))
	if err != nil {
		t.Fatal(err)
	}
	if err = jsonParsed.Delete("outter", "1"); err != nil {
		t.Error(err)
	}

	expected = `{"outter":[{"bar":[20,42,92],"foo":{"value1":10,"value2":22,"value3":32}}]}`
	if actual := jsonParsed.String(); actual != expected {
		t.Errorf("Unexpected result from array deletes: %v != %v", actual, expected)
	}

	jsonParsed, err = ParseJSON([]byte(rawJSON))
	if err != nil {
		t.Fatal(err)
	}
	if err = jsonParsed.Delete("outter", "0", "bar", "0"); err != nil {
		t.Error(err)
	}

	expected = `{"outter":[{"bar":[42,92],"foo":{"value1":10,"value2":22,"value3":32}},{"baz":{"value1":null,"value2":null,"value3":null}}]}`
	if actual := jsonParsed.String(); actual != expected {
		t.Errorf("Unexpected result from array deletes: %v != %v", actual, expected)
	}

	jsonParsed, err = ParseJSON([]byte(rawJSON))
	if err != nil {
		t.Fatal(err)
	}
	if err = jsonParsed.Delete("outter", "0", "bar", "1"); err != nil {
		t.Error(err)
	}

	expected = `{"outter":[{"bar":[20,92],"foo":{"value1":10,"value2":22,"value3":32}},{"baz":{"value1":null,"value2":null,"value3":null}}]}`
	if actual := jsonParsed.String(); actual != expected {
		t.Errorf("Unexpected result from array deletes: %v != %v", actual, expected)
	}

	jsonParsed, err = ParseJSON([]byte(rawJSON))
	if err != nil {
		t.Fatal(err)
	}
	if err = jsonParsed.Delete("outter", "0", "bar", "2"); err != nil {
		t.Error(err)
	}

	expected = `{"outter":[{"bar":[20,42],"foo":{"value1":10,"value2":22,"value3":32}},{"baz":{"value1":null,"value2":null,"value3":null}}]}`
	if actual := jsonParsed.String(); actual != expected {
		t.Errorf("Unexpected result from array deletes: %v != %v", actual, expected)
	}
}

func TestExamples(t *testing.T) {
	jsonParsed, err := ParseJSON([]byte(`{
	"outter":{
		"inner":{
			"value1":10,
			"value2":22
		},
		"contains.dots.in.key":{
			"value1":11
		},
		"contains~tildes~in~key":{
			"value1":12
		},
		"alsoInner":{
			"value1":20,
			"array1":[
				30, 40
			]
		}
	}
}`))

	var value float64
	var ok bool

	value, ok = jsonParsed.Path("outter.inner.value1").Data().(float64)
	if value != 10.0 || !ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	value, ok = jsonParsed.Path("outter.contains~1dots~1in~1key.value1").Data().(float64)
	if value != 11.0 || !ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	value, ok = jsonParsed.Path("outter.contains~0tildes~0in~0key.value1").Data().(float64)
	if value != 12.0 || !ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	value, ok = jsonParsed.Search("outter", "inner", "value1").Data().(float64)
	if value != 10.0 || !ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	gObj, err := jsonParsed.JSONPointer("/outter/alsoInner/array1/1")
	if err != nil {
		t.Fatal(err)
	}
	value, ok = gObj.Data().(float64)
	if value != 40.0 || !ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	value, ok = jsonParsed.Path("does.not.exist").Data().(float64)
	if value != 0.0 || ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	jsonParsed, _ = ParseJSON([]byte(`{"array":[ "first", "second", "third" ]}`))

	expected := []string{"first", "second", "third"}

	children := jsonParsed.S("array").Children()
	for i, child := range children {
		if expected[i] != child.Data().(string) {
			t.Errorf("Child unexpected: %v != %v", expected[i], child.Data().(string))
		}
	}
}

func TestSetAppendArray(t *testing.T) {
	content := []byte(`{
	"nested": {
		"source": [
			"foo", "bar"
		]
	}
}`)

	gObj, err := ParseJSON(content)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = gObj.Set("baz", "nested", "source", "-"); err != nil {
		t.Fatal(err)
	}
	exp := `{"nested":{"source":["foo","bar","baz"]}}`
	if act := gObj.String(); act != exp {
		t.Errorf("Wrong result: %v != %v", act, exp)
	}
}

func TestExamples2(t *testing.T) {
	var err error

	jsonObj := New()

	_, err = jsonObj.Set(10, "outter", "inner", "value")
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}
	_, err = jsonObj.SetP(20, "outter.inner.value2")
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}
	_, err = jsonObj.Set(30, "outter", "inner2", "value3")
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	expected := `{"outter":{"inner":{"value":10,"value2":20},"inner2":{"value3":30}}}`
	if jsonObj.String() != expected {
		t.Errorf("Non matched output: %v != %v", expected, jsonObj.String())
	}

	jsonObj = Wrap(map[string]interface{}{})

	jsonObj.Array("array")

	jsonObj.ArrayAppend(10, "array")
	jsonObj.ArrayAppend(20, "array")
	jsonObj.ArrayAppend(30, "array")

	expected = `{
      "array": [
        10,
        20,
        30
      ]
    }`
	result := jsonObj.StringIndent("    ", "  ")
	if result != expected {
		t.Errorf("Non matched output: %v != %v", expected, result)
	}
}

func TestExamples3(t *testing.T) {
	jsonObj := New()

	jsonObj.ArrayP("foo.array")

	jsonObj.ArrayAppend(10, "foo", "array")
	jsonObj.ArrayAppend(20, "foo", "array")
	jsonObj.ArrayAppend(30, "foo", "array")

	result := jsonObj.String()
	expected := `{"foo":{"array":[10,20,30]}}`

	if result != expected {
		t.Errorf("Non matched output: %v != %v", result, expected)
	}
}

func TestArrayConcat(t *testing.T) {
	jsonObj := New()

	jsonObj.ArrayP("foo.array")

	jsonObj.ArrayConcat(10, "foo", "array")
	jsonObj.ArrayConcat([]interface{}{20, 30}, "foo", "array")

	result := jsonObj.String()
	expected := `{"foo":{"array":[10,20,30]}}`

	if result != expected {
		t.Errorf("Non matched output: %v != %v", result, expected)
	}

	jsonObj = New()

	jsonObj.ArrayP("foo.array")

	jsonObj.ArrayConcat([]interface{}{10, 20}, "foo", "array")
	jsonObj.ArrayConcat(30, "foo", "array")

	result = jsonObj.String()
	expected = `{"foo":{"array":[10,20,30]}}`

	if result != expected {
		t.Errorf("Non matched output: %v != %v", result, expected)
	}

	jsonObj = New()

	jsonObj.ArrayP("foo.array")

	jsonObj.ArrayConcat([]interface{}{10}, "foo", "array")
	jsonObj.ArrayConcat([]interface{}{20}, "foo", "array")
	jsonObj.ArrayConcat([]interface{}{30}, "foo", "array")

	result = jsonObj.String()
	expected = `{"foo":{"array":[10,20,30]}}`

	if result != expected {
		t.Errorf("Non matched output: %v != %v", result, expected)
	}
}

func TestArrayConcatP(t *testing.T) {
	jsonObj := New()

	jsonObj.ArrayP("foo.array")

	jsonObj.ArrayConcatP(10, "foo.array")
	jsonObj.ArrayConcatP([]interface{}{20, 30}, "foo.array")

	result := jsonObj.String()
	expected := `{"foo":{"array":[10,20,30]}}`

	if result != expected {
		t.Errorf("Non matched output: %v != %v", result, expected)
	}

	jsonObj = New()

	jsonObj.ArrayP("foo.array")

	jsonObj.ArrayConcatP([]interface{}{10, 20}, "foo.array")
	jsonObj.ArrayConcatP(30, "foo.array")

	result = jsonObj.String()
	expected = `{"foo":{"array":[10,20,30]}}`

	if result != expected {
		t.Errorf("Non matched output: %v != %v", result, expected)
	}

	jsonObj = New()

	jsonObj.ArrayP("foo.array")

	jsonObj.ArrayConcatP([]interface{}{10}, "foo.array")
	jsonObj.ArrayConcatP([]interface{}{20}, "foo.array")
	jsonObj.ArrayConcatP([]interface{}{30}, "foo.array")

	result = jsonObj.String()
	expected = `{"foo":{"array":[10,20,30]}}`

	if result != expected {
		t.Errorf("Non matched output: %v != %v", result, expected)
	}
}

func TestDotNotation(t *testing.T) {
	sample := []byte(`{"test":{"inner":{"value":10}},"test2":20}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if result, _ := val.Path("test.inner.value").Data().(float64); result != 10 {
		t.Errorf("Expected 10, received: %v", result)
	}
}

func TestModify(t *testing.T) {
	sample := []byte(`{"test":{"value":10},"test2":20}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if _, err := val.S("test").Set(45.0, "value"); err != nil {
		t.Errorf("Failed to set field")
	}

	if result, ok := val.Search([]string{"test", "value"}...).Data().(float64); ok {
		if result != 45 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find test.value")
	}

	if out := val.String(); `{"test":{"value":45},"test2":20}` != out {
		t.Errorf("Incorrectly serialized: %v", out)
	}

	if out := val.Search("test").String(); `{"value":45}` != out {
		t.Errorf("Incorrectly serialized: %v", out)
	}
}

func TestChildren(t *testing.T) {
	json1, _ := ParseJSON([]byte(`{
		"objectOne":{
		},
		"objectTwo":{
		},
		"objectThree":{
		}
	}`))

	objects := json1.Children()
	for _, object := range objects {
		object.Set("hello world", "child")
	}

	expected := `{"objectOne":{"child":"hello world"},"objectThree":{"child":"hello world"}` +
		`,"objectTwo":{"child":"hello world"}}`
	received := json1.String()
	if expected != received {
		t.Errorf("json1: expected %v, received %v", expected, received)
	}

	json2, _ := ParseJSON([]byte(`{
		"values":[
			{
				"objectOne":{
				}
			},
			{
				"objectTwo":{
				}
			},
			{
				"objectThree":{
				}
			}
		]
	}`))

	json3, _ := ParseJSON([]byte(`{
		"values":[
		]
	}`))

	numChildren1, _ := json2.ArrayCount("values")
	numChildren2, _ := json3.ArrayCount("values")
	if _, err := json3.ArrayCount("valuesNOTREAL"); err == nil {
		t.Errorf("expected numChildren3 to fail")
	}

	if numChildren1 != 3 || numChildren2 != 0 {
		t.Errorf("CountElements, expected 3 and 0, received %v and %v",
			numChildren1, numChildren2)
	}

	objects = json2.S("values").Children()
	for _, object := range objects {
		object.Set("hello world", "child")
		json3.ArrayAppend(object.Data(), "values")
	}

	expected = `{"values":[{"child":"hello world","objectOne":{}},{"child":"hello world",` +
		`"objectTwo":{}},{"child":"hello world","objectThree":{}}]}`
	received = json2.String()
	if expected != received {
		t.Errorf("json2: expected %v, received %v", expected, received)
	}

	received = json3.String()
	if expected != received {
		t.Errorf("json3: expected %v, received %v", expected, received)
	}
}

func TestChildrenMap(t *testing.T) {
	json1, _ := ParseJSON([]byte(`{
		"objectOne":{"num":1},
		"objectTwo":{"num":2},
		"objectThree":{"num":3}
	}`))

	objectMap := json1.ChildrenMap()
	if len(objectMap) != 3 {
		t.Errorf("Wrong num of elements in objectMap: %v != %v", len(objectMap), 3)
		return
	}

	for key, val := range objectMap {
		if "objectOne" == key {
			if val := val.S("num").Data().(float64); val != 1 {
				t.Errorf("%v != %v", val, 1)
			}
		} else if "objectTwo" == key {
			if val := val.S("num").Data().(float64); val != 2 {
				t.Errorf("%v != %v", val, 2)
			}
		} else if "objectThree" == key {
			if val := val.S("num").Data().(float64); val != 3 {
				t.Errorf("%v != %v", val, 3)
			}
		} else {
			t.Errorf("Unexpected key: %v", key)
		}
	}

	objectMap["objectOne"].Set(500, "num")
	if val := json1.Path("objectOne.num").Data().(int); val != 500 {
		t.Errorf("set objectOne failed: %v != %v", val, 500)
	}
}

func TestNestedAnonymousArrays(t *testing.T) {
	json1, _ := ParseJSON([]byte(`{
		"array":[
			[ 1, 2, 3 ],
			[ 4, 5, 6 ],
			[ 7, 8, 9 ],
			[{ "test" : 50 }]
		]
	}`))

	childTest := json1.S("array").Index(0).Children()

	if val := childTest[0].Data().(float64); val != 1 {
		t.Errorf("child test: %v != %v", val, 1)
	}
	if val := childTest[1].Data().(float64); val != 2 {
		t.Errorf("child test: %v != %v", val, 2)
	}
	if val := childTest[2].Data().(float64); val != 3 {
		t.Errorf("child test: %v != %v", val, 3)
	}

	if val := json1.Path("array").Index(1).Index(1).Data().(float64); val != 5 {
		t.Errorf("nested child test: %v != %v", val, 5)
	}

	if val := json1.Path("array").Index(3).Index(0).S("test").Data().(float64); val != 50 {
		t.Errorf("nested child object test: %v != %v", val, 50)
	}

	json1.Path("array").Index(3).Index(0).Set(200, "test")

	if val := json1.Path("array").Index(3).Index(0).S("test").Data().(int); val != 200 {
		t.Errorf("set nested child object: %v != %v", val, 200)
	}
}

func TestArrays(t *testing.T) {
	json1, _ := ParseJSON([]byte(`{
		"languages":{
			"english":{
				"places":0
			},
			"french": {
				"places": [
					"france",
					"belgium"
				]
			}
		}
	}`))

	json2, _ := ParseJSON([]byte(`{
		"places":[
			"great_britain",
			"united_states_of_america",
			"the_world"
		]
	}`))

	if englishPlaces := json2.Search("places").Data(); englishPlaces != nil {
		json1.Path("languages.english").Set(englishPlaces, "places")
	} else {
		t.Errorf("Didn't find places in json2")
	}

	if englishPlaces := json1.Search("languages", "english", "places").Data(); englishPlaces != nil {

		englishArray, ok := englishPlaces.([]interface{})
		if !ok {
			t.Errorf("places in json1 (%v) was not an array", englishPlaces)
		}

		if len(englishArray) != 3 {
			t.Errorf("wrong length of array: %v", len(englishArray))
		}

	} else {
		t.Errorf("Didn't find places in json1")
	}

	for i := 0; i < 3; i++ {
		if err := json2.ArrayRemove(0, "places"); err != nil {
			t.Errorf("Error removing element: %v", err)
		}
	}

	json2.ArrayAppend(map[string]interface{}{}, "places")
	json2.ArrayAppend(map[string]interface{}{}, "places")
	json2.ArrayAppend(map[string]interface{}{}, "places")

	// Using float64 for this test even though it's completely inappropriate because
	// later on the API might do something clever with types, in which case all numbers
	// will become float64.
	for i := 0; i < 3; i++ {
		obj, _ := json2.ArrayElement(i, "places")
		obj2, _ := obj.Object(fmt.Sprintf("object%v", i))
		obj2.Set(float64(i), "index")
	}

	children := json2.S("places").Children()
	for i, obj := range children {
		if id, ok := obj.S(fmt.Sprintf("object%v", i)).S("index").Data().(float64); ok {
			if id != float64(i) {
				t.Errorf("Wrong index somehow, expected %v, received %v", i, id)
			}
		} else {
			t.Errorf("Failed to find element %v from %v", i, obj)
		}
	}

	if err := json2.ArrayRemove(1, "places"); err != nil {
		t.Errorf("Error removing element: %v", err)
	}

	expected := `{"places":[{"object0":{"index":0}},{"object2":{"index":2}}]}`
	received := json2.String()

	if expected != received {
		t.Errorf("Wrong output, expected: %v, received: %v", expected, received)
	}
}

func TestArraysTwo(t *testing.T) {
	json1 := New()

	test1, err := json1.ArrayOfSize(4, "test1")
	if err != nil {
		t.Error(err)
	}

	if _, err = test1.ArrayOfSizeI(2, 0); err != nil {
		t.Error(err)
	}
	if _, err = test1.ArrayOfSizeI(2, 1); err != nil {
		t.Error(err)
	}
	if _, err = test1.ArrayOfSizeI(2, 2); err != nil {
		t.Error(err)
	}
	if _, err = test1.ArrayOfSizeI(2, 3); err != nil {
		t.Error(err)
	}

	if _, err = test1.ArrayOfSizeI(2, 4); err != ErrOutOfBounds {
		t.Errorf("Index should have been out of bounds")
	}

	if _, err = json1.S("test1").Index(0).SetIndex(10, 0); err != nil {
		t.Error(err)
	}
	if _, err = json1.S("test1").Index(0).SetIndex(11, 1); err != nil {
		t.Error(err)
	}

	if _, err = json1.S("test1").Index(1).SetIndex(12, 0); err != nil {
		t.Error(err)
	}
	if _, err = json1.S("test1").Index(1).SetIndex(13, 1); err != nil {
		t.Error(err)
	}

	if _, err = json1.S("test1").Index(2).SetIndex(14, 0); err != nil {
		t.Error(err)
	}
	if _, err = json1.S("test1").Index(2).SetIndex(15, 1); err != nil {
		t.Error(err)
	}

	if _, err = json1.S("test1").Index(3).SetIndex(16, 0); err != nil {
		t.Error(err)
	}
	if _, err = json1.S("test1").Index(3).SetIndex(17, 1); err != nil {
		t.Error(err)
	}

	if val := json1.S("test1").Index(0).Index(0).Data().(int); val != 10 {
		t.Errorf("create array: %v != %v", val, 10)
	}
	if val := json1.S("test1").Index(0).Index(1).Data().(int); val != 11 {
		t.Errorf("create array: %v != %v", val, 11)
	}

	if val := json1.S("test1").Index(1).Index(0).Data().(int); val != 12 {
		t.Errorf("create array: %v != %v", val, 12)
	}
	if val := json1.S("test1").Index(1).Index(1).Data().(int); val != 13 {
		t.Errorf("create array: %v != %v", val, 13)
	}

	if val := json1.S("test1").Index(2).Index(0).Data().(int); val != 14 {
		t.Errorf("create array: %v != %v", val, 14)
	}
	if val := json1.S("test1").Index(2).Index(1).Data().(int); val != 15 {
		t.Errorf("create array: %v != %v", val, 15)
	}

	if val := json1.S("test1").Index(3).Index(0).Data().(int); val != 16 {
		t.Errorf("create array: %v != %v", val, 16)
	}
	if val := json1.S("test1").Index(3).Index(1).Data().(int); val != 17 {
		t.Errorf("create array: %v != %v", val, 17)
	}
}

func TestArraysThree(t *testing.T) {
	json1 := New()

	test, err := json1.ArrayOfSizeP(1, "test1.test2")
	if err != nil {
		t.Fatal(err)
	}

	test.SetIndex(10, 0)
	if val := json1.S("test1", "test2").Index(0).Data().(int); val != 10 {
		t.Error(err)
	}
}

func TestSetJSONPointer(t *testing.T) {
	type testCase struct {
		input   string
		pointer string
		value   interface{}
		output  string
	}
	tests := []testCase{
		{
			input:   `{"foo":{"bar":"baz"}}`,
			pointer: "/foo/bar",
			value:   "qux",
			output:  `{"foo":{"bar":"qux"}}`,
		},
		{
			input:   `{"foo":["bar","ignored","baz"]}`,
			pointer: "/foo/2",
			value:   "qux",
			output:  `{"foo":["bar","ignored","qux"]}`,
		},
		{
			input:   `{"foo":["bar","ignored",{"bar":"baz"}]}`,
			pointer: "/foo/2/bar",
			value:   "qux",
			output:  `{"foo":["bar","ignored",{"bar":"qux"}]}`,
		},
	}

	for _, test := range tests {
		gObj, err := ParseJSON([]byte(test.input))
		if err != nil {
			t.Errorf("Failed to parse '%v': %v", test.input, err)
			continue
		}
		if _, err = gObj.SetJSONPointer(test.value, test.pointer); err != nil {
			t.Error(err)
			continue
		}
		if exp, act := test.output, gObj.String(); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestArrayReplace(t *testing.T) {
	json1 := New()

	json1.Set(1, "first")
	json1.ArrayAppend(2, "first")
	json1.ArrayAppend(3, "second")

	expected := `{"first":[1,2],"second":[3]}`
	received := json1.String()

	if expected != received {
		t.Errorf("Wrong output, expected: %v, received: %v", expected, received)
	}
}

func TestArraysRoot(t *testing.T) {
	sample := []byte(`["test1"]`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	val.ArrayAppend("test2")
	val.ArrayAppend("test3")
	if obj, err := val.ObjectI(2); err != nil {
		t.Error(err)
	} else {
		obj.Set("bar", "foo")
	}

	if expected, actual := `["test1","test2",{"foo":"bar"}]`, val.String(); expected != actual {
		t.Errorf("expected %v, received: %v", expected, actual)
	}
}

func TestLargeSample(t *testing.T) {
	sample := []byte(`{
		"test":{
			"innerTest":{
				"value":10,
				"value2":22,
				"value3":{
					"moreValue":45
				}
			}
		},
		"test2":20
	}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	if result, ok := val.Search("test", "innerTest", "value3", "moreValue").Data().(float64); ok {
		if result != 45 {
			t.Errorf("Wrong value of result: %v", result)
		}
	} else {
		t.Errorf("Didn't find value")
	}
}

func TestShorthand(t *testing.T) {
	json, _ := ParseJSON([]byte(`{
		"outter":{
			"inner":{
				"value":5,
				"value2":10,
				"value3":11
			},
			"inner2":{
			}
		},
		"outter2":{
			"inner":0
		}
	}`))

	missingValue := json.S("outter").S("doesntexist").S("alsodoesntexist").S("inner").S("value").Data()
	if missingValue != nil {
		t.Errorf("missing value was actually found: %v\n", missingValue)
	}

	realValue := json.S("outter").S("inner").S("value2").Data().(float64)
	if realValue != 10 {
		t.Errorf("real value was incorrect: %v\n", realValue)
	}

	_, err := json.S("outter2").Set(json.S("outter").S("inner").Data(), "inner")
	if err != nil {
		t.Errorf("error setting outter2: %v\n", err)
	}

	compare := `{"outter":{"inner":{"value":5,"value2":10,"value3":11},"inner2":{}}` +
		`,"outter2":{"inner":{"value":5,"value2":10,"value3":11}}}`
	out := json.String()
	if out != compare {
		t.Errorf("wrong serialized structure: %v\n", out)
	}

	compare2 := `{"outter":{"inner":{"value":6,"value2":10,"value3":11},"inner2":{}}` +
		`,"outter2":{"inner":{"value":6,"value2":10,"value3":11}}}`

	json.S("outter").S("inner").Set(6, "value")
	out = json.String()
	if out != compare2 {
		t.Errorf("wrong serialized structure: %v\n", out)
	}
}

func TestInvalid(t *testing.T) {
	invalidJSONSamples := []string{
		`{dfads"`,
		``,
		// `""`,
		// `"hello"`,
		"{}\n{}",
	}

	for _, sample := range invalidJSONSamples {
		if _, err := ParseJSON([]byte(sample)); err == nil {
			t.Errorf("parsing invalid JSON '%v' did not return error", sample)
		}
	}

	if _, err := ParseJSON(nil); err == nil {
		t.Errorf("parsing nil did not return error")
	}

	validObj, err := ParseJSON([]byte(`{}`))
	if err != nil {
		t.Errorf("failed to parse '{}'")
	}

	invalidStr := validObj.S("Doesn't exist").String()
	if "null" != invalidStr {
		t.Errorf("expected 'null', received: %v", invalidStr)
	}
}

func TestCreation(t *testing.T) {
	json, _ := ParseJSON([]byte(`{}`))
	inner, err := json.ObjectP("test.inner")
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	inner.Set(10, "first")
	inner.Set(20, "second")

	inner.Array("array")
	inner.ArrayAppend("first element of the array", "array")
	inner.ArrayAppend(2, "array")
	inner.ArrayAppend("three", "array")

	expected := `{"test":{"inner":{"array":["first element of the array",2,"three"],` +
		`"first":10,"second":20}}}`
	actual := json.String()
	if actual != expected {
		t.Errorf("received incorrect output from json object: %v\n", actual)
	}
}

type outterJSON struct {
	FirstInner  innerJSON
	SecondInner innerJSON
	ThirdInner  innerJSON
}

type innerJSON struct {
	NumberType float64
	StringType string
}

type jsonStructure struct {
	FirstOutter  outterJSON
	SecondOutter outterJSON
}

var jsonContent = []byte(`{
	"firstOutter":{
		"firstInner":{
			"numberType":11,
			"stringType":"hello world, first first"
		},
		"secondInner":{
			"numberType":12,
			"stringType":"hello world, first second"
		},
		"thirdInner":{
			"numberType":13,
			"stringType":"hello world, first third"
		}
	},
	"secondOutter":{
		"firstInner":{
			"numberType":21,
			"stringType":"hello world, second first"
		},
		"secondInner":{
			"numberType":22,
			"stringType":"hello world, second second"
		},
		"thirdInner":{
			"numberType":23,
			"stringType":"hello world, second third"
		}
	}
}`)

/*
Simple use case, compares unmarshalling declared structs vs dynamically searching for
the equivalent hierarchy. Hopefully we won't see too great a performance drop from the
dynamic approach.
*/

func BenchmarkStatic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var jsonObj jsonStructure
		json.Unmarshal(jsonContent, &jsonObj)

		if val := jsonObj.FirstOutter.SecondInner.NumberType; val != 12 {
			b.Errorf("Wrong value of FirstOutter.SecondInner.NumberType: %v\n", val)
		}
		expected := "hello world, first second"
		if val := jsonObj.FirstOutter.SecondInner.StringType; val != expected {
			b.Errorf("Wrong value of FirstOutter.SecondInner.StringType: %v\n", val)
		}
		if val := jsonObj.SecondOutter.ThirdInner.NumberType; val != 23 {
			b.Errorf("Wrong value of SecondOutter.ThirdInner.NumberType: %v\n", val)
		}
		expected = "hello world, second second"
		if val := jsonObj.SecondOutter.SecondInner.StringType; val != expected {
			b.Errorf("Wrong value of SecondOutter.SecondInner.StringType: %v\n", val)
		}
	}
}

func BenchmarkDynamic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jsonObj, err := ParseJSON(jsonContent)
		if err != nil {
			b.Errorf("Error parsing json: %v\n", err)
		}

		FOSI := jsonObj.S("firstOutter", "secondInner")
		SOSI := jsonObj.S("secondOutter", "secondInner")
		SOTI := jsonObj.S("secondOutter", "thirdInner")

		if val := FOSI.S("numberType").Data().(float64); val != 12 {
			b.Errorf("Wrong value of FirstOutter.SecondInner.NumberType: %v\n", val)
		}
		expected := "hello world, first second"
		if val := FOSI.S("stringType").Data().(string); val != expected {
			b.Errorf("Wrong value of FirstOutter.SecondInner.StringType: %v\n", val)
		}
		if val := SOTI.S("numberType").Data().(float64); val != 23 {
			b.Errorf("Wrong value of SecondOutter.ThirdInner.NumberType: %v\n", val)
		}
		expected = "hello world, second second"
		if val := SOSI.S("stringType").Data().(string); val != expected {
			b.Errorf("Wrong value of SecondOutter.SecondInner.StringType: %v\n", val)
		}
	}
}

func TestBadIndexes(t *testing.T) {
	jsonObj, err := ParseJSON([]byte(`{"array":[1,2,3]}`))
	if err != nil {
		t.Error(err)
	}
	if act := jsonObj.Index(0).Data(); nil != act {
		t.Errorf("Unexpected value returned: %v != %v", nil, act)
	}
	if act := jsonObj.S("array").Index(4).Data(); nil != act {
		t.Errorf("Unexpected value returned: %v != %v", nil, act)
	}
}

func TestNilSet(t *testing.T) {
	obj := Container{nil}
	if _, err := obj.Set("bar", "foo"); err != nil {
		t.Error(err)
	}
	if _, err := obj.Set("new", "foo", "bar"); err != ErrPathCollision {
		t.Errorf("Expected ErrPathCollision: %v, %s", err, obj.Data())
	}
	if _, err := obj.SetIndex("new", 0); err != ErrNotArray {
		t.Errorf("Expected ErrNotArray: %v, %s", err, obj.Data())
	}
}

func TestLargeSampleWithHtmlEscape(t *testing.T) {
	sample := []byte(`{
	"test": {
		"innerTest": {
			"value": 10,
			"value2": "<title>Title</title>",
			"value3": {
				"moreValue": 45
			}
		}
	},
	"test2": 20
}`)

	sampleWithHTMLEscape := []byte(`{
	"test": {
		"innerTest": {
			"value": 10,
			"value2": "\u003ctitle\u003eTitle\u003c/title\u003e",
			"value3": {
				"moreValue": 45
			}
		}
	},
	"test2": 20
}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	exp := string(sample)
	res := string(val.EncodeJSON(EncodeOptIndent("", "\t")))
	if exp != res {
		t.Errorf("Wrong conversion without html escaping: %s != %s", res, exp)
	}

	exp = string(sampleWithHTMLEscape)
	res = string(val.EncodeJSON(EncodeOptHTMLEscape(true), EncodeOptIndent("", "\t")))
	if exp != res {
		t.Errorf("Wrong conversion with html escaping: %s != %s", exp, res)
	}
}

func TestMergeCases(t *testing.T) {
	type testCase struct {
		first    string
		second   string
		expected string
	}

	testCases := []testCase{
		{
			first:    `{"outter":{"value1":"one"}}`,
			second:   `{"outter":{"inner":{"value3": "threre"}},"outter2":{"value2": "two"}}`,
			expected: `{"outter":{"inner":{"value3":"threre"},"value1":"one"},"outter2":{"value2":"two"}}`,
		},
		{
			first:    `{"outter":["first"]}`,
			second:   `{"outter":["second"]}`,
			expected: `{"outter":["first","second"]}`,
		},
		{
			first:    `{"outter":["first",{"inner":"second"}]}`,
			second:   `{"outter":["third"]}`,
			expected: `{"outter":["first",{"inner":"second"},"third"]}`,
		},
		{
			first:    `{"outter":["first",{"inner":"second"}]}`,
			second:   `{"outter":"third"}`,
			expected: `{"outter":["first",{"inner":"second"},"third"]}`,
		},
		{
			first:    `{"outter":"first"}`,
			second:   `{"outter":"second"}`,
			expected: `{"outter":["first","second"]}`,
		},
		{
			first:    `{"outter":{"inner":"first"}}`,
			second:   `{"outter":{"inner":"second"}}`,
			expected: `{"outter":{"inner":["first","second"]}}`,
		},
		{
			first:    `{"outter":{"inner":"first"}}`,
			second:   `{"outter":"second"}`,
			expected: `{"outter":[{"inner":"first"},"second"]}`,
		},
		{
			first:    `{"outter":{"inner":"second"}}`,
			second:   `{"outter":{"inner":{"inner2":"first"}}}`,
			expected: `{"outter":{"inner":["second",{"inner2":"first"}]}}`,
		},
		{
			first:    `{"outter":{"inner":["second"]}}`,
			second:   `{"outter":{"inner":{"inner2":"first"}}}`,
			expected: `{"outter":{"inner":["second",{"inner2":"first"}]}}`,
		},
		{
			first:    `{"outter":"second"}`,
			second:   `{"outter":{"inner":"first"}}`,
			expected: `{"outter":["second",{"inner":"first"}]}`,
		},
		{
			first:    `{"outter":"first"}`,
			second:   `{"outter":["second"]}`,
			expected: `{"outter":["first","second"]}`,
		},
	}

	for i, test := range testCases {
		var firstContainer, secondContainer *Container
		var err error

		firstContainer, err = ParseJSON([]byte(test.first))
		if err != nil {
			t.Errorf("[%d] Failed to parse '%v': %v", i, test.first, err)
		}

		secondContainer, err = ParseJSON([]byte(test.second))
		if err != nil {
			t.Errorf("[%d] Failed to parse '%v': %v", i, test.second, err)
		}

		if err = firstContainer.Merge(secondContainer); err != nil {
			t.Errorf("[%d] Failed to merge: '%v': %v", i, test.first, err)
		}

		if exp, act := test.expected, firstContainer.String(); exp != act {
			t.Errorf("[%d] Wrong result: %v != %v", i, act, exp)
		}
	}
}

func TestMarshalsJSON(t *testing.T) {
	sample := []byte(`{"test":{"value":10},"test2":20}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Fatal(err)
	}

	marshaled, err := json.Marshal(val)
	if err != nil {
		t.Fatal(err)
	}

	if exp, act := string(sample), string(marshaled); exp != act {
		t.Errorf("Unexpected result: %v != %v", act, exp)
	}
}

func TestFlatten(t *testing.T) {
	type testCase struct {
		input  string
		output string
	}
	tests := []testCase{
		{
			input:  `{"foo":{"bar":"baz"}}`,
			output: `{"foo.bar":"baz"}`,
		},
		{
			input:  `{"foo":[{"bar":"1"},{"bar":"2"}]}`,
			output: `{"foo.0.bar":"1","foo.1.bar":"2"}`,
		},
		{
			input:  `[{"bar":"1"},{"bar":"2"}]`,
			output: `{"0.bar":"1","1.bar":"2"}`,
		},
		{
			input:  `[["1"],["2","3"]]`,
			output: `{"0.0":"1","1.0":"2","1.1":"3"}`,
		},
		{
			input:  `{"foo":{"bar":null}}`,
			output: `{"foo.bar":null}`,
		},
		{
			input:  `{"foo":{"bar":{}}}`,
			output: `{}`,
		},
		{
			input:  `{"foo":{"bar":[]}}`,
			output: `{}`,
		},
	}

	for _, test := range tests {
		gObj, err := ParseJSON([]byte(test.input))
		if err != nil {
			t.Fatalf("Failed to parse '%v': %v", test.input, err)
		}
		var res map[string]interface{}
		if res, err = gObj.Flatten(); err != nil {
			t.Error(err)
			continue
		}
		if exp, act := test.output, Wrap(res).String(); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestFlattenIncludeEmpty(t *testing.T) {
	type testCase struct {
		input  string
		output string
	}
	tests := []testCase{
		{
			input:  `{"foo":{"bar":"baz"}}`,
			output: `{"foo.bar":"baz"}`,
		},
		{
			input:  `{"foo":[{"bar":"1"},{"bar":"2"}]}`,
			output: `{"foo.0.bar":"1","foo.1.bar":"2"}`,
		},
		{
			input:  `[{"bar":"1"},{"bar":"2"}]`,
			output: `{"0.bar":"1","1.bar":"2"}`,
		},
		{
			input:  `[["1"],["2","3"]]`,
			output: `{"0.0":"1","1.0":"2","1.1":"3"}`,
		},
		{
			input:  `{"foo":{"bar":null}}`,
			output: `{"foo.bar":null}`,
		},
		{
			input:  `{"foo":{"bar":{}}}`,
			output: `{"foo.bar":{}}`,
		},
		{
			input:  `{"foo":{"bar":[]}}`,
			output: `{"foo.bar":[]}`,
		},
	}

	for _, test := range tests {
		gObj, err := ParseJSON([]byte(test.input))
		if err != nil {
			t.Fatalf("Failed to parse '%v': %v", test.input, err)
		}
		var res map[string]interface{}
		if res, err = gObj.FlattenIncludeEmpty(); err != nil {
			t.Error(err)
			continue
		}
		if exp, act := test.output, Wrap(res).String(); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}
