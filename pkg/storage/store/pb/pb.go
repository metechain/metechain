package pb

import (
	"metachain/pkg/storage/miscellaneous"
	"metachain/pkg/storage/store"

	"github.com/cockroachdb/pebble"
)

func New(db *pebble.DB) store.DB {
	return &pbStore{db}
}

func (db *pbStore) Sync() error {
	return nil
}

func (db *pbStore) Close() error {
	return db.db.Close()
}

func (db *pbStore) Del(k []byte) error {
	return db.db.Delete(k, nil)
}

func (db *pbStore) Set(k, v []byte) error {
	return db.db.Set(k, v, nil)
}

func (db *pbStore) Get(k []byte) ([]byte, error) {
	v, c, err := db.db.Get(k)
	if err != nil {
		return nil, err
	}
	val := make([]byte, len(v))
	copy(val, v)
	c.Close()
	return v, nil
}

func (db *pbStore) NewTransaction() store.Transaction {
	return &pbTransaction{db: db.db}
}

func (tx *pbTransaction) Cancel() error {

	return nil
}

func (tx *pbTransaction) Commit() error {
	return nil
}

func (tx *pbTransaction) Del(k []byte) error {
	return tx.db.Delete(k, nil)
}

func (tx *pbTransaction) Set(k, v []byte) error {
	return tx.db.Set(k, v, nil)
}

func (tx *pbTransaction) Get(k []byte) ([]byte, error) {
	v, c, err := tx.db.Get(k)
	if err != nil {
		return nil, err
	}
	val := make([]byte, len(v))
	copy(val, v)
	c.Close()
	return v, nil
}

func del(tx *pebble.DB, k []byte) error {
	return tx.Delete(k, nil)
}

func set(tx *pebble.DB, k, v []byte) error {
	return tx.Set(k, v, nil)
}

func get(tx *pebble.DB, k []byte) ([]byte, error) {
	v, c, err := tx.Get(k)
	if err != nil {
		return nil, err
	}
	val := make([]byte, len(v))
	copy(val, v)
	c.Close()
	return v, nil
}

