package tagging

// Tag is the tuple (key, value) interface for all tag types.
type Tag interface {
	WriteValueToBuffer(dst []byte) []byte
	WriteKeyValueToBuffer(dst []byte) []byte
	Key() Key
}

// tagString is the tuple (key, value) implementation for tags of value type
// string.
type tagString struct {
	*keyString
	v string
}

func (ts *tagString) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ts *tagString) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ts *tagString) Key() Key {
	return ts.keyString
}

func (ts *tagString) Value() string {
	return ts.v
}

// tagInt64 is the tuple (key, value) implementation for tags of value type
// int64.
type tagInt64 struct {
	*keyInt64
	v int64
}

func (ti *tagInt64) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ti *tagInt64) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ti *tagInt64) Key() Key {
	return ti.keyInt64
}

func (ti *tagInt64) Value() int64 {
	return ti.v
}

// tagFloat64 is the tuple (key, value) implementation for tags of value type
// float64.
type tagFloat64 struct {
	*keyFloat64
	v float64
}

func (tf *tagFloat64) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (tf *tagFloat64) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (tf *tagFloat64) Key() Key {
	return tf.keyFloat64
}

func (tf *tagFloat64) Value() float64 {
	return tf.v
}

// tagBool is the tuple (key, value) implementation for tags of value type
// bool.
type tagBool struct {
	*keyBool
	v bool
}

func (tb *tagBool) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (tb *tagBool) WriteKeyValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (tb *tagBool) Key() Key {
	return tb.keyBool
}

func (tb *tagBool) Value() bool {
	return tb.v
}
