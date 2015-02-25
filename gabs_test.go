package gabs

import (
	"encoding/json"
	"fmt"
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
}

func TestFindArray(t *testing.T) {
	sample := []byte(`{"test":{"array":[{"value":1}, {"value":2}, {"value":3}]}}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	target := val.Path("test.array.value")
	expected := "[1,2,3]"
	result := target.String()

	if expected != result {
		t.Errorf("Expected %v, received %v", expected, result)
	}
}

func TestFindArray2(t *testing.T) {
	sample := []byte(`{
		"test":{
			"array":[
				{
					"values":[
						{"more":1},
						{"more":2},
						{"more":3}
					]
				},
				{
					"values":[
						{"more":4},
						{"more":5},
						{"more":6}
					]
				},
				{
					"values":[
						{"more":7},
						{"more":8},
						{"more":9}
					]
				}
			]
		}
	}`)

	val, err := ParseJSON(sample)
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
		return
	}

	target := val.Path("test.array.values.more")
	expected := "[[1,2,3],[4,5,6],[7,8,9]]"
	result := target.String()

	if expected != result {
		t.Errorf("Expected %v, received %v", expected, result)
	}
}

func TestExamples(t *testing.T) {
	jsonParsed, _ := ParseJSON([]byte(`{
		"outter":{
			"inner":{
				"value1":10,
				"value2":22
			},
			"alsoInner":{
				"value1":20
			}
		}
	}`))

	var value float64
	var ok bool

	value, ok = jsonParsed.Path("outter.inner.value1").Data().(float64)
	if value != 10.0 || !ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	value, ok = jsonParsed.Search("outter", "inner", "value1").Data().(float64)
	if value != 10.0 || !ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	value, ok = jsonParsed.Path("does.not.exist").Data().(float64)
	if value != 0.0 || ok {
		t.Errorf("wrong value: %v, %v", value, ok)
	}

	jsonParsed, _ = ParseJSON([]byte(`{"array":[ "first", "second", "third" ]}`))

	expected := []string{"first", "second", "third"}

	children, err := jsonParsed.S("array").Children()
	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}
	for i, child := range children {
		if expected[i] != child.Data().(string) {
			t.Errorf("Child unexpected: %v != %v", expected[i], child.Data().(string))
		}
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

	jsonObj, _ = Consume(map[string]interface{}{})

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

	jsonObj.Array("foo", "array")

	jsonObj.ArrayAppend(10, "foo", "array")
	jsonObj.ArrayAppend(20, "foo", "array")
	jsonObj.ArrayAppend(30, "foo", "array")

	result := jsonObj.String()
	expected := `{"foo":{"array":[10,20,30]}}`

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

	objects, _ := json1.Children()
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

	objects, _ = json2.S("values").Children()
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
		if err := json2.RemoveElement("places", 0); err != nil {
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

	children, _ := json2.S("places").Children()
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
		`""`,
		`"hello"`,
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
	if "{}" != invalidStr {
		t.Errorf("expected '{}', received: %v", invalidStr)
	}
}

func TestCreation(t *testing.T) {
	json, _ := ParseJSON([]byte(`{}`))
	inner, err := json.Object("test", "inner")
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
