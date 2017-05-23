package tags

// Tag is the tuple (key, value) used only when extracting []Tag from a TagSet.
type Tag struct {
	Key          Key
	ValueAsBytes []byte
}
