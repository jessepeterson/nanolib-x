package kvbolt

import (
	"context"
	"sync"

	"github.com/micromdm/nanolib/storage/kv"
	bolt "go.etcd.io/bbolt"
)

// Get retrieves the value at key from the bolt bucket.
// Rolled-back transactions may return nil.
func (b *KVBolt) Get(_ context.Context, key string) ([]byte, error) {
	var val []byte
	err := b.execTxn(false, false, func(_ *sync.WaitGroup, bucket *bolt.Bucket) error {
		txValue := bucket.Get([]byte(key))
		if txValue == nil {
			return kv.ErrKeyNotFound
		}
		// this is a potentially wasteful copy() depending on if
		// a value is used purely within a transaction (per bolt docs).
		// however it is safer to just copy the bytes.
		val = make([]byte, len(txValue))
		copy(val, txValue)
		return nil
	})
	return val, err
}

// Set sets key to value in the bolt bucket.
func (b *KVBolt) Set(_ context.Context, key string, value []byte) error {
	return b.execTxn(true, false, func(_ *sync.WaitGroup, bucket *bolt.Bucket) error {
		return bucket.Put([]byte(key), value)
	})
}

// Has checks that key is found in the bolt bucket.
func (b *KVBolt) Has(ctx context.Context, key string) (bool, error) {
	var found bool
	err := b.execTxn(false, false, func(_ *sync.WaitGroup, bucket *bolt.Bucket) error {
		if bucket.Get([]byte(key)) != nil {
			found = true
		}
		return nil
	})
	return found, err
}

// Delete deletes key in the bolt bucket.
func (b *KVBolt) Delete(ctx context.Context, key string) error {
	return b.execTxn(true, false, func(_ *sync.WaitGroup, bucket *bolt.Bucket) error {
		return bucket.Delete([]byte(key))
	})
}
