package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestLoadDataValid(t *testing.T) {
	testCases := []struct {
		testData string
		expected *DataSet
	}{
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

func TestLoadDataInvalid(t *testing.T) {
	testCases := []struct {
		testData string
		expected error
	}{
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

func TestParseRowValid(t *testing.T) {
	testCases := []struct {
		row      string
		expected Point
	}{
		{"1,2", Point{1, 2}},
		{"1123,2098", Point{1123, 2098}},
	}

	for _, test := range testCases {
		if r, err := parseRow(test.row); err != nil {
			t.Errorf("Expected: %v\n Actual Error: %v", test.expected, err)
		} else {
			if !(test.expected.x == r.x) || !(test.expected.y == r.y) {
				t.Errorf("Expected: %v\n Actual: %v", test.expected, r)
			}
		}
	}
}

func TestParseRowInvalid(t *testing.T) {
	testCases := []struct {
		row      string
		expected error
	}{
		{"1", errors.New("coordinates require 2 values")},
		{"foo", errors.New("coordinates require 2 values")},
		{"foo,1", errors.New("foo is not a number")},
		{"1,bar", errors.New("bar is not a number")},
		{"foo,1,2", errors.New("foo is not a number")},
		{"1,bar,2", errors.New("bar is not a number")},
		{"foo,bar", errors.New("foo is not a number")},
		{"1.0,2", errors.New("1.0 is not a number")},
		{"1,2.0", errors.New("2.0 is not a number")},
	}

	for _, test := range testCases {
		if _, err := parseRow(test.row); err != nil {
			if test.expected.Error() != err.Error() {
				t.Errorf("Expected: %v\n Actual: %v", test.expected, err)
			}
		} else {
			t.Errorf("Expected error, but none found!")
		}
	}
}
