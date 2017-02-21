package tagging

type Key interface {
	Name() string
}

type keyString struct {
	name string
}

func (ks *keyString) Name() string {
	return ks.name
}

func (ks *keyString) CreateKeyValue(v string) *keyValueString {
	return &keyValueString{
		keyString: ks,
		v:         v,
	}
}

func (ks *keyString) RetrieveKeyValue(sig []byte) *keyValueString {
	// TODO(acetechnologist): implement
	return nil
}

type keyInt64 struct {
	name string
}

func (ki *keyInt64) Name() string {
	return ki.name
}

func (ki *keyInt64) CreateKeyValue(v int64) *keyValueInt64 {
	return &keyValueInt64{
		keyInt64: ki,
		v:        v,
	}
}

func (ki *keyInt64) RetrieveKeyValue(sig []byte) *keyValueInt64 {
	// TODO(acetechnologist): implement
	return nil
}

type keyFloat64 struct {
	name string
}

func (kf *keyFloat64) Name() string {
	return kf.name
}

func (kf *keyFloat64) CreateKeyValue(v float64) *keyValueFloat64 {
	return &keyValueFloat64{
		keyFloat64: kf,
		v:          v,
	}
}

func (kf *keyFloat64) RetrieveKeyValue(sig []byte) *keyValueFloat64 {
	// TODO(acetechnologist): implement
	return nil
}

type keyBool struct {
	name string
}

func (kb *keyBool) Name() string {
	return kb.name
}

func (kb *keyBool) CreateKeyValue(v bool) *keyValueBool {
	return &keyValueBool{
		keyBool: kb,
		v:       v,
	}
}

func (kb *keyBool) RetrieveKeyValue(sig []byte) *keyValueBool {
	// TODO(acetechnologist): implement
	return nil
}
