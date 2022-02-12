package callbacks

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type Counter struct {
	sync.RWMutex
	count int
}

func (c *Counter) Increment() {
	c.Lock()
	defer c.Unlock()
	c.count++
}

func (c *Counter) Value() int {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func TestCallbacksDefaults(t *testing.T) {
	router := New(Config{})
	counter := &Counter{}

	t.Run("Add New Callback", func(t *testing.T) {
		router.RegisterCallback("counter", "++", func([]byte) ([]byte, error) {
			counter.Increment()
			return []byte(""), nil
		})
	})

	t.Run("Call Callback", func(t *testing.T) {
		_, err := router.Callback(context.Background(), "default", "counter", "++", []byte(""))
		if err != nil {
			t.Errorf("Unexpected error when calling Callback function for registered callback - %s", err)
		}

		if counter.Value() != 1 {
			t.Errorf("Counter was not called")
		}
	})
}

func TestCallbacks(t *testing.T) {
	postCount := &Counter{}
	router := New(Config{
		PreFunc: func(namespace, op string, b []byte) ([]byte, error) {
			if namespace == "badfunc" {
				return []byte(""), fmt.Errorf("Forced Error")
			}
			return []byte(""), nil
		},
		PostFunc: func(CallbackResult) {
			postCount.Increment()
		},
	})
	counter := &Counter{}
	ctx, cancel := context.WithCancel(context.Background())

	t.Run("Add New Callback", func(t *testing.T) {
		router.RegisterCallback("counter", "++", func([]byte) ([]byte, error) {
			counter.Increment()
			return []byte(""), nil
		})
	})

	t.Run("Call Callback", func(t *testing.T) {
		_, err := router.Callback(ctx, "default", "counter", "++", []byte(""))
		if err != nil {
			t.Errorf("Unexpected error when calling Callback function for registered callback - %s", err)
		}

		if counter.Value() != 1 {
			t.Errorf("Counter was not called")
		}
	})

	t.Run("Call Callback with expired context", func(t *testing.T) {
		cancel()
		_, err := router.Callback(ctx, "default", "counter", "++", []byte(""))
		if err == nil {
			t.Errorf("Expected error when calling Callback function with expired context")
		}

		if counter.Value() == 2 {
			t.Errorf("Counter was unexpectedly called")
		}
	})

	t.Run("Delete Callback", func(t *testing.T) {
		router.DelCallback("counter", "++")

		_, err := router.Callback(context.Background(), "default", "counter", "++", []byte(""))
		if err == nil {
			t.Errorf("Expected error when calling Callback function after deletion")
		}
	})

	t.Run("Empty namespace", func(t *testing.T) {
		_, err := router.Callback(context.Background(), "default", "", "++", []byte(""))
		if err == nil {
			t.Errorf("Expected error when calling Callback function, got nil")
		}
	})

	t.Run("Empty op", func(t *testing.T) {
		_, err := router.Callback(context.Background(), "default", "counter", "", []byte(""))
		if err == nil {
			t.Errorf("Expected error when calling Callback function, got nil")
		}
	})

	t.Run("Bad PreFunc Callback", func(t *testing.T) {
		// Add a Nil Function
		router.RegisterCallback("badfunc", "nil", nil)
		_, err := router.Callback(ctx, "default", "badfunc", "nil", []byte(""))
		if err == nil {
			t.Errorf("Expected error when calling Callback function with nil func")
		}
	})

	t.Run("Nil Func Callback", func(t *testing.T) {
		// Add a Nil Function
		router.RegisterCallback("badfunc2", "nil", nil)
		_, err := router.Callback(ctx, "default2", "badfunc", "nil", []byte(""))
		if err == nil {
			t.Errorf("Expected error when calling Callback function with nil func")
		}
	})

	t.Run("Verify PostFunc was called", func(t *testing.T) {
		router.RegisterCallback("goodfunc1", "post", func([]byte) ([]byte, error) { return []byte(""), nil })
		_, err := router.Callback(context.Background(), "default", "goodfunc1", "post", []byte(""))
		if err != nil {
			t.Errorf("Unexpected error calling good callback - %s", err)
		}

		<-time.After(time.Duration(10) * time.Second)

		if postCount.Value() < 1 {
			t.Errorf("Expected post func to be called, but it was not - %d", postCount.Value())
		}
	})
}
