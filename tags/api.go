package tags

var CreateTagKeyString func(name string) (KeyString, error)

var CreateTagKeyInt64 func(name string) (KeyInt64, error)

var CreateTagKeyBool func(name string) (KeyBool, error)

// key is the interface for all key types.
type Key interface {
	Name() string
	ID() int32
}

type KeyString interface {
	Key
	CreateChange(v string, op TagOp) TagChange
}

type KeyInt64 interface {
	Key
	CreateChange(v int64, op TagOp) TagChange
}

type KeyBool interface {
	Key
	CreateChange(v bool, op TagOp) TagChange
}

// Tag is the tuple (key, value) interface for all tag types.
type Tag interface {
	Key() Key
	ValueAsBytes() []byte
	EncodeValueToBuffer(dst *buffer)
	EncodeKeyToBuffer(dst *buffer)
}

type TagString interface {
	Tag
	Value() string
}

type TagInt64 interface {
	Tag
	Value() int64
}

type TagBool interface {
	Tag
	Value() bool
}

// TagChange is the type representing a TagChange as applied to a set of tags.
type TagChange struct {
	Tag
	TagOp
}

func (tc *TagChange) 

// TagOp defines the types of operations allowed when creating a tag change.
type TagOp byte

const (
	// TagOp is not a valid operation. It is here just to detect that a TagOp isn't set.
	TagOpInvalid TagOp = iota

	// TagInsert adds the (key, value) to a set if the set doesn't already
	// contain a tag with the same key. Otherwise it is a no-op.
	TagOpInsert

	// TagOpSet adds the (key, value) to a set regardless if the set doesn't
	// contains a (key, value) pair with the same key. Otherwise it is a no-op.
	TagOpSet

	// TagOpReplace replaces the (key, value) in a set if the set contains a
	// (key, value) pair with the same key. Otherwise it is a no-op.
	TagOpReplace
)
