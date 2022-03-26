package protocol

import (
	"bytes"
	"math"
	"testing"
)

func TestExtractExtensionData(t *testing.T) {
	var tests = []struct {
		input    []byte
		expected string
	}{
		{
			input:    []byte{byte(BEP41OptionTypeEndOfOptions)},
			expected: "",
		},
		{
			input:    []byte{byte(BEP41OptionTypeNOP)},
			expected: "",
		},
		{
			input:    []byte{byte(BEP41OptionTypeURLData), 4, '/', 'a', 'b', 'c'},
			expected: "/abc",
		},
		{
			input:    []byte{byte(BEP41OptionTypeURLData), 0},
			expected: "",
		},
		{
			input:    []byte{byte(BEP41OptionTypeURLData), math.MaxUint8},
			expected: "",
		},
	}

	for _, test := range tests {
		reader := bytes.NewReader(test.input)
		actual, err := ExtractExtensionData(reader)
		if err != nil {
			t.Fatalf("ExtractExtensionData(%v) returned error: %v", test.input, err)
		}

		if actual != test.expected {
			t.Fatalf("ExtractExtensionData(%v) returned %v, expected %v", test.input, actual, test.expected)
		}
	}
}
