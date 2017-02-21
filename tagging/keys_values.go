package tagging

type KeyValue interface {
	WriteValueToBuffer(dst []byte) []byte
	WriteKeyValueToBuffer(dst []byte) []byte

	AddKeyValueToBuffer(dst []byte) (added bool)          // Add if not exist already otherwise no-op
	ReplaceKeyValueToBuffer(dst []byte) (replaced bool)   // Replace if exist otherwise no-op
	AddOrReplaceKeyValueToBuffer(dst []byte) (added bool) // Add if not exist already otherwise replace
}

type keyValueString struct {
	*keyString
	v string
}

func (ks *keyValueString) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ks *keyValueString) Key() string {
	return ks.name
}

func (ks *keyValueString) Value() string {
	return ks.v
}

type keyValueInt64 struct {
	*keyInt64
	v int64
}

func (ki *keyValueInt64) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (ki *keyValueInt64) Key() string {
	return ki.name
}

func (ki *keyValueInt64) Value() int64 {
	return ki.v
}

type keyValueFloat64 struct {
	*keyFloat64
	v float64
}

func (kf *keyValueFloat64) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (kf *keyValueFloat64) Key() string {
	return kf.name
}

func (kf *keyValueFloat64) Value() float64 {
	return kf.v
}

type keyValueBool struct {
	*keyBool
	v bool
}

func (kb *keyValueBool) WriteValueToBuffer(dst []byte) []byte {
	// TODO(acetechnologist): implement
	return nil
}

func (kb *keyValueBool) Key() string {
	return kb.name
}

func (kb *keyValueBool) Value() bool {
	return kb.v
}
