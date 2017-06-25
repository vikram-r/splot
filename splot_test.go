package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type TestLoadDataPair struct {
	testData string
	expected *DataSet
}

type TestLoadInvalidDataPair struct {
	testData string
	expected error
}

func TestLoadValidData(t *testing.T) {
	testCases := []TestLoadDataPair{
		{"x,y", &DataSet{[]Point{}, "x", "y"}},
		{"x,y\n", &DataSet{[]Point{}, "x", "y"}},
		{"x,y,ignored", &DataSet{[]Point{}, "x", "y"}},
		{"x,y\n1,2", &DataSet{[]Point{{1, 2}}, "x", "y"}},
		{"x,y\n1,2\n3,4", &DataSet{[]Point{{1, 2}, {3, 4}}, "x", "y"}},
		{"x,y\n3,4\n1,2", &DataSet{[]Point{{1, 2}, {3, 4}}, "x", "y"}},
		{"x,y\n3,4\n1,2\n2,8", &DataSet{[]Point{{1, 2}, {2, 8}, {3, 4}}, "x", "y"}},
	}

	for _, test := range testCases {
		if r, err := loadData(strings.NewReader(test.testData)); err != nil {
			t.Errorf("Expected: %v\n Actual Error: %v", test.expected, err)
		} else {
			if !reflect.DeepEqual(test.expected, r) {
				t.Errorf("DataSet for testData: %s was incorrect!\nExpected: %v\nActual: %v", test.testData, test.expected, r)
			}
		}
	}
}

func TestLoadInvalidData(t *testing.T) {
	testCases := []TestLoadInvalidDataPair{
		{"", errors.New("No data found")},
		{"x", errors.New("Header with 2 elements required")},
		{"x,", errors.New("Header with 2 elements required")},
		{"x,y\n1", errors.New("Could not parse row 1: \"1\", reason: \"coordinates require 2 values\"")},
		{"x,y\n1,2\n3", errors.New("Could not parse row 2: \"3\", reason: \"coordinates require 2 values\"")},
		{"x,y\ngarbagex,2", errors.New("Could not parse row 1: \"garbagex,2\", reason: \"garbagex is not a number\"")},
		{"x,y\n2,garbagey", errors.New("Could not parse row 1: \"2,garbagey\", reason: \"garbagey is not a number\"")},
	}

	for _, test := range testCases {
		if _, err := loadData(strings.NewReader(test.testData)); err != nil {
			if test.expected.Error() != err.Error() {
				t.Errorf("Expected: %v\n Actual: %v", test.expected, err)
			}
		} else {
			t.Errorf("Expected error, but none found!")
		}
	}
}
