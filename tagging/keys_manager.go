package tagging

type keysManager struct{}

func DefaultKeyManager() *keysManager {
	return &keysManager{}
}

func (km *keysManager) CreateKeyString(name string) *keyString {
	return &keyString{
		name: name,
	}
}

func (km *keysManager) CreateKeyInt64(name string) *keyInt64 {
	return &keyInt64{
		name: name,
	}
}

func (km *keysManager) CreateKeyFloat64(name string) *keyFloat64 {
	return &keyFloat64{
		name: name,
	}
}

func (km *keysManager) CreateKeyBool(name string) *keyBool {
	return &keyBool{
		name: name,
	}
}
