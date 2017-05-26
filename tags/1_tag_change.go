package tags

// TagOp defines the types of operations allowed.
type TagOp byte

const (
	// TagOp is not a valid operation. It is here just to detect that a TagOp isn't set.
	TagOpInvalid TagOp = iota

	// TagInsert adds the (key, value) to a set if the set doesn't already
	// contain a tag with the same key. Otherwise it is a no-op.
	TagOpInsert

	// TagOpUpdate replaces the (key, value) in a set if the set contains a
	// (key, value) pair with the same key. Otherwise it is a no-op.
	TagOpUpdate

	// TagOpUpsert adds the (key, value) to a set regardless if the set does
	// contain or doesn't contain a (key, value) pair with the same key.
	TagOpUpsert

	// TagOpDelete deletes the (key, value) from a set if it contain a pair
	// with the same key. Otherwise it is a no-op.
	TagOpDelete
)

// TagChange is the interface for tag changes. It is not expected to have
// multiple types implement it. Its main purpose is to only allow read
// operations on its fields and hide its the write operations.
type TagChange interface {
	Key() Key
	Value() []byte
	Op() TagOp
}

// tagChange implements TagChange
type tagChange struct {
	k  Key
	v  []byte
	op TagOp
}

func (tc *tagChange) Key() Key {
	return tc.k
}

func (tc *tagChange) Value() []byte {
	return tc.v
}

func (tc *tagChange) Op() TagOp {
	return tc.op
}
