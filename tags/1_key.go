package tags

import "encoding/binary"

var keys []Key

// Key is the interface for all key types.
type Key interface {
	Name() string
	ID() uint16
}

// KeyString implements the Key interface and is used to represent keys for
// which the value type is a string.
type KeyString struct {
	name string
	id   uint16
}

func (k *KeyString) CreateChange(s string, op TagOp) TagChange {
	return &tagChange{
		k:  k,
		v:  []byte(s),
		op: op,
	}
}

func (k *KeyString) Name() string{
	return k.name
}

func (k *KeyString) ID() uint16 {
	return k.id
}

// KeyBool implements the Key interface and is used to represent keys for which
// the value type is a string.
type KeyBool struct {
	name string
	id   uint16
}

func (k *KeyBool) CreateChange(b bool, op TagOp) TagChange {
	tc := &tagChange{
		k:  k,
		op: op,
	}
	if b {
		tc.v = []byte{1}
		return tc
	}
	tc.v = []byte{0}
	return tc
}

func (k *KeyBool) Name() string {
	return k.name
}

func (k *KeyBool) ID() uint16{
	return k.id
}

// KeyInt64 implements the Key interface and is used to represent keys for
// which the value type is a int64.
type KeyInt64 struct {
	name string
	id   uint16
}

func (k *KeyInt64) CreateChange(i int64, op TagOp) TagChange {
	tc := &tagChange{
		k:  k,
		op: op,
	}
	tc.v = make([]byte, 8)
	binary.LittleEndian.PutUint64(tc.v, uint64(i))
	return tc
}

func (k *KeyInt64) Name() string {
	return k.name
}

func (k *KeyInt64) ID() uint16 {
	return k.id
}

func getKeyByID(id uint16) Key {
	if int(id) >= len(keys) {
		return nil
	}
	return keys[id]
}