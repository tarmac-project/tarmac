package main

import (
	"os"
	"testing"
)

func TestHTTPHandler(t *testing.T) {
	os.Setenv("REMOTE_ADDR", "127.0.0.1:9000")
	defer os.Setenv("REMOTE_ADDR", "")

	t.Run("No Payload", func(t *testing.T) {
		r := HTTPHandler()
		if r != 200 {
			t.Errorf("Unexpected HTTP Return code got %d", r)
		}
	})

	t.Run("Good Payload", func(t *testing.T) {
		os.Setenv("HTTP_METHOD", "POST")
		defer os.Setenv("HTTP_METHOD", "")
		os.Setenv("HTTP_PAYLOAD", "VGhpcyBpcyBzdHVmZgo=")
		defer os.Setenv("HTTP_PAYLOAD", "")
		r := HTTPHandler()
		if r != 200 {
			t.Errorf("Unexpected HTTP Return code got %d", r)
		}
	})

	t.Run("Bad Payload", func(t *testing.T) {
		os.Setenv("HTTP_PAYLOAD", "this@#$@is!@##!@not!@#!@#!@base64")
		defer os.Setenv("HTTP_PAYLOAD", "")
		r := HTTPHandler()
		if r < 400 {
			t.Errorf("Unexpected HTTP Return code got %d", r)
		}
	})
}
