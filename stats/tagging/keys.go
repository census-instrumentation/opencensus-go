package tagging

// Key is the interface for all key types.
type Key interface {
	Name() string
}

// keyString is implementation for keys which values are of type string.
type keyString struct {
	name string
}

func (ks *keyString) Name() string {
	return ks.name
}

func (ks *keyString) CreateMutation(v string, mb MutationBehavior) *mutationString {
	return &mutationString{
		tagString: &tagString{
			keyString: ks,
			v:         v,
		},
		behavior: mb,
	}
}

func (ks *keyString) createTag(s string) *tagString {
	return &tagString{
		keyString: ks,
		v:         s,
	}
}

func (ks *keyString) RetrieveMutation(sig []byte) *mutationString {
	// TODO(acetechnologist): implement
	return nil
}

// keyInt64 is implementation for keys which values are of type int64.
type keyInt64 struct {
	name string
}

func (ki *keyInt64) Name() string {
	return ki.name
}

func (ki *keyInt64) CreateMutation(v int64, mb MutationBehavior) *mutationInt64 {
	return &mutationInt64{
		tagInt64: &tagInt64{
			keyInt64: ki,
			v:        v,
		},
		behavior: mb,
	}
}

func (ki *keyInt64) createTag(i int64) *tagInt64 {
	return &tagInt64{
		keyInt64: ki,
		v:        i,
	}
}

func (ki *keyInt64) RetrieveMutation(sig []byte) *mutationInt64 {
	// TODO(acetechnologist): implement
	return nil
}

// keyFloat64 is implementation for keys which values are of type float64.
type keyFloat64 struct {
	name string
}

func (kf *keyFloat64) Name() string {
	return kf.name
}

func (kf *keyFloat64) CreateMutation(v float64, mb MutationBehavior) *mutationFloat64 {
	return &mutationFloat64{
		tagFloat64: &tagFloat64{
			keyFloat64: kf,
			v:          v,
		},
		behavior: mb,
	}
}

func (kf *keyFloat64) createTag(f float64) *tagFloat64 {
	return &tagFloat64{
		keyFloat64: kf,
		v:          f,
	}
}

func (kf *keyFloat64) RetrieveMutation(sig []byte) *mutationFloat64 {
	// TODO(acetechnologist): implement
	return nil
}

// keyBool is implementation for keys which values are of type bool.
type keyBool struct {
	name string
}

func (kb *keyBool) Name() string {
	return kb.name
}

func (kb *keyBool) CreateMutation(v bool, mb MutationBehavior) *mutationBool {
	return &mutationBool{
		tagBool: &tagBool{
			keyBool: kb,
			v:       v,
		},
		behavior: mb,
	}
}

func (kb *keyBool) createTag(b bool) *tagBool {
	return &tagBool{
		keyBool: kb,
		v:       b,
	}
}

func (kb *keyBool) RetrieveMutation(sig []byte) *mutationBool {
	// TODO(acetechnologist): implement
	return nil
}
