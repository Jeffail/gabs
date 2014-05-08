/*
Copyright (c) 2014 Ashley Jeffs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package gabs implements a simplified wrapper around json parsing an unknown structure
package gabs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
)

/*
Container - an internal structure that holds a reference to the core interface map of the parsed
json. Use this container to move context.
*/
type Container struct {
	object interface{}
}

/*
Path - Search for a value using dot notation.
*/
func (g *Container) Path(path string) *Container {
	return g.Search(strings.Split(path, ".")...)
}

/*
Search - Attempt to find and return an object within the JSON structure by specifying the hierarchy of
field names to locate the target.
*/
func (g *Container) Search(hierarchy ...string) *Container {
	var object interface{}

	object = g.object
	for target := 0; target < len(hierarchy); target++ {
		if mmap, ok := object.(map[string]interface{}); ok {
			object = mmap[hierarchy[target]]
		} else {
			return &Container{nil}
		}
	}

	return &Container{object}
}

/*
S - Shorthand method, does the same thing as Search.
*/
func (g *Container) S(hierarchy ...string) *Container {
	return g.Search(hierarchy...)
}

/*
Data - Return the contained data as an interface{}.
*/
func (g *Container) Data() interface{} {
	return g.object
}

/*
Children - Return a slice of all the children of the array. This also works for objects,
however, the children returned for an object will NOT be in order and you lose the names
of the returned objects this way.
*/
func (g *Container) Children() ([]*Container, error) {
	if array, ok := g.Data().([]interface{}); ok {

		children := make([]*Container, len(array))
		for i := 0; i < len(array); i++ {
			children[i] = &Container{array[i]}
		}

		return children, nil

	} else if mmap, ok := g.Data().(map[string]interface{}); ok {

		children := []*Container{}
		for _, obj := range mmap {
			children = append(children, &Container{obj})
		}

		return children, nil
	}

	return nil, errors.New("parent was not a valid JSON object or array")
}

/*
Set - Set the value for a child of a JSON object. The child doesn't have to already exist.
*/
func (g *Container) Set(target string, value interface{}) error {
	if mmap, ok := g.Data().(map[string]interface{}); ok {
		mmap[target] = value
	} else {
		return errors.New("parent was not a valid JSON object")
	}

	return nil
}

/*
Push - Push a value onto a JSON array.
*/
func (g *Container) Push(target string, value interface{}) error {
	if mmap, ok := g.Data().(map[string]interface{}); ok {
		arrayTarget := mmap[target]
		if array, ok := arrayTarget.([]interface{}); ok {
			mmap[target] = append(array, value)
		} else {
			return errors.New("target object was not an array")
		}
	} else {
		return errors.New("parent was not a valid JSON object")
	}

	return nil
}

/*
CreateObject - Create a new JSON object. Returns a container of the new object.
*/
func (g *Container) CreateObject(name string) *Container {
	if mmap, ok := g.Data().(map[string]interface{}); ok {
		mmap[name] = map[string]interface{}{}
		return &Container{mmap[name]}
	}

	return &Container{nil}
}

/*
CO - Shorthand method for CreateObject.
*/
func (g *Container) CO(name string) *Container {
	return g.CreateObject(name)
}

/*
CreateArray - Create a new JSON array.
*/
func (g *Container) CreateArray(name string) error {
	if mmap, ok := g.Data().(map[string]interface{}); ok {
		mmap[name] = []interface{}{}
		return nil
	}
	return errors.New("container was not a valid object")
}

/*
CA - Shorthand method for CreateArray.
*/
func (g *Container) CA(name string) error {
	return g.CreateArray(name)
}

/*
String - Converts the contained object back to a JSON formatted string.
*/
func (g *Container) String() string {
	if g.object != nil {
		if bytes, err := json.Marshal(g.object); err == nil {
			return string(bytes)
		}
	}

	return "{}"
}

/*
Consume - Gobble up an already converted JSON object.
*/
func Consume(root interface{}) (*Container, error) {
	if _, ok := root.(map[string]interface{}); ok {
		return &Container{root}, nil
	}
	return nil, errors.New("root was not a valid JSON object")
}

/*
ParseJSON - Convert a string into a representation of the parsed JSON.
*/
func ParseJSON(sample []byte) (*Container, error) {
	var gabs Container

	if err := json.Unmarshal(sample, &gabs.object); err != nil {
		return nil, err
	}

	if _, ok := gabs.object.(map[string]interface{}); ok {
		return &gabs, nil
	}

	return nil, errors.New("json appears to contain no data")
}

/*
ParseJSONFile - Read a file and convert into a representation of the parsed JSON.
*/
func ParseJSONFile(path string) (*Container, error) {
	if len(path) > 0 {
		cBytes, err := ioutil.ReadFile(path)
		if err != nil {
			container, err := ParseJSON(cBytes)
			if err != nil {
				return container, nil
			}
			return nil, err
		}
		return nil, err
	}
	return nil, errors.New("file path was invalid")
}
