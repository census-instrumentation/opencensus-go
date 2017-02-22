package tagging

// TagBehavior defines the types of mutations allowed.
type TagBehavior byte

const (
	BehaviorUnknown TagBehavior = iota

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
