package hashmap

import (
	"crypto/sha256"
	"encoding/gob"
)

func Sum(x interface{}) [32]byte {
	h := sha256.New()
	err := gob.NewEncoder(h).Encode(x)
	if err != nil {
		panic(err)
	}
	return [32]byte(h.Sum([]byte{}))
}

type HashMap[K any, V any] map[[32]byte]struct {
	Key   K
	Value V
}

func (h HashMap[K, V]) Get(k K) (V, bool) {
	x, ok := h[Sum(k)]
	return x.Value, ok
}
func (h HashMap[K, V]) Put(k K, v V) {
	h[Sum(k)] = struct {
		Key   K
		Value V
	}{Key: k, Value: v}
}
