package bolt

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ensure a bucket can calculate stats.
func TestBucketStat(t *testing.T) {
	withOpenDB(func(db *DB, path string) {
		db.Do(func(txn *Transaction) error {
			// Add bucket with lots of keys.
			txn.CreateBucket("widgets")
			b := txn.Bucket("widgets")
			for i := 0; i < 100000; i++ {
				b.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
			}

			// Add bucket with fewer keys but one big value.
			txn.CreateBucket("woojits")
			b = txn.Bucket("woojits")
			for i := 0; i < 500; i++ {
				b.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
			}
			b.Put([]byte("really-big-value"), []byte(strings.Repeat("*", 10000)))

			// Add a bucket that fits on a single root leaf.
			txn.CreateBucket("whozawhats")
			b = txn.Bucket("whozawhats")
			b.Put([]byte("foo"), []byte("bar"))

			return nil
		})
		db.With(func(txn *Transaction) error {
			b := txn.Bucket("widgets")
			stat := b.Stat()
			assert.Equal(t, stat.BranchPageCount, 15)
			assert.Equal(t, stat.LeafPageCount, 1281)
			assert.Equal(t, stat.OverflowPageCount, 0)
			assert.Equal(t, stat.KeyCount, 100000)
			assert.Equal(t, stat.MaxDepth, 3)

			b = txn.Bucket("woojits")
			stat = b.Stat()
			assert.Equal(t, stat.BranchPageCount, 1)
			assert.Equal(t, stat.LeafPageCount, 6)
			assert.Equal(t, stat.OverflowPageCount, 2)
			assert.Equal(t, stat.KeyCount, 501)
			assert.Equal(t, stat.MaxDepth, 2)

			b = txn.Bucket("whozawhats")
			stat = b.Stat()
			assert.Equal(t, stat.BranchPageCount, 0)
			assert.Equal(t, stat.LeafPageCount, 1)
			assert.Equal(t, stat.OverflowPageCount, 0)
			assert.Equal(t, stat.KeyCount, 1)
			assert.Equal(t, stat.MaxDepth, 1)

			return nil
		})
	})
}
