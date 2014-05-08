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

	if err := val.S("test").Set("value", 45.0); err != nil {
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
		object.Set("child", "hello world")
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

	objects, _ = json2.S("values").Children()
	for _, object := range objects {
		object.Set("child", "hello world")
		json3.Push("values", object.Data())
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
		json1.Path("languages.english").Set("places", englishPlaces)
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

	json2.Push("places", map[string]interface{}{})
	json2.Push("places", map[string]interface{}{})
	json2.Push("places", map[string]interface{}{})

	// Using float64 for this test even though it's completely inappropriate because
	// later on the API might do something clever with types, in which case all numbers
	// will become float64.
	for i := 0; i < 3; i++ {
		json2.GetElement("places", i).CreateObject(fmt.Sprintf("object%v", i)).Set("index", float64(i))
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

	if err := json2.RemoveElement("places", 1); err != nil {
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

	err := json.S("outter2").Set("inner", json.S("outter").S("inner").Data())
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

	json.S("outter").S("inner").Set("value", 6)
	out = json.String()
	if out != compare2 {
		t.Errorf("wrong serialized structure: %v\n", out)
	}
}

func TestInvalid(t *testing.T) {
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
	inner := json.CO("test").CO("inner")

	inner.Set("first", 10)
	inner.Set("second", 20)

	inner.CA("array")
	inner.Push("array", "first element of the array")
	inner.Push("array", 2)
	inner.Push("array", "three")

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
