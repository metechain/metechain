package pb

import (
	"metachain/pkg/storage/store"

	"github.com/cockroachdb/pebble"
)

func (db *pbStore) NewIterator(prefix []byte, start []byte) store.Iterator {
	keyUpperBound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}
	prefixIterOptions := func(prefix []byte) *pebble.IterOptions {
		return &pebble.IterOptions{
			LowerBound: prefix,
			UpperBound: keyUpperBound(prefix),
		}
	}
	itr := db.db.NewIter(prefixIterOptions(prefix))
	itr.SeekGE(start)
	return &pbIterator{
		itr: itr,
	}
}

func (itr *pbIterator) Next() bool {
	itr.itr.Next()
	return itr.itr.Valid()
}

func (itr *pbIterator) Error() error {
	return nil
}

func (itr *pbIterator) Key() []byte {
	return append([]byte{}, itr.itr.Key()...)
}

func (itr *pbIterator) Value() []byte {
	return append([]byte{}, itr.itr.Value()...)
}

func (itr *pbIterator) Release() {
	itr.itr.Close()
}
