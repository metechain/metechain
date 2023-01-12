package pb

import "github.com/cockroachdb/pebble"

type pbStore struct {
	db *pebble.DB
}

type pbTransaction struct {
	db *pebble.DB
}

type pbIterator struct {
	err error
	itr *pebble.Iterator
}
