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

func (ks *keyString) CreateMutation(v string, mb TagBehavior) *mutationString {
	return &mutationString{
		keyString: ks,
		v:         v,
		behavior:  mb,
	}
}

func (ks *keyString) RetrieveMutation(sig []byte) *mutationString {
	// TODO(acetechnologist): implement
	return nil
}

type keyInt64 struct {
	name string
}

func (ki *keyInt64) Name() string {
	return ki.name
}

func (ki *keyInt64) CreateMutation(v int64, mb TagBehavior) *mutationInt64 {
	return &mutationInt64{
		keyInt64: ki,
		v:        v,
		behavior: mb,
	}
}

func (ki *keyInt64) RetrieveMutation(sig []byte) *mutationInt64 {
	// TODO(acetechnologist): implement
	return nil
}

type keyFloat64 struct {
	name string
}

func (kf *keyFloat64) Name() string {
	return kf.name
}

func (kf *keyFloat64) CreateMutation(v float64, mb TagBehavior) *mutationFloat64 {
	return &mutationFloat64{
		keyFloat64: kf,
		v:          v,
		behavior:   mb,
	}
}

func (kf *keyFloat64) RetrieveMutation(sig []byte) *mutationFloat64 {
	// TODO(acetechnologist): implement
	return nil
}

type keyBool struct {
	name string
}

func (kb *keyBool) Name() string {
	return kb.name
}

func (kb *keyBool) CreateMutation(v bool, mb TagBehavior) *mutationBool {
	return &mutationBool{
		keyBool:  kb,
		v:        v,
		behavior: mb,
	}
}

func (kb *keyBool) RetrieveMutation(sig []byte) *mutationBool {
	// TODO(acetechnologist): implement
	return nil
}
