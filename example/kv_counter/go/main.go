// KV Counter is a small Tarmac Function that provides a KV store backed incremental counter. This WASM function
// can be called as a HTTP Request, Scheduled Task, or even invoked via a message broker.
package main

import (
	"encoding/base64"
	"fmt"
	"github.com/valyala/fastjson"
	wapc "github.com/wapc/wapc-guest-tinygo"
	"strconv"
)

func main() {
	// Tarmac uses waPC to facilitate WASM module execution. Modules must register their custom handlers under the
	// appropriate method as shown below.
	wapc.RegisterFunctions(wapc.Functions{
		// Register a GET request handler
		"http:GET": Count,
		// Register a POST request handler
		"http:POST": IncCount,
		// Register a PUT request handler
		"http:PUT": IncCount,
		// Register a DELETE request handler
		"http:DELETE": NoHandler,

		// Register a handler for scheduled tasks
		"scheduler:RUN": IncCount,
	})
}

// NoHandler is a custom Tarmac Handler function that will return a tarmac.ServerResponse JSON that denies
// the client request.
func NoHandler(payload []byte) ([]byte, error) {
	return []byte(`{"status":{"code":503,"status":"Not Implemented"}}`), nil
}

// Count is a custom Tarmac Handler functiont hat will receive a tarmac.ServerRequest JSON payload and
// will return a tarmac.ServerResponse JSON with a counter value. No value found is always 0.
func Count(payload []byte) ([]byte, error) {
	// Fetch current value from Database
	b, err := wapc.HostCall("tarmac", "kvstore", "get", []byte(`{"key":"kv_counter_example"}`))
	if err != nil {
		return []byte(fmt.Sprintf(`{"payload":"%s","status":{"code":200,"status":"Success"}}`, base64.StdEncoding.EncodeToString([]byte("0")))), nil
	}
	j, err := fastjson.ParseBytes(b)
	if err != nil {
		return []byte(fmt.Sprintf(`{"status":{"code":500,"status":"Failed to call parse json - %s"}}`, err)), nil
	}

	// Check if value is missing and return 0 if empty
	if j.GetInt("status", "code") != 200 {
		return []byte(fmt.Sprintf(`{"payload":"%s","status":{"code":200,"status":"Success"}}`, base64.StdEncoding.EncodeToString([]byte("0")))), nil
	}

	// Return KV Stored data
	return []byte(fmt.Sprintf(`{"payload":"%s","status":{"code":200,"status":"Success"}}`, j.GetStringBytes("data"))), nil
}

// IncCount is a custom Tarmac Handler function that will fetch a counter from the datastore, increase it by one
// and then write that counter to the datastore
func IncCount(payload []byte) ([]byte, error) {
	i := 0

	// Fetch current value from Database
	b, err := wapc.HostCall("tarmac", "kvstore", "get", []byte(`{"key":"kv_counter_example"}`))
	if err == nil {
		j, err := fastjson.ParseBytes(b)
		if err != nil {
			return []byte(fmt.Sprintf(`{"status":{"code":500,"status":"Failed to call parse json - %s"}}`, err)), nil
		}

		// Check if value is missing and return 0 if empty
		if j.GetInt("status", "code") == 200 {
			s, err := base64.StdEncoding.DecodeString(string(j.GetStringBytes("data")))
			if err == nil {
				n, err := strconv.Atoi(fmt.Sprintf("%s", s))
				if err == nil {
					i = n
				}
			}
		}
	}

	// Increment Counter
	i += 1
	s := strconv.Itoa(i)

	// Store new Counter value
	_, err = wapc.HostCall("tarmac", "kvstore", "set", []byte(fmt.Sprintf(`{"key":"kv_counter_example","data":"%s"}`, base64.StdEncoding.EncodeToString([]byte(s)))))
	if err != nil {
		return []byte(fmt.Sprintf(`{"status":{"code":500,"status":"Failed to call host callback - %s"}}`, err)), nil
	}

	// Return Counter value to user
	return []byte(fmt.Sprintf(`{"payload":"%s","status":{"code":200,"status":"Success"}}`, base64.StdEncoding.EncodeToString([]byte(s)))), nil
}
