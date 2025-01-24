package nats

import (
	"errors"

	"github.com/Mattilsynet/map-managed-environment/gen/mattilsynet/map-kv/key-value"
	"github.com/bytecodealliance/wasm-tools-go/cm"
)

type (
	KeyValue      struct{}
	KeyValueEntry struct {
		key   string
		value []byte
	}
)

func (e *KeyValueEntry) Key() string   { return e.key }
func (e *KeyValueEntry) Value() []byte { return e.value }

func (js *JetStreamContext) KeyValue() (*KeyValue, error) {
	js.bucket = KeyValue{}
	return &js.bucket, nil
}

func (js *KeyValue) Get(key string) (*KeyValueEntry, error) {
	result := keyvalue.Get(key)
	if result.IsOK() {
		resVal := result.OK().Value.Slice()
		resKey := result.OK().Key
		return &KeyValueEntry{resKey, resVal}, nil
	}
	if result.IsErr() {
		return nil, errors.New(*result.Err())
	}
	return nil, errors.New("unknown error when getting keyvalue from map-kv with key: " + key)
}

func (js *KeyValue) GetAll() ([]*KeyValueEntry, error) {
	listKeys := keyvalue.ListKeys()
	if listKeys.IsOK() {
		keys := listKeys.OK().Slice()
		var entries []*KeyValueEntry
		for _, key := range keys {
			result := keyvalue.Get(key)
			if result.IsOK() {
				resVal := result.OK().Value.Slice()
				resKey := result.OK().Key
				entries = append(entries, &KeyValueEntry{resKey, resVal})
			}
			if result.IsErr() {
				return nil, errors.New(*result.Err())
			}
		}
		return entries, nil
	}
	if listKeys.IsErr() {
		return nil, errors.New(*listKeys.Err())
	}
	return nil, errors.New("unknown error when getting all keyvalues from map-kv")
}

func (js *KeyValue) Put(key string, value []byte) error {
	result := keyvalue.Put(key, cm.ToList(value))
	if result.IsOK() {
		return nil
	}
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return errors.New("unknown error when putting keyvalue in map-kv with key: " + key)
}

func (js *KeyValue) Create(key string, value []byte) error {
	result := keyvalue.Create(key, cm.ToList(value))
	if result.IsOK() {
		return nil
	}
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return errors.New("unknown error when creating keyvalue in map-kv with key: " + key)
}

func (js *KeyValue) Delete(key string) error {
	result := keyvalue.Delete(key)
	if result.IsOK() {
		return nil
	}
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return errors.New("unknown error when deleting keyvalue in map-kv with key: " + key)
}

func (js *KeyValue) ListKeys() ([]string, error) {
	result := keyvalue.ListKeys()
	if result.IsOK() {
		return result.OK().Slice(), nil
	}
	if result.IsErr() {
		return nil, errors.New(*result.Err())
	}
	return nil, errors.New("unknown error when listing keys in map-kv")
}
