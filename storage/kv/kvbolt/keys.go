package kvbolt

import (
	"bytes"
	"context"
	"sync"

	bolt "go.etcd.io/bbolt"
)

// Keys returns all keys in the bolt bucket.
// The keys channel should be closed if cancel was provided and closed.
func (b *KVBolt) Keys(ctx context.Context, cancel <-chan struct{}) <-chan string {
	r := make(chan string)
	b.execTxn(false, true, func(wg *sync.WaitGroup, bucket *bolt.Bucket) error {
		wg.Add(1)
		go func() {
			defer close(r)
			defer wg.Done()
			c := bucket.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				select {
				case <-cancel:
					return
				case r <- string(k):
				}
			}
		}()
		return nil
	})
	return r
}

// Keys returns all keys starting with prefix in the bolt bucket.
// The keys channel should be closed if cancel was provided and closed.
func (b *KVBolt) KeysPrefix(ctx context.Context, prefix string, cancel <-chan struct{}) <-chan string {
	r := make(chan string)
	b.execTxn(false, true, func(wg *sync.WaitGroup, bucket *bolt.Bucket) error {
		wg.Add(1)
		go func() {
			defer close(r)
			defer wg.Done()
			c := bucket.Cursor()
			for k, _ := c.Seek([]byte(prefix)); k != nil && bytes.HasPrefix(k, []byte(prefix)); k, _ = c.Next() {
				select {
				case <-cancel:
					return
				case r <- string(k):
				}
			}
		}()
		return nil
	})
	return r
}