func (db *pbStore) Mclear(m []byte) error {
	if err := mclear(db.db, m); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Mdel(m, k []byte) error {
	if err := mdel(db.db, m, k); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Mset(m, k, v []byte) error {
	if err := mset(db.db, m, k, v); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Mget(m, k []byte) ([]byte, error) {
	return mget(db.db, m, k)
}

func (db *pbStore) Mkeys(m []byte) ([][]byte, error) {
	return mkeys(db.db, m)
}

func (db *pbStore) Mvals(m []byte) ([][]byte, error) {
	return mvals(db.db, m)
}

func (db *pbStore) Mkvs(m []byte) ([][]byte, [][]byte, error) {
	return mkvs(db.db, m)
}

func (db *pbStore) Llen(k []byte) int64 {
	return llen(db.db, k)
}

func (db *pbStore) Lclear(k []byte) error {
	if err := llclear(db.db, k); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Llpush(k, v []byte) (int64, error) {
	return llpush(db.db, k, v)
}

func (db *pbStore) Llpop(k []byte) ([]byte, error) {
	return llpop(db.db, k)
}

func (db *pbStore) Lrpush(k, v []byte) (int64, error) {
	return lrpush(db.db, k, v)
}

func (db *pbStore) Lrpop(k []byte) ([]byte, error) {
	return lrpop(db.db, k)
}

func (db *pbStore) Lrange(k []byte, start, end int64) ([][]byte, error) {
	return lrange(db.db, k, start, end)
}

func (db *pbStore) Lset(k []byte, idx int64, v []byte) error {
	if err := lset(db.db, k, idx, v); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Lindex(k []byte, idx int64) ([]byte, error) {
	return lindex(db.db, k, idx)
}

func (db *pbStore) Sclear(k []byte) error {
	if err := sclear(db.db, k); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Sdel(k, v []byte) error {
	if err := sdel(db.db, k, v); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Sadd(k, v []byte) error {
	if err := sadd(db.db, k, v); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Selem(k, v []byte) (bool, error) {
	return selem(db.db, k, v)
}

func (db *pbStore) Smembers(k []byte) ([][]byte, error) {
	return smembers(db.db, k)
}

func (db *pbStore) Zclear(k []byte) error {
	if err := zclear(db.db, k); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Zdel(k, v []byte) error {
	if err := zdel(db.db, k, v); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Zadd(k []byte, score int32, v []byte) error {
	if err := zadd(db.db, k, score, v); err != nil {
		return err
	}
	return nil
}

func (db *pbStore) Zscore(k, v []byte) (int32, error) {
	return zscore(db.db, k, v)
}

func (db *pbStore) Zrange(k []byte, start, end int32) ([][]byte, error) {
	return zrange(db.db, k, start, end)
}

func (tx *pbTransaction) Mclear(m []byte) error {
	return mclear(tx.db, m)
}

func (tx *pbTransaction) Mdel(m, k []byte) error {
	return mdel(tx.db, m, k)
}

func (tx *pbTransaction) Mset(m, k, v []byte) error {
	return mset(tx.db, m, k, v)
}

func (tx *pbTransaction) Mget(m, k []byte) ([]byte, error) {
	return mget(tx.db, m, k)
}

func (tx *pbTransaction) Mkeys(m []byte) ([][]byte, error) {
	return mkeys(tx.db, m)
}

func (tx *pbTransaction) Mvals(m []byte) ([][]byte, error) {
	return mvals(tx.db, m)
}

func (tx *pbTransaction) Mkvs(m []byte) ([][]byte, [][]byte, error) {
	return mkvs(tx.db, m)
}

func (tx *pbTransaction) Llen(k []byte) int64 {
	return llen(tx.db, k)
}

func (tx *pbTransaction) Lclear(k []byte) error {
	return llclear(tx.db, k)
}

func (tx *pbTransaction) Llpush(k, v []byte) (int64, error) {
	return llpush(tx.db, k, v)
}

func (tx *pbTransaction) Llpop(k []byte) ([]byte, error) {
	return llpop(tx.db, k)
}

func (tx *pbTransaction) Lrpush(k, v []byte) (int64, error) {
	return lrpush(tx.db, k, v)
}

func (tx *pbTransaction) Lrpop(k []byte) ([]byte, error) {
	return lrpop(tx.db, k)
}

func (tx *pbTransaction) Lrange(k []byte, start, end int64) ([][]byte, error) {
	return lrange(tx.db, k, start, end)
}

func (tx *pbTransaction) Lset(k []byte, idx int64, v []byte) error {
	return lset(tx.db, k, idx, v)
}

func (tx *pbTransaction) Lindex(k []byte, idx int64) ([]byte, error) {
	return lindex(tx.db, k, idx)
}

func (tx *pbTransaction) Sclear(k []byte) error {
	return sclear(tx.db, k)
}

func (tx *pbTransaction) Sdel(k, v []byte) error {
	return sdel(tx.db, k, v)
}

func (tx *pbTransaction) Sadd(k, v []byte) error {
	return sadd(tx.db, k, v)
}

func (tx *pbTransaction) Selem(k, v []byte) (bool, error) {
	return selem(tx.db, k, v)
}

func (tx *pbTransaction) Smembers(k []byte) ([][]byte, error) {
	return smembers(tx.db, k)
}

func (tx *pbTransaction) Zclear(k []byte) error {
	return zclear(tx.db, k)
}

func (tx *pbTransaction) Zdel(k, v []byte) error {
	return zdel(tx.db, k, v)
}

func (tx *pbTransaction) Zadd(k []byte, score int32, v []byte) error {
	return zadd(tx.db, k, score, v)
}

func (tx *pbTransaction) Zscore(k, v []byte) (int32, error) {
	return zscore(tx.db, k, v)
}

func (tx *pbTransaction) Zrange(k []byte, start, end int32) ([][]byte, error) {
	return zrange(tx.db, k, start, end)
}

func mclear(tx *pebble.DB, m []byte) error {
	k := eMapKey(m, []byte{})
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		if err := tx.Delete(itr.Key(), nil); err != nil {
			return err
		}
	}
	return nil
}

func mdel(tx *pebble.DB, m, k []byte) error {
	return del(tx, eMapKey(m, k))
}

func mset(tx *pebble.DB, m, k, v []byte) error {
	return set(tx, eMapKey(m, k), v)
}

func mget(tx *pebble.DB, m, k []byte) ([]byte, error) {
	return get(tx, eMapKey(m, k))
}

func mkeys(tx *pebble.DB, m []byte) ([][]byte, error) {
	var ks [][]byte

	k := eMapKey(m, []byte{})
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		ks = append(ks, append([]byte{}, itr.Key()...))
	}
	return ks, nil
}

func mvals(tx *pebble.DB, m []byte) ([][]byte, error) {
	var vs [][]byte

	k := eMapKey(m, []byte{})
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		vs = append(vs, append([]byte{}, itr.Value()...))
	}
	return vs, nil
}

func mkvs(tx *pebble.DB, m []byte) ([][]byte, [][]byte, error) {
	var ks, vs [][]byte

	k := eMapKey(m, []byte{})
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		ks = append(ks, append([]byte{}, itr.Key()...))
		vs = append(vs, append([]byte{}, itr.Value()...))
	}
	return ks, vs, nil
}

func lnew(tx *pebble.DB, k []byte) error {
	return set(tx, eListMetaKey(k), eListMetaValue(0, 0))
}

func llen(tx *pebble.DB, k []byte) int64 {
	if start, end, err := listStartEnd(tx, k); err != nil {
		return 0
	} else {
		return end - start
	}
}

func llclear(tx *pebble.DB, k []byte) error {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return err
	}
	for ; start < end; start++ {
		if err := del(tx, eListKey(k, start)); err != nil {
			return err
		}
	}
	return del(tx, eListMetaKey(k))
}

func llpush(tx *pebble.DB, k, v []byte) (int64, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		if err = lnew(tx, k); err != nil {
			return -1, err
		}
	}
	if start-1 == end {
		return -1, store.OutOfSize
	}
	if err := set(tx, eListKey(k, start-1), v); err != nil {
		return -1, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start-1, end)); err != nil {
		return -1, err
	}
	return end - start + 1, nil
}

func llpop(tx *pebble.DB, k []byte) ([]byte, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	if start == end {
		return nil, store.OutOfSize
	}
	v, err := get(tx, eListKey(k, start))
	if err != nil {
		return nil, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start+1, end)); err != nil {
		return nil, err
	}
	return v, nil
}

func lrpush(tx *pebble.DB, k, v []byte) (int64, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		if err = lnew(tx, k); err != nil {
			return -1, err
		}
	}
	if start == end+1 {
		return -1, store.OutOfSize
	}
	if err := set(tx, eListKey(k, end), v); err != nil {
		return -1, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start, end+1)); err != nil {
		return -1, err
	}
	return end - start + 1, nil
}

