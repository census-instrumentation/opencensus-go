package tagging

// Mutation is the interface that all mutations types need to implements. A
// mutation is a data structure holding a key, a value and a behavior. The
// mutations value types supported are string, int64 and bool.
type Mutation interface {
	WriteValueToBuffer(dst []byte) []byte
	WriteKeyValueToBuffer(dst []byte) []byte
	Tag() Tag
	Behavior() MutationBehavior
}

// mutationString represents a mutation for a tag of type string.
type mutationString struct {
	*tagString
	behavior MutationBehavior
}

func (ms *mutationString) Tag() Tag {
	return ms.tagString
}

func (ms *mutationString) Behavior() MutationBehavior {
	return ms.behavior
}

// mutationInt64 represents a mutation for a tag of type int64.
type mutationInt64 struct {
	*tagInt64
	behavior MutationBehavior
}

func (mi *mutationInt64) Tag() Tag {
	return mi.tagInt64
}

func (mi *mutationInt64) Behavior() MutationBehavior {
	return mi.behavior
}

// mutationFloat64 represents a mutation for a tag of type float64.
type mutationFloat64 struct {
	*tagFloat64
	behavior MutationBehavior
}

func (mf *mutationFloat64) Tag() Tag {
	return mf.tagFloat64
}

func (mf *mutationFloat64) Behavior() MutationBehavior {
	return mf.behavior
}

// mutationBool represents a mutation for a tag of type bool.
type mutationBool struct {
	*tagBool
	behavior MutationBehavior
}

func (mb *mutationBool) Tag() Tag {
	return mb.tagBool
}

func (mb *mutationBool) Behavior() MutationBehavior {
	return mb.behavior
}

// MutationBehavior defines the types of mutations allowed.
type MutationBehavior byte

const (
	// BehaviorUnknown is not a valid behavior. It it is here just to detect
	// that a MutationBehavior isn't set.
	BehaviorUnknown MutationBehavior = iota

	// BehaviorReplace replaces the (key, value) in a set if the set already
	// contains a (key, value) pair with the same key. Otherwise it is a no-op.
	BehaviorReplace

	// BehaviorAdd adds the (key, value) in a set if the set doesn't contains a
	// (key, value) pair with the same key. Otherwise it is a no-op.
	BehaviorAdd

	// BehaviorAddOrReplace replaces the (key, value) in a set if the set
	// contains a (key, value) pair with the same key. Otherwise it adds the
	// (key, value) to the set.
	BehaviorAddOrReplace
)
