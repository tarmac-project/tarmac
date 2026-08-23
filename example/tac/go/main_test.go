package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/tarmac-project/sdk"
	"github.com/tarmac-project/sdk/kv"
)

type fakeKVClient struct {
	getData  []byte
	getErr   error
	setErr   error
	getCalls int
	setCalls int
	setKey   string
	setValue []byte
}

func (f *fakeKVClient) Config() sdk.RuntimeConfig { return sdk.RuntimeConfig{} }

func (f *fakeKVClient) Get(string) ([]byte, error) {
	f.getCalls++
	return f.getData, f.getErr
}

func (f *fakeKVClient) Set(key string, value []byte) error {
	f.setCalls++
	f.setKey = key
	f.setValue = append([]byte(nil), value...)
	return f.setErr
}

func (f *fakeKVClient) Delete(string) error { return nil }

func (f *fakeKVClient) Keys() ([]string, error) { return nil, nil }

func (f *fakeKVClient) Close() error { return nil }

type fakeLogger struct {
	errorCalls int
}

func (*fakeLogger) Info(string)    {}
func (*fakeLogger) Warn(string)    {}
func (f *fakeLogger) Error(string) { f.errorCalls++ }
func (*fakeLogger) Debug(string)   {}
func (*fakeLogger) Trace(string)   {}

func TestHandler(t *testing.T) {
	readErr := errors.New("cache unavailable")
	writeErr := errors.New("cache write failed")

	tests := []struct {
		name           string
		payload        []byte
		client         *fakeKVClient
		want           []byte
		wantErr        string
		wantWrappedErr error
		wantGetCalls   int
		wantSetCalls   int
		wantErrorCalls int
	}{
		{
			name:    "rejects empty payload before cache access",
			payload: nil,
			client:  &fakeKVClient{},
			wantErr: "payload cannot be empty",
		},
		{
			name:         "returns cached payload",
			payload:      []byte("abc"),
			client:       &fakeKVClient{getData: []byte("cached")},
			want:         []byte("cached"),
			wantGetCalls: 1,
		},
		{
			name:         "reverses and caches on miss",
			payload:      []byte("abc"),
			client:       &fakeKVClient{getErr: kv.ErrKeyNotFound},
			want:         []byte("cba"),
			wantGetCalls: 1,
			wantSetCalls: 1,
		},
		{
			name:           "wraps cache read errors",
			payload:        []byte("abc"),
			client:         &fakeKVClient{getErr: readErr},
			wantErr:        "unable to read cached payload: cache unavailable",
			wantWrappedErr: readErr,
			wantGetCalls:   1,
		},
		{
			name:           "keeps cache writes best effort",
			payload:        []byte("abc"),
			client:         &fakeKVClient{getErr: kv.ErrKeyNotFound, setErr: writeErr},
			want:           []byte("cba"),
			wantGetCalls:   1,
			wantSetCalls:   1,
			wantErrorCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &fakeLogger{}
			kvStore = tt.client
			logger = log

			got, err := Handler(append([]byte(nil), tt.payload...))
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("Handler() error = nil")
				}
				if tt.wantWrappedErr != nil && !errors.Is(err, tt.wantWrappedErr) {
					t.Fatalf("Handler() error = %v, want wrapped %v", err, tt.wantWrappedErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("Handler() error = %q, want %q", err, tt.wantErr)
				}
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Handler() = %q, want %q", got, tt.want)
			}
			if tt.client.getCalls != tt.wantGetCalls {
				t.Errorf("Get() calls = %d, want %d", tt.client.getCalls, tt.wantGetCalls)
			}
			if tt.client.setCalls != tt.wantSetCalls {
				t.Errorf("Set() calls = %d, want %d", tt.client.setCalls, tt.wantSetCalls)
			}
			if log.errorCalls != tt.wantErrorCalls {
				t.Errorf("Error() calls = %d, want %d", log.errorCalls, tt.wantErrorCalls)
			}
			if tt.wantSetCalls > 0 {
				if tt.client.setKey != string(tt.payload) {
					t.Errorf("Set() key = %q, want %q", tt.client.setKey, tt.payload)
				}
				if !bytes.Equal(tt.client.setValue, tt.want) {
					t.Errorf("Set() value = %q, want %q", tt.client.setValue, tt.want)
				}
			}
		})
	}
}
