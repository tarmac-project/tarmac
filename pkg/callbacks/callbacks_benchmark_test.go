package callbacks

import (
	"context"
	"testing"
)

func BenchmarkCallback(b *testing.B) {
	router := New(Config{})
	counter := &Counter{}

	router.RegisterCallback("counter", "++", func([]byte) ([]byte, error) {
		counter.Increment()
		return []byte(""), nil
	})

	for i := 0; i < b.N; i++ {
		_, err := router.Callback(context.Background(), "default", "counter", "++", []byte(""))
		if err != nil {
			b.Errorf("Unexpected error when calling Callback function for registered callback - %s", err)
		}
	}

	if counter.Value() != b.N {
		b.Errorf("Counter value did not match the number of iterations")
	}
}
