package callbacks

import (
	"context"
	"sync"
	"testing"
)

type Counter struct {
	sync.RWMutex
	count int
}

func (c *Counter) Increment() {
	c.Lock()
	defer c.Unlock()
	c.count += 1
}

func (c *Counter) Value() int {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func TestCallbacks(t *testing.T) {
	router := New(Config{})
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
}
