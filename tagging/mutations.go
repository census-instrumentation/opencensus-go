package tagging

// Mutation is the interface that all mutations types need to implements. A
// mutation is a data structure holding a key, a value and a behavior. The
// mutations value types supported are string, int64 and bool.
type Mutation interface {
	WriteValueToBuffer(dst []byte) []byte
	WriteKeyValueToBuffer(dst []byte) []byte
	Key() Key
	Behavior() TagBehavior
}

// mutationString represents the mutations of type string.
type mutationString struct {
	*keyString
	behavior TagBehavior
	v        string
}

func (ms *mutationString) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ms *mutationString) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ms *mutationString) Key() Key {
	return ms
}

func (ms *mutationString) Behavior() TagBehavior {
	return ms.behavior
}

func (ms *mutationString) Value() string {
	return ms.v
}

// mutationInt64 represents the mutations of type int64.
type mutationInt64 struct {
	*keyInt64
	behavior TagBehavior
	v        int64
}

func (mi *mutationInt64) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (mi *mutationInt64) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (mi *mutationInt64) Key() Key {
	return mi
}

func (mi *mutationInt64) Behavior() TagBehavior {
	return mi.behavior
}

func (mi *mutationInt64) Value() int64 {
	return mi.v
}

// mutationFloat64 represents the mutations of type float64.
type mutationFloat64 struct {
	*keyFloat64
	behavior TagBehavior
	v        float64
}

func (mf *mutationFloat64) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (mf *mutationFloat64) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (mf *mutationFloat64) Key() Key {
	return mf
}

func (mf *mutationFloat64) Behavior() TagBehavior {
	return mf.behavior
}

func (mf *mutationFloat64) Value() float64 {
	return mf.v
}

// mutationBool represents the mutations of type bool.
type mutationBool struct {
	*keyBool
	behavior TagBehavior
	v        bool
}

func (mb *mutationBool) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (mb *mutationBool) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (mb *mutationBool) Key() Key {
	return mb
}

func (mb *mutationBool) Behavior() TagBehavior {
	return mb.behavior
}

func (mb *mutationBool) Value() bool {
	return mb.v
}
