package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

//export HTTPHandler
func HTTPHandler() int {
	d, err := base64.StdEncoding.DecodeString(os.Getenv("HTTP_PAYLOAD"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid Payload")
		return 400
	}
	if os.Getenv("HTTP_METHOD") == "POST" || os.Getenv("HTTP_METHOD") == "PUT" {
		fmt.Fprintf(os.Stdout, "%s", d)
	}
	return 200
}

func main() {}
