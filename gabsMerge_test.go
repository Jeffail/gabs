package gabs_test

import (
	"testing"

	"github.com/TobiEiss/gabs"
)

func TestMergeSimpleJsons(t *testing.T) {
	jsonParsed1, err := gabs.ParseJSON([]byte(`{"value1": "one"}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	jsonParsed2, err := gabs.ParseJSON([]byte(`{"value2": "two"}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	err = jsonParsed1.Merge(jsonParsed2)
	if err != nil {
		t.Errorf("Failed to merge: %v", err)
	}

	if !jsonParsed1.Exists("value2") {
		t.Fail()
	}
}

func TestMergeSimpleJsonsWithFailure(t *testing.T) {
	jsonParsed1, err := gabs.ParseJSON([]byte(`{"value1": "one"}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	jsonParsed2, err := gabs.ParseJSON([]byte(`{"value1": "two"}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	err = jsonParsed1.Merge(jsonParsed2)
	if err == nil {
		t.Errorf("failure is expected!")
	}

}

func TestMergeSimpleJsonsWithMultipleValues(t *testing.T) {
	jsonParsed1, err := gabs.ParseJSON([]byte(`{"value1": "one"}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	jsonParsed2, err := gabs.ParseJSON([]byte(`{"value2": "two", "value3": "three"}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	err = jsonParsed1.Merge(jsonParsed2)
	if err != nil {
		t.Errorf("Failed to merge: %v", err)
	}

	if !jsonParsed1.Exists("value2") || !jsonParsed1.Exists("value3") {
		t.Fail()
	}
}
