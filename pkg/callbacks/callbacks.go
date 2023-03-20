/*
Package callbacks provides a Host Callback router for wapc-go servers.

With this package, users can register Callback functions with the specified namespace and function name. Users can then set this router's Callback method in the waPC server.

When the waPC server calls the callback function, the provided namespace and function name will be used as a lookup to execute the pre-defined host callback.

	router := callbacks.New()
	router.RegisterCallback("database:kv", "Get", myFunc)
	router.RegisterCallback("database:kv", "Set", myFunc2)
	router.RegisterCallback("database:kv", "Delete", myFunc3)

	module, err := wapc.New(someCode, router.Callback)
	if err != nil {
	  // do stuff
	}
*/
package callbacks

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNotFound = fmt.Errorf("callback not found")
	ErrCanceled = fmt.Errorf("callback context has expired")
)

// Router is a callback router, which provides a Callback function that users can register with wapc-go. This callback function will
// determine which registered callback to execute based on the namespace and function name when invoked.
type Router struct {
	sync.RWMutex

	// callbacks stores registered callbacks for lookups when the router is called.
	callbacks map[string]*Callback

	// preFunc holds a user-defined function called as part of the primary Callback method but before the user-specified callback.
	// This preFunc can act as middleware enabling users to restrict access to specific callbacks or perform standard functions with all callbacks.
	preFunc func(string, string, []byte) ([]byte, error)

	// postFunc is a user-defined function called after the execution of the registered callback function. This postFunc is
	// used to enable tracking of callback execution as well as the results of each execution.
	postFunc func(CallbackResult)
}

// Config will configure the Callbacks router and allows users to specify items such as the PreFunc or any router level configurations.
type Config struct {
	// PreFunc holds a user-defined function called as part of the primary Callback method but before the user-specified callback.
	// This PreFunc can act as middleware enabling users to restrict access to specific callbacks or perform standard functions with all callbacks.
	PreFunc func(string, string, []byte) ([]byte, error)

	// PostFunc is a user-defined function called after the execution of the registered callback function. This PostFunc is
	// used to enable tracking of callback execution as well as the results of each execution.
	PostFunc func(CallbackResult)
}

// Callback is a type that holds the details and function used for callback execution. This type is primarily used internally but
// also returned from some public methods.
type Callback struct {
	// Namespace is the common namespace for this callback, and an example could be "database" or for a specific type of
	// database, "database:kv".
	Namespace string

	// Operation is the function within the namespace to call. For example, a "database:kv" namespace may have a "Get" function.
	Operation string

	// Func is the user-provided callback function. This function will execute when the router receives a call with the
	// specified Namespace and Operation key.
	Func func([]byte) ([]byte, error)
}

// CallbackResult provides detailed information regarding Callback execution. The callback result is the input to the
// user-defined PostFunc added in the callback configuration.
type CallbackResult struct {
	// Namespace is the common namespace for this callback, and an example could be "database" or for a specific type of
	// database, "database:kv".
	Namespace string

	// Operation is the function within the namespace to call. For example, a "database:kv" namespace may have a "Get" function.
	Operation string

	// Input is the WASM function supplied input data.
	Input []byte

	// Output is the Callback function returned output data.
	Output []byte

	// Err is the Callback function returned error.
	Err error

	// StartTime provides the time captured at the begging of Callback execution.
	StartTime time.Time

	// EndTime provides the time captured at the end of Callback execution.
	EndTime time.Time
}

// New will return a Router instance that users can use to register callbacks and provide a generic host callback function.
func New(cfg Config) *Router {
	r := &Router{
		preFunc:  cfg.PreFunc,
		postFunc: cfg.PostFunc,
	}
	r.callbacks = make(map[string]*Callback)
	return r
}

// Lookup will fetch the Callback requested from the internal map storage.
func (r *Router) Lookup(key string) (*Callback, error) {
	r.RLock()
	defer r.RUnlock()
	if c, ok := r.callbacks[key]; ok {
		return c, nil
	}
	return &Callback{}, ErrNotFound
}

// Callback is the public callback method, users can register this method as part of a waPC-go module, and when called,
// it will determine if the received host call has a registered callback or not. If yes, this method will execute the
// registered Callback; if not, it will return an error indicating the callback method not found.
func (r *Router) Callback(ctx context.Context, _, namespace, op string, data []byte) ([]byte, error) {
	if namespace == "" || op == "" {
		return []byte(""), fmt.Errorf("namespace and op cannot be nil")
	}

	// Lookup registered Callback
	key := fmt.Sprintf("%s:%s", namespace, op)
	c, err := r.Lookup(key)
	if err != nil {
		return []byte(""), err
	}

	// If callback context is canceled return error
	if ctx.Err() != nil {
		return []byte(""), ErrCanceled
	}

	// Call the user-defined PreFunc returning any errors to the caller.
	if r.preFunc != nil {
		b, err := r.preFunc(namespace, op, data)
		if err != nil {
			return b, err
		}
	}

	// Run the callback returning to user
	if c.Func != nil {
		result := CallbackResult{
			Namespace: namespace,
			Operation: op,
			Input:     data,
			StartTime: time.Now(),
		}

		// Call user-function and capture results
		result.Output, result.Err = c.Func(result.Input)
		result.EndTime = time.Now()

		// Call user-defined PostFunc
		if r.postFunc != nil {
			go r.postFunc(result)
		}
		return result.Output, result.Err
	}
	return []byte(""), fmt.Errorf("unable to execute callback function, function is nil")
}

// RegisterCallback will add the provided function into the internal map store of Callbacks, which will
// enable the Callback method to find and execute the supplied function when appropriate.
func (r *Router) RegisterCallback(namespace, op string, f func([]byte) ([]byte, error)) {
	r.Lock()
	defer r.Unlock()
	r.callbacks[fmt.Sprintf("%s:%s", namespace, op)] = &Callback{
		Namespace: namespace,
		Operation: op,
		Func:      f,
	}
}

// DelCallback will remove any Callback functions saved for the specified namespace and operation.
func (r *Router) DelCallback(namespace, op string) {
	r.Lock()
	defer r.Unlock()
	delete(r.callbacks, fmt.Sprintf("%s:%s", namespace, op))
}
