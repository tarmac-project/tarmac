package sanitize

import (
	"testing"
)

type TestCase struct {
	input    string
	expected string
}

func TestSanitize(t *testing.T) {
	tt := []TestCase{
		{"hello\nworld", "helloworld"},
		{"hello\rworld", "helloworld"},
		{"hello\r\nworld", "helloworld"},
		{"hello world", "hello world"},
		{`{ "hello": "world" }`, `{ "hello": "world" }`},
	}

	for _, tc := range tt {
		if got := String(tc.input); got != tc.expected {
			t.Errorf("Sanitize(%s) = %s; want %s", tc.input, got, tc.expected)
		}
	}
}
