// Package kvbolt is a NanoLIB kv store that uses bbolt for the backend storage.
package kvbolt

import (
	"errors"
	"fmt"
	"sync"

	bolt "go.etcd.io/bbolt"
)

var ErrBucketDoesNotExist = errors.New("bucket does not exist")
var ErrNoTxn = errors.New("no txn")

// KVBolt provides a key-value store utilizng bolt.
// When a transaction has not begun it will run individual KV operations
// within individual transactions.
type KVBolt struct {
	db         *bolt.DB
	txn        *bolt.Tx
	txnBucket  *bolt.Bucket
	bucketName []byte
	writable   bool
}

// new is a helper for creating KVBolt stores and transactions.
func new(db *bolt.DB, bucketName []byte, newTxn bool, writable bool) (*KVBolt, error) {
	if db == nil {
		return nil, errors.New("nil db")
	}
	b := &KVBolt{
		bucketName: bucketName,
		writable:   writable,
	}
	if newTxn {
		var err error
		b.txn, err = db.Begin(writable)
		if err != nil {
			return b, err
		}
		b.txnBucket = b.txn.Bucket(b.bucketName)
		if b.txnBucket == nil {
			return b, ErrBucketDoesNotExist
		}
	} else {
		b.db = db
	}
	return b, nil
}

// New creates a new key-value store utilizing bolt.
// It is assumed bucketName already exists in db.
func New(db *bolt.DB, bucketName []byte) (*KVBolt, error) {
	return new(db, bucketName, false, true)
}

// boltTxnFn runs a function utilizing a transaction.
type boltTxnFn func(func(*bolt.Tx) error) error

// boltOpFn runs transaction operations on the store.
type boltOpFn func(*sync.WaitGroup, *bolt.Bucket) error

// execTxn executes operations on the bolt db.
// Either within a larger transaction or creating a new transaction.
func (b *KVBolt) execTxn(writable bool, exitEarly bool, fn boltOpFn) error {
	var runner boltTxnFn = func(txnFn func(*bolt.Tx) error) error {
		return txnFn(b.txn)
	}
	if b.txn == nil {
		// use bolt's View or Update methods to start new transactions
		runner = b.db.View
		if writable && b.writable {
			runner = b.db.Update
		}
	}
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// we run in a goroutine to potentially exit early
		// for i.e. the Keys and KeysPrefix methods.
		defer wg.Done()
		err = runner(func(tx *bolt.Tx) error {
			bucket := b.txnBucket
			if b.txn == nil {
				// if we're in a new/individual transaction then we
				// have to get a reference the bucket.
				bucket = tx.Bucket(b.bucketName)
				if bucket == nil {
					return ErrBucketDoesNotExist
				}
			}
			var wgOp sync.WaitGroup
			err := fn(&wgOp, bucket)
			// most fn's won't add to wgOp so the
			// Wait here will exit immediately
			wgOp.Wait()
			if exitEarly {
				// don't bother using err since an early-exiting
				// won't have access to it anyway.
				return nil
			}
			if err != nil {
				// add some context to the error for debugging
				err = fmt.Errorf("txn runner: %w", err)
			}
			return err
		})
	}()
	if !exitEarly {
		wg.Wait()
		if err != nil {
			// add some context to the error for debugging
			err = fmt.Errorf("txn: %w", err)
		}
		return err
	}
	return nil
}
