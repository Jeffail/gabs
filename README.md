![#Gabs](http://www.creepybit.co.uk/images/gabs_logo.png "Gabs")

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

jsonParsed, err := gabs.ParseJson([]byte(`{
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

/* Search returns an object of the same type as jsonParsed which should contain the target
 * data. Data returns the interface{} wrapped target object, it's then safe to attempt to cast
 * this object in order to determine whether the search obtained what you expected.
 *
 * These calls will accept a zero or greater number of string args.
 */
if valueOne, ok := jsonParsed.Search("outter", "inner", "value1").Data().(float64); ok {
	// outter.inner.value1 was found and its value is now stored in valueOne.
} else {
	// outter.inner.value1 was either non-existant in the JSON structure or was of a different type.
}

if err := jsonParsed.Set(10, "outter", "inner", "value2"); err == nil {
	// outter.inner.value2 was found and has been set to 10.
} else {
	// outter.inner.value2 was not found in the JSON structure.
}

...
```

Doing things like merging different JSON structures is also fairly simple.

```go
...

import "github.com/jeffail/gabs"

json1, _ := gabs.ParseJson([]byte(`{
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

json2, _ := gabs.ParseJson([]byte(`{
	"places":[
		"great_britain",
		"united_states_of_america",
		"the_world"
	]
}`))

if english_places := json2.Search("places").Data(); english_places != nil {
	json1.Set(english_places, "languages", "english", "places")
}

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

jsonParsedObj := gabs.ParseJson([]byte(`{
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

jsonParsedObj := gabs.ParseJson([]byte(`{
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

// If, however, "outter" was not found, or the container was invalid, String() returns "{}"

...
```