func lrpop(tx *pebble.DB, k []byte) ([]byte, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	if start == end {
		return nil, store.OutOfSize
	}
	v, err := get(tx, eListKey(k, end-1))
	if err != nil {
		return nil, err
	}
	if err := set(tx, eListMetaKey(k), eListMetaValue(start, end-1)); err != nil {
		return nil, err
	}
	return v, nil
}

func lset(tx *pebble.DB, k []byte, idx int64, v []byte) error {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return err
	}
	switch {
	case idx >= 0:
		idx += start
	default:
		idx += end
	}
	if idx < start || idx >= end {
		return store.OutOfSize
	}
	return set(tx, eListKey(k, idx+start), v)
}

func lindex(tx *pebble.DB, k []byte, idx int64) ([]byte, error) {
	start, end, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	switch {
	case idx >= 0:
		idx += start
	default:
		idx += end
	}
	if idx < start || idx >= end {
		return []byte{}, nil
	}
	return get(tx, eListKey(k, idx))
}

func lrange(tx *pebble.DB, k []byte, start, end int64) ([][]byte, error) {
	var vs [][]byte

	x, y, err := listStartEnd(tx, k)
	if err != nil {
		return nil, err
	}
	if end < 0 {
		end += y
	} else {
		end += x
	}
	if start < 0 {
		start += y
	} else {
		start += x
	}
	if start < x {
		start = x
	}
	if end >= y {
		end = y
	}
	for ; start <= end; start++ {
		if v, err := get(tx, eListKey(k, start)); err != nil {
			continue
		} else {
			vs = append(vs, v)
		}
	}
	return vs, nil
}

func sclear(tx *pebble.DB, k []byte) error {
	k = eSetKey(k, []byte{})
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		if err := tx.Delete(itr.Key(), nil); err != nil {
			return err
		}
	}
	return nil
}

func sdel(tx *pebble.DB, k, v []byte) error {
	return del(tx, eSetKey(k, v))
}

func sadd(tx *pebble.DB, k, v []byte) error {
	return set(tx, eSetKey(k, v), []byte{})
}

func selem(tx *pebble.DB, k, v []byte) (bool, error) {
	_, err := get(tx, eSetKey(k, v))
	switch {
	case err == nil:
		return true, nil
	case err == store.NotExist:
		return false, nil
	default:
		return false, err
	}
}

func smembers(tx *pebble.DB, k []byte) ([][]byte, error) {
	var vs [][]byte

	k = eSetKey(k, []byte{})
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		vs = append(vs, append([]byte{}, itr.Value()...))
	}
	return vs, nil
}

func zclear(tx *pebble.DB, k []byte) error {
	key := []byte{}
	key = append([]byte("sz"), miscellaneous.E32func(uint32(len(k)))...)
	key = append(key, k...)
	key = append(key, byte('+'))
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		score, v := dZetScore(itr.Key())
		if err := del(tx, eZetKey(k, v)); err != nil {
			return err
		}
		if err := del(tx, eZetScore(k, v, score)); err != nil {
			return err
		}
	}
	return nil
}

func zdel(tx *pebble.DB, k, v []byte) error {
	key := eZetKey(k, v)
	buf, err := get(tx, key)
	if err != nil {
		return err
	}
	if err := del(tx, key); err != nil {
		return err
	}
	score, _ := miscellaneous.D32func(buf)
	if err := del(tx, eZetScore(k, v, int32(score))); err != nil {
		return err
	}
	return nil
}

