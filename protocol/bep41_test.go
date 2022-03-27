package protocol

import (
	"bytes"
	"math"
	"net/url"
	"testing"
)

func TestExtractExtensionData(t *testing.T) {
	var tests = []struct {
		input    []byte
		expected []byte
	}{
		{
			input:    []byte{byte(BEP41OptionTypeEndOfOptions)},
			expected: nil,
		},
		{
			input:    []byte{byte(BEP41OptionTypeNOP)},
			expected: nil,
		},
		{
			input:    []byte{byte(BEP41OptionTypeURLData), 4, '/', 'a', 'b', 'c'},
			expected: []byte("/abc"),
		},
		{
			input:    []byte{byte(BEP41OptionTypeURLData), 0},
			expected: nil,
		},
		{
			input:    []byte{byte(BEP41OptionTypeURLData), math.MaxUint8},
			expected: nil,
		},
	}

	for _, test := range tests {
		reader := bytes.NewReader(test.input)
		actual, err := ExtractExtensionData(reader)
		if err != nil {
			t.Fatalf("ExtractExtensionData(%v) returned error: %v", test.input, err)
		}

		if !bytes.Equal(actual, test.expected) {
			t.Fatalf("ExtractExtensionData(%v) returned %v, expected %v", test.input, actual, test.expected)
		}
	}
}

func TestConvertUrlDataToUrl(t *testing.T) {
	var tests = []struct {
		input    []byte
		expected *url.URL
	}{
		{
			input:    []byte{},
			expected: &url.URL{},
		},
		{
			input:    []byte{'s'},
			expected: &url.URL{},
		},
		{
			input:    []byte{'/', 'a', 'b', 'c'},
			expected: &url.URL{Path: "/abc"},
		},
	}

	for _, test := range tests {
		actual, _ := ConvertUrlDataToUrl(test.input)
		if actual.String() != test.expected.String() {
			t.Fatalf("ConvertUrlDataToUrl(%v) returned %v, expected %v", test.input, actual, test.expected)
		}
	}
}
