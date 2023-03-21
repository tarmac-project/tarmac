package function

import (
	"fmt"
	"testing"
)

func TestFunction(t *testing.T) {
	// Initialize function with mock hostCall
	function, err := New(Config{Namespace: "default", HostCall: func(namespace string, capability string, function string, input []byte) ([]byte, error) {
		if namespace != "default" || capability != "function" || function != "test-func" {
			t.Errorf("hostcall signature invalid %s, %s, %s", namespace, capability, function)
		}
		if len(input) != len([]byte("Hey hey hey")) {
			t.Errorf("unexpected input vs. payload")
		}
		return []byte("Success"), nil
	}})
	if err != nil {
		t.Errorf("unexpected error initializing function - %s", err)
	}

	// Call function method
	d, err := function.Call("test-func", []byte("Hey hey hey"))
	if err != nil {
		t.Errorf("unexpected error returned from function - %s", err)
	}

	if len(d) != len([]byte("Success")) {
		t.Errorf("unexpected output from function to function call")
	}
}

func TestFunctionFail(t *testing.T) {
	// Initialize function with mock hostCall
	function, err := New(Config{Namespace: "default", HostCall: func(namespace string, capability string, function string, input []byte) ([]byte, error) {
		if namespace != "default" || capability != "function" || function != "test-func" {
			t.Errorf("hostcall signature invalid %s, %s, %s", namespace, capability, function)
		}
		if len(input) != len([]byte("Hey hey hey")) {
			t.Errorf("unexpected input vs. payload")
		}
		return []byte(""), fmt.Errorf("This is an error")
	}})
	if err != nil {
		t.Errorf("unexpected error initializing function - %s", err)
	}

	// Call function method
	_, err = function.Call("test-func", []byte("Hey hey hey"))
	if err == nil {
		t.Errorf("unexpected success returned from function")
	}
}
