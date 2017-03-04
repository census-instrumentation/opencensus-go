package adapter

import "github.com/google/instrumentation-go/stats/tagging"

func encodeToGrpcFormat(ts tagging.TagsSet) ([]byte, error) {
	return nil, nil
}

func decodeFromGrpcFormat(bytes []byte) (tagging.TagsSet, error) {
	return nil, nil
}
