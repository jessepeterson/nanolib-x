package kvbolt

import (
	"context"

	"github.com/micromdm/nanolib/storage/kv"
)

// BeginCRUDBucketTxn begins a new transaction.
func (b *KVBolt) BeginCRUDBucketTxn(context.Context) (kv.CRUDBucketTxnCompleter, error) {
	return new(b.db, b.bucketName, true, b.writable)
}

// BeginKeysPrefixTraversingBucketTxn begins a new transaction.
func (b *KVBolt) BeginKeysPrefixTraversingBucketTxn(context.Context) (kv.KeysPrefixTraversingBucketTxnCompleter, error) {
	return new(b.db, b.bucketName, true, b.writable)
}

// BeginBucketTxn begins a new transaction.
func (b *KVBolt) BeginBucketTxn(context.Context) (kv.BucketTxnCompleter, error) {
	return new(b.db, b.bucketName, true, b.writable)
}

// Commit commits the transaction to the bolt database.
func (b *KVBolt) Commit(context.Context) error {
	if b.txn == nil {
		return ErrNoTxn
	}
	return b.txn.Commit()
}

// Rollback rolls back the transaction in the bolt database.
func (b *KVBolt) Rollback(context.Context) error {
	if b.txn == nil {
		return ErrNoTxn
	}
	return b.txn.Rollback()
}
