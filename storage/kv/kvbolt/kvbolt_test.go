package kvbolt

import (
	"context"
	"path"
	"testing"

	bolt "go.etcd.io/bbolt"

	"github.com/micromdm/nanolib/storage/kv/test"
)

func assureBucket(db *bolt.DB, bucketName []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	return err
}

func newKVBolt(db *bolt.DB, bucketName []byte) (*KVBolt, error) {
	err := assureBucket(db, bucketName)
	if err != nil {
		return nil, err
	}
	return New(db, bucketName)
}

func TestKVBolt(t *testing.T) {
	db, err := bolt.Open(path.Join(t.TempDir(), "test.db"), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	b, err := newKVBolt(db, []byte("testBucket"))
	if err != nil {
		t.Fatal(err)
	}

	b2, err := newKVBolt(db, []byte("testBucket2"))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	test.TestBucketSimple(t, ctx, b)
	test.TestKeysTraversing(t, ctx, b)
	t.Run("TestTxnSimple", func(t *testing.T) { test.TestTxnSimple(t, ctx, b, test.WithNoReadAfterRollback()) })
	t.Run("TestKVTxnKeys", func(t *testing.T) { test.TestKVTxnKeys(t, ctx, b2) })
}
