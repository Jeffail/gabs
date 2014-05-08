![Gabs](http://www.creepybit.co.uk/images/gabs_logo.png?v=2 "Gabs")

Gabs is a small utility for dealing with dynamic or unknown JSON structures in golang. It's pretty much just a helpful wrapper around the golang json.Marshal/json.Unmarshal behaviour and map[string]interface{} objects. It does nothing spectacular except for being fabulous.

https://godoc.org/github.com/Jeffail/gabs

##How to install:

```bash
go get github.com/jeffail/gabs
```

##How to use

###Parsing JSON

```go
...

import "github.com/jeffail/gabs"

jsonParsed, err := gabs.ParseJSON([]byte(`{
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

if err != nil {
	// You done goofed
}

value, ok := jsonParsed.Path("outter.inner.value1").Data().(float64)
if ok {
	// outter.inner.value1 was found and its value is now stored in valueOne.
} else {
	// outter.inner.value1 was either non-existant in the JSON structure or
	// was of a different type.
}

// Alternatively, break the search down into individual strings
value, ok = jsonParsed.Search("outter", "inner", "value1").Data().(float64)

// S() is shorthand for Search()
value, ok = jsonParsed.S("outter", "inner", "value1").Data().(float64)

if err := jsonParsed.Path("outter.inner").Set("value2", 10f); err == nil {
	// outter.inner.value2 has been set to 10.
}

jsonParsed2, _ := gabs.ParseJSON([]byte(`{"array":[]"}`))

for _, object := range jsonParsed.S("outter").Children() {
	// Do something with object
}

// And there are helper functions for modifying and receiving array values

if err := jsonParsed2.RemoveElement("array", 1); err != nil {
	// Index was out of bounds or the array doesn't exist
}

value, ok = jsonParsed2.GetElement("array", 0).Path("value1").Data().(float64)
// value will either be 10 or 20, we don't know because object children
// aren't iterated in order

...
```

All search and path queries return a container of the underlying JSON object. If the object doesn't exist you will still receive a valid container with an underlying value of nil. Calling Data() returns this underlying value, which you can then attempt to cast in order to validate the value was found and is the expected type.

You can set the value of a child of an object with Set. If the child doesn't already exist it is created, and an error is returned if the containing object either doesn't exist or isn't of the type map[string]interface{} (a JSON object).

NOTE: The Set method accepts interface{}, so this can potentially be any type and will be serialized following the same rules as json.Marshal. Gabs currently doesn't do anything clever with these values, so don't expect to Set using integer values and then receive back float64's.

Gabs tries to make building a JSON structure dynamically as easy as parsing it.

```go
...

json, _ := gabs.ParseJSON([]byte(`{}`))

// CreateObject creates a new JSON object, and also returns it as a gabs container.
// CO("") is shorthand for CreateObject("")
inner := json.CreateObject("test").CO("inner")

inner.Set("first", 10f)
inner.Set("second", 20f)

// CreateArray creates a new JSON array.
// CA("") is shorthand for CreateArray("").
inner.CreateArray("array")

// Push pushes new values onto an existing JSON array.
inner.Push("array", "one")
inner.Push("array", 2)
inner.Push("array", "three")

fmt.Println(json.String())
// This should display:
// `{"test":{"inner":{"array":["one",2,"three"],"first":10,"second":20}}}`

...
```

Doing things like merging different JSON structures is also fairly simple.

```go
...

import "github.com/jeffail/gabs"

json1, _ := gabs.ParseJSON([]byte(`{
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

json2, _ := gabs.ParseJSON([]byte(`{
	"places":[
		"great_britain",
		"united_states_of_america",
		"the_world"
	]
}`))

// The following

if englishPlaces := json2.Search("places").Data(); englishPlaces != nil {
	json1.Path("languages.english").Set("places", englishPlaces)
}

// Could also be written as

if englishPlaces = json2.Path("places").Data(); englishPlaces != nil {
	json1.S("languages", "english").Set("places", englishPlaces)
}

// NOTE: The internal structure of json1 now contains a pointer to the structure
// within json2, so editing json2 will also effect json1. This behaviour also means
// that the structure can contain circular references if you aren't careful.

/* If all went well then the structure of json1 should now be:
	"languages":{
		"english":{
			"places":[
				"great_britain",
				"united_states_of_america",
				"the_world"
			]
		},
		"french": {
			"places": [
				"france",
				"belgium"
			]
		}
*/

...
```

###Converting back to JSON

This is the easiest part:

```go
...

jsonParsedObj := gabs.ParseJSON([]byte(`{
	"outter":{
		"values":{
			"first":10,
			"second":11
		}
	},
	"outter2":"hello world"
}`))

jsonOutput := jsonParsedObj.String()
// Becomes `{"outter":{"values":{"first":10,"second":11}},"outter2":"hello world"}`

...
```

And to serialize a specific segment is as simple as:

```go
...

jsonParsedObj := gabs.ParseJSON([]byte(`{
	"outter":{
		"values":{
			"first":10,
			"second":11
		}
	},
	"outter2":"hello world"
}`))

jsonOutput := jsonParsedObj.Search("outter").String()
// Becomes `{"values":{"first":10,"second":11}}`

// If, however, "outter" was not found, or the container was invalid,
// String() returns "{}"

...
```
