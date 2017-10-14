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

func TestMergeJsonWithInner(t *testing.T) {
	jsonParsed1, err := gabs.ParseJSON([]byte(`{"outter": {"value1": "one"}}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	jsonParsed2, err := gabs.ParseJSON([]byte(`{"outter": {"value2": "two"}}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	err = jsonParsed1.Merge(jsonParsed2)
	if !jsonParsed1.Exists("outter", "value2") {
		t.Errorf("outter.value2 is missing")
	}
}

func TestMergeJsonWithComplexInner(t *testing.T) {
	jsonParsed1, err := gabs.ParseJSON([]byte(`{"outter": {"value1": "one"}}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	jsonParsed2, err := gabs.ParseJSON([]byte(`{"outter": {"value2": "two", "inner": {"value3": "threre"}}}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	err = jsonParsed1.Merge(jsonParsed2)
	if !jsonParsed1.Exists("outter", "value2") || !jsonParsed1.Exists("outter", "inner", "value3") {
		t.Errorf("outter.value2 is missing")
	}
}

func TestMergeJsonWithComplexerInner(t *testing.T) {
	jsonParsed1, err := gabs.ParseJSON([]byte(`{"outter": {"value1": "one"}}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	jsonParsed2, err := gabs.ParseJSON([]byte(`{"outter": {"inner": {"value3": "threre"}}, "outter2": {"value2": "two"}}`))
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}

	err = jsonParsed1.Merge(jsonParsed2)
	if !jsonParsed1.Exists("outter2", "value2") || !jsonParsed1.Exists("outter", "inner", "value3") {
		t.Errorf("outter.value2 is missing")
	}
}