func zscore(tx *pebble.DB, k, v []byte) (int32, error) {
	if buf, err := get(tx, eZetKey(k, v)); err != nil {
		return -1, err
	} else {
		score, _ := miscellaneous.D32func(buf)
		return int32(score), nil
	}
}

func zadd(tx *pebble.DB, k []byte, score int32, v []byte) error {
	if err := set(tx, eZetKey(k, v), miscellaneous.E32func(uint32(score))); err != nil {
		return err
	}
	if err := set(tx, eZetScore(k, v, score), []byte{}); err != nil {
		return err
	}
	return nil
}

func zrange(tx *pebble.DB, k []byte, start, end int32) ([][]byte, error) {
	var vs [][]byte

	key := []byte{}
	key = append([]byte("sz"), miscellaneous.E32func(uint32(len(k)))...)
	key = append(key, k...)
	key = append(key, byte('+'))
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
	itr := tx.NewIter(prefixIterOptions(k))
	defer itr.Close()
	for itr.First(); itr.Valid(); itr.Next() {
		if score, v := dZetScore(itr.Key()); score > end {
			break
		} else {
			vs = append(vs, miscellaneous.Dup(v))
		}
	}
	return vs, nil
}

func listStartEnd(tx *pebble.DB, k []byte) (int64, int64, error) {
	if v, err := get(tx, eListMetaKey(k)); err != nil {
		return 0, 0, err
	} else {
		start, end := dListMetaValue(v)
		return start, end, nil
	}
}

// 'l' + k
func eListMetaKey(k []byte) []byte {
	return append([]byte{'l'}, k...)
}

func dListMetaKey(buf []byte) []byte {
	return buf[1:]
}

// start + end
func eListMetaValue(start, end int64) []byte {
	return append(miscellaneous.E64func(uint64(start)), miscellaneous.E64func(uint64(end))...)
}

func dListMetaValue(buf []byte) (int64, int64) {
	start, _ := miscellaneous.D64func(buf[:8])
	end, _ := miscellaneous.D64func(buf[8:16])
	return int64(start), int64(end)
}

// 'l' + k + index
func eListKey(k []byte, idx int64) []byte {
	buf := []byte{}
	buf = append([]byte{'l'}, k...)
	buf = append(buf, miscellaneous.E64func(uint64(idx))...)
	return buf
}

func dListKey(buf []byte) []byte {
	n := len(buf)
	return buf[1 : n-8]
}

// 'm' + mlen + m + '+' + k
func eMapKey(m, k []byte) []byte {
	buf := []byte{}
	buf = append([]byte{'m'}, miscellaneous.E32func(uint32(len(m)))...)
	buf = append(buf, m...)
	buf = append(buf, byte('+'))
	buf = append(buf, k...)
	return buf
}

func dMapKey(buf []byte) []byte {
	buf = buf[1:]
	n, _ := miscellaneous.D32func(buf[:4])
	return buf[5+n:]
}

// 's' + klen + k + '+' + v
func eSetKey(k, v []byte) []byte {
	buf := []byte{}
	buf = append([]byte{'s'}, miscellaneous.E32func(uint32(len(k)))...)
	buf = append(buf, k...)
	buf = append(buf, byte('+'))
	buf = append(buf, v...)
	return buf
}

func dSetKey(buf []byte) []byte {
	buf = buf[1:]
	n, _ := miscellaneous.D32func(buf[:4])
	return buf[5+n:]
}

// 'z' + klen + k + '+' + v
func eZetKey(k, v []byte) []byte {
	buf := []byte{}
	buf = append([]byte{'z'}, miscellaneous.E32func(uint32(len(k)))...)
	buf = append(buf, k...)
	buf = append(buf, byte('+'))
	buf = append(buf, v...)
	return buf
}

func dZetKey(buf []byte) []byte {
	buf = buf[1:]
	n, _ := miscellaneous.D32func(buf[:4])
	return buf[5+n:]
}

// 'sz' + klen + k + '+' + score + v
func eZetScore(k, v []byte, score int32) []byte {
	buf := []byte{}
	buf = append([]byte("sz"), miscellaneous.E32func(uint32(len(k)))...)
	buf = append(buf, k...)
	buf = append(buf, byte('+'))
	buf = append(buf, miscellaneous.EB32func(uint32(score))...)
	buf = append(buf, v...)
	return buf
}

func dZetScore(buf []byte) (int32, []byte) {
	buf = buf[2:]
	n, _ := miscellaneous.D32func(buf[:4])
	score, _ := miscellaneous.DB32func(buf[5+n : 9+n])
	return int32(score), buf[9+n:]
}
